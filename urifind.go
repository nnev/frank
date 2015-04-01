package main

import (
	"code.google.com/p/go.net/html"
	_ "crypto/sha512"
	"errors"
	"golang.org/x/net/html/atom"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// how many URLs can the cache store
const cacheSize = 500

// how many hours an entry should be considered valid
const cacheValidHours = 24

// how many kilo bytes should be considered when looking for the title
// tag.
const httpReadKByte = 100

// don’t repost the same title within this period
const noRepostWithinSeconds = 30

// matches all whitespace and zero bytes. Additionally, all Unicode
// characters of class Cf (format chars, e.g. right-to-left) and Cc
// (control chars) are matched.
var whitespaceRegex = regexp.MustCompile(`[\s\0\p{Cf}\p{Cc}]+`)

var ignoreDomainsRegex = regexp.MustCompile(`^http://p\.nnev\.de`)

var twitterDomainRegex = regexp.MustCompile(`(?i)^https?://(?:[a-z0-9]\.)?twitter.com`)
var twitterPicsRegex = regexp.MustCompile(`(?i)(?:\b|^)pic\.twitter\.com/[a-z0-9]+(?:\b|$)`)

var noSpoilerRegex = regexp.MustCompile(`(?i)(don't|no|kein|nicht) spoiler`)

// blacklist pointless titles /////////////////////////////////////////
var pointlessTitles = []string{"",
	"imgur: the simple image sharer",
	"Fefes Blog",
	"Gmane Loom",
	"i3 - A better tiling and dynamic window manager",
	"i3 - improved tiling wm",
	"IT-News, c't, iX, Technology Review, Telepolis | heise online",
	"debian Pastezone",
	"Index of /docs/",
	"NoName e.V. pastebin",
	"Nopaste - powered by project-mindstorm IT Services",
	"Diff NoName e.V. pastebin",
	"pr0gramm.com",
	"Google"}

func listenerUrifind(parsed Message) bool {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	if parsed.Command != "PRIVMSG" {
		return true
	}

	msg := parsed.Trailing

	if noSpoilerRegex.MatchString(msg) {
		log.Printf("not spoilering this line: %s", msg)
		return true
	}

	urls := extract(msg)

	for _, url := range urls {
		if url == "" {
			continue
		}

		if cp := cacheGetByUrl(url); cp != nil {
			log.Printf("using cache for URL: %s", cp.url)
			ago := cacheGetTimeAgo(cp)
			postTitle(parsed, cp.title, "cached "+ago+" ago")
			// Hack: add title to the cache again so we can correctly check
			// for reposts, even if the original link has been cached quite
			// some time ago. Since the repost check searches by title, but
			// here we search by URL wie get the correct time when it was
			// cached while still preventing people from using frank to
			// multiply their spamming.
			cacheAdd("", cp.title)
			continue
		}

		go func(url string) {
			if ignoreDomainsRegex.MatchString(url) {
				log.Printf("ignoring this URL: %s", url)
				return
			}

			log.Printf("testing URL: %s", url)
			title, _, err := TitleGet(url)
			if err != nil {
				//postTitle(conn, line, err.Error(), "Error")
			} else if !IsIn(title, pointlessTitles) {
				postTitle(parsed, title, "")
				cacheAdd(url, title)
			}
		}(url)
	}

	return true
}

// regexing ////////////////////////////////////////////////////////////

func extract(msg string) []string {
	results := make([]string, 0)
	for idx := strings.Index(msg, "http"); idx > -1; idx = strings.Index(msg, "http") {
		url := msg[idx:]
		if !strings.HasPrefix(url, "http://") &&
			!strings.HasPrefix(url, "https://") {
			msg = msg[idx+len("http"):]
			continue
		}

		// End on commas, but only if they are followed by a space.
		// spiegel.de URLs have commas in them, that would be a
		// false positive otherwise.
		if end := strings.Index(url, ", "); end > -1 {
			url = url[:end]
		}

		// use special handling if the URL contains closing parens
		closingParen := strings.Index(url, ")")
		if closingParen > -1 {
			absPos := idx + closingParen + 1
			if len(msg) > absPos && msg[absPos] == ')' {
				// if an URL ends with double closing parens, assume that the
				// former one belongs to the URL
				url = url[:closingParen+1]
			} else if idx > 0 && msg[idx-1] == '(' {
				// if it ends on a single closing parens (follow by other chars)
				// only remove that closing parens if the URL is directly
				// preceded by one
				url = url[:closingParen]
			}
		}

		// Whitespace always ends a URL.
		if end := strings.IndexAny(url, " \t"); end > -1 {
			url = url[:end]
		}

		results = append(results, url)
		msg = msg[idx+len(url):]
	}
	return results
}

// http/html stuff /////////////////////////////////////////////////////

func TitleGet(url string) (string, string, error) {
	c := http.Client{Timeout: 10 * time.Second}

	r, err := c.Get(url)
	if err != nil {
		log.Printf("WTF: could not resolve %s: %s\n", url, err)
		return "", url, err
	}
	defer r.Body.Close()

	lastUrl := r.Request.URL.String()

	// TODO: r.Body → utf8?
	title, tweet := titleParseHtml(io.LimitReader(r.Body, 1024*httpReadKByte))

	if r.StatusCode != 200 {
		return "", lastUrl, errors.New("[" + strconv.Itoa(r.StatusCode) + "] " + title)
	}

	if tweet != "" && twitterDomainRegex.MatchString(lastUrl) {
		title = tweet
	}

	log.Printf("Title for URL %s: %s\n", url, title)

	return title, lastUrl, nil
}

// parses the incoming HTML fragment and tries to extract text from
// suitable tags. Currently this is the page’s title tag and tweets
// when the HTML-code is similar enough to twitter.com. Returns
// title and tweet.
func titleParseHtml(r io.Reader) (string, string) {
	doc, err := html.Parse(r)
	if err != nil {
		log.Printf("WTF: html parser blew up: %s\n", err)
		return "", ""
	}

	title := ""
	tweetText := ""
	tweetUserName := ""
	tweetUserScreenName := ""
	tweetPicUrl := ""

	var f func(*html.Node)
	f = func(n *html.Node) {
		if title == "" && n.Type == html.ElementNode && n.DataAtom == atom.Title {
			title = extractText(n)
			return
		}

		if hasClass(n, "permalink-tweet") {
			tweetUserName = getAttr(n, "data-name")
			tweetUserScreenName = getAttr(n, "data-screen-name")
			// find next child “tweet-text”
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if hasClass(c, "tweet-text") {
					tweetText = extractText(c)
					break
				}
			}
		}

		isMedia := hasClass(n, "media") || hasClass(n, "media-thumbnail")
		if tweetPicUrl == "" && isMedia && !hasClass(n, "profile-picture") {
			attrVal := getAttr(n, "data-url")
			if attrVal != "" {
				tweetPicUrl = attrVal
				return
			}
		}

		// recurse down
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}

	}
	f(doc)

	// cleanup
	tweet := ""
	tweetUser := ""
	if tweetText != "" {
		tweetText = twitterPicsRegex.ReplaceAllString(tweetText, "")
		tweetUser = tweetUserName + " (@" + tweetUserScreenName + "): "
		tweet = tweetUser + tweetText + " " + tweetPicUrl
		tweet = clean(tweet)
	}

	return strings.TrimSpace(title), strings.TrimSpace(tweet)
}

