package frank

import (
	"code.google.com/p/go.net/html"
	"code.google.com/p/go.net/html/atom"
	"errors"
	irc "github.com/fluffle/goirc/client"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// how many URLs can the cache store
const cacheSize = 100

// how many hours an entry should be considered valid
const cacheValidHours = 24

// how many kilo bytes should be considered when looking for the title
// tag.
const httpReadKByte = 100

// abort HTTP requests if it takes longer than X seconds. Not sure, it’s
// definitely magic involved. Must be larger than 5.
const httpGetDeadline = 10

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

func UriFind(conn *irc.Conn, line *irc.Line) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	msg := line.Args[1]

	if noSpoilerRegex.MatchString(msg) {
		log.Printf("not spoilering this line: %s", msg)
		return
	}

	urls := extract(msg)

	for _, url := range urls {
		if url == "" {
			continue
		}

		if title := cacheGetTitleByUrl(url); title != "" {
			log.Printf("using cache for URL: %s", url)
			postTitle(conn, line, title, "Cache Info")
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
			} else if title != "" {
				postTitle(conn, line, title, "")
				cacheAdd(url, title)
			}
		}(url)
	}
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

		// End on closing paren, but only if there is an opening
		// paren before the URL (should fix most false-positives).
		if end := strings.Index(url, ")"); idx > 0 && msg[idx-1] == '(' && end > -1 {
			url = url[:end]
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
	// via http://www.reddit.com/r/golang/comments/10awvj/timeout_on_httpget/c6bz49s
	c := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(time.Second * httpGetDeadline)
				c, err := net.DialTimeout(netw, addr, time.Second*(httpGetDeadline-5))
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

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
	tweetUser := ""
	tweetPicUrl := ""

	var f func(*html.Node)
	f = func(n *html.Node) {
		if title == "" && n.Type == html.ElementNode && n.DataAtom == atom.Title {
			title = extractText(n)
			return
		}

		if tweetText == "" && hasClass(n, "tweet-text") {
			tweetText = extractText(n)
			return
		}

		if tweetUser == "" && hasClass(n, "js-user-profile-link") {
			tweetUser = extractText(n)
			return
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
	if tweetText != "" {
		tweetText = twitterPicsRegex.ReplaceAllString(tweetText, "")
		tweetUser = strings.Replace(tweetUser, "@", "(@", 1) + "): "
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
	if strings.Contains(" "+getAttr(n, "class")+" ", class) {
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

func cacheGetTitleByUrl(url string) string {
	for _, cc := range cache {
		if cc.url == url && time.Since(cc.date).Hours() <= cacheValidHours {
			return cc.title
		}
	}
	return ""
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

func postTitle(conn *irc.Conn, line *irc.Line, title string, prefix string) {
	tgt := line.Args[0]

	secondsAgo := cacheGetSecondsToLastPost(title)
	if secondsAgo <= noRepostWithinSeconds {
		log.Printf("Skipping, because posted %d seconds ago (“%s”)", secondsAgo, title)
		return
	}

	log.Printf("nick=%s, target=%s, title=%s", line.Nick, tgt, title)
	// if target is our current nick, it was a private message.
	// Answer the users in this case.
	if tgt == conn.Me().Nick {
		tgt = line.Nick
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
	conn.Privmsg(tgt, "["+prefix+"] "+title)
}

func clean(text string) string {
	text = whitespaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