func extractText(n *html.Node) string {
	text := ""
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			text += c.Data
		} else {
			text += extractText(c)
		}
	}
	return clean(text)
}

func hasClass(n *html.Node, class string) bool {
	if n.Type != html.ElementNode {
		return false
	}

	class = " " + strings.TrimSpace(class) + " "
	attr := strings.Replace(getAttr(n, "class"), "\n", " ", -1)
	if strings.Contains(" "+attr+" ", class) {
		return true
	}
	return false
}

func getAttr(n *html.Node, findAttr string) string {
	for _, attr := range n.Attr {
		if attr.Key == findAttr {
			return attr.Val
		}
	}
	return ""
}

// Cache ///////////////////////////////////////////////////////////////
type Cache struct {
	url   string
	title string
	date  time.Time
}

var cache = [cacheSize]Cache{}
var cacheIndex = 0

func cacheAdd(url string, title string) {
	if len(cache) == cacheIndex {
		cacheIndex = 0
	}
	cache[cacheIndex] = Cache{url, title, time.Now()}
	cacheIndex += 1
}

func cacheGetByUrl(url string) *Cache {
	for _, cc := range cache {
		if cc.url == url && time.Since(cc.date).Hours() <= cacheValidHours {
			return &cc
		}
	}
	return nil
}

func cacheGetTimeAgo(cc *Cache) string {
	ago := time.Since(cc.date).Minutes()
	if ago < 60 {
		return strconv.Itoa(int(ago)) + "m"
	} else {
		hours := strconv.Itoa(int(ago/60.0 + 0.5))
		return hours + "h"
	}
}

func cacheGetSecondsToLastPost(title string) int {
	var secondsAgo = int(^uint(0) >> 1)
	for _, cc := range cache {
		var a = int(time.Since(cc.date).Seconds())
		if cc.title == title && a < secondsAgo {
			secondsAgo = a
		}
	}
	return secondsAgo
}

// util ////////////////////////////////////////////////////////////////

func postTitle(parsed Message, title string, prefix string) {
	tgt := Target(parsed)

	secondsAgo := cacheGetSecondsToLastPost(title)
	if secondsAgo <= noRepostWithinSeconds {
		log.Printf("Skipping, because posted %d seconds ago (“%s”)", secondsAgo, title)
		return
	}

	if *verbose {
		log.Printf("Title was last posted: %#v (“%s”)", secondsAgo, title)
	}

	log.Printf("nick=%s, target=%s, title=%s", Nick(parsed), tgt, title)
	// if target is our current nick, it was a private message.
	// Answer the users in this case.
	if IsPrivateQuery(parsed) {
		tgt = Nick(parsed)
	}
	if prefix == "" {
		prefix = "Link Info"
	} else {
		prefix = clean(prefix)
	}
	title = clean(title)
	// the IRC spec states that notice should be used instead of msg
	// and that bots should not react to notice at all. However, no
	// real world bot adheres to this. Furthermore, people who can’t
	// configure their client to not highlight them on notices will
	// complain.
	Privmsg(tgt, "["+prefix+"] "+title)
}

func clean(text string) string {
	text = whitespaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
