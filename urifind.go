package main

import (
	"bufio"
	"bytes"
	_ "crypto/sha512"
	"compress/zlib"
	"errors"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/transform"
	"gopkg.in/sorcix/irc.v2"
)

// how many URLs can the cache store
const cacheSize = 500

// how many hours an entry should be considered valid
const cacheValidHours = 24

// how many bytes should be considered when looking for the title tag.
const httpReadByte = 1024 * 100
const httpReadBytePDF = 1024 * 1024 * 3 // 3 MB

// don’t repost the same title within this period
const noRepostWithinSeconds = 30

const titleMaxAllowedLength = 500

// matches all whitespace and zero bytes. Additionally, all Unicode
// characters of class Cf (format chars, e.g. right-to-left) and Cc
// (control chars) are matched.
var whitespaceRegex = regexp.MustCompile(`[\s\0\p{Cf}\p{Cc}]+`)

var ignoreDomainsRegex = regexp.MustCompile(`^http://p\.nnev\.de`)

var noSpoilerRegex = regexp.MustCompile(`(?i)(don't|no|kein|nicht) *spoiler`)

// extract data from a PDF's document information dictionary
var pdfAuthorRegex = regexp.MustCompile(`/Author\(([^)]+?)\)`)
var pdfTitleRegex = regexp.MustCompile(`/Title\(([^)]+?)\)`)
var pdfSubjectRegex = regexp.MustCompile(`/Subject\(([^)]+?)\)`)

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

func runnerUrifind(parsed *irc.Message) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	if parsed.Command != "PRIVMSG" {
		return nil
	}

	msg := parsed.Trailing()

	if noSpoilerRegex.MatchString(msg) {
		log.Printf("not spoilering this line: %s", msg)
		return nil
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
			title := ""
			if strings.HasSuffix(strings.ToLower(url), ".pdf") {
				title = PDFTitleGet(url)
			} else {
				c := http.Client{Timeout: 10 * time.Second}
				title, _, _ = TitleGet(&c, url)
			}
			if !IsIn(title, pointlessTitles) {
				postTitle(parsed, title, "")
				cacheAdd(url, title)
			}
		}(url)
	}
	return nil
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

// PDF stuff ///////////////////////////////////////////////////////////

func PDFTitleGet(url string) string {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Coding error in PDFTitleGet: %v", r)
		}
	}()

	c := http.Client{Timeout: 10 * time.Second}
	gTitle, _, gErr := TitleGet(&c, "https://webcache.googleusercontent.com/search?q=cache:"+url)
	if gErr == nil && len(gTitle) > 0 {
		return gTitle
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("WTF: could not make http request %s: %s", url, err)
		return ""
	}
	req.Header.Set("User-Agent", "frank IRC Bot")

	r, err := c.Do(req)
	if err != nil {
		log.Printf("WTF: could not resolve %s: %s", url, err)
		return ""
	}
	defer r.Body.Close()

	reader := bufio.NewReader(io.LimitReader(r.Body, httpReadBytePDF))

	author := ""
	title := ""

	inDictionary := false
	cnt := 0
	for {
		cnt++
		line, err := reader.ReadString('\n')

		if err == io.EOF {
			break
		}

		// PDF 32000-1:2008 -- 7.3.7 Dictionary Objects
		if strings.HasPrefix(line, "<<") {
			inDictionary = true
		}

		if inDictionary {
			if m := pdfAuthorRegex.FindStringSubmatch(line); len(m) == 2 {
				author = clean(m[1])
			}

			if m := pdfTitleRegex.FindStringSubmatch(line); len(m) == 2 {
				title = clean(m[1])
			}

			if m := pdfSubjectRegex.FindStringSubmatch(line); len(m) == 2 && title == "" {
				title = clean(m[1])
			}
		}

		if strings.HasPrefix(line, ">>") || strings.HasSuffix(line, ">>") {
			inDictionary = false
		}
	}

	if title == "" {
		return ""
	}

	if author == "" {
		return title
	} else {
		return title + " by " + author
	}
}

// http/html stuff /////////////////////////////////////////////////////

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

func TitleGet(doer Doer, url string) (string, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("WTF: could not make http request %s: %s", url, err)
		return "", url, err
	}
	req.Header.Set("User-Agent", "frank IRC Bot")

	r, err := doer.Do(req)
	if err != nil {
		log.Printf("WTF: could not resolve %s: %s", url, err)
		return "", url, err
	}
	// from https://github.com/gaul/anaconda/commit/ba67efed60e7dcf8a96fd9053531fc1d517fd7df
	// Twitter returns deflate data despite the client only requesting gzip
	// data.  net/http automatically handles the latter but not the former:
	// https://github.com/golang/go/issues/18779
	if r.Header.Get("Content-Encoding") == "deflate" {
		var err error
		r.Body, err = zlib.NewReader(r.Body)
		if err != nil {
			return "", url, err
		}
	}
	defer r.Body.Close()

	lastUrl := r.Request.URL.String()

	head := make([]byte, 1024)

	bytesRead, err := io.ReadFull(r.Body, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		log.Printf("Could not read from %s: %s", url, err)
		return "", url, err
	}

	limited := io.LimitedReader{r.Body, int64(httpReadByte - bytesRead)}
	reader := io.MultiReader(bytes.NewReader(head[:bytesRead]), &limited)

	contentType := r.Header.Get("Content-Type")
	encoding, _, _ := charset.DetermineEncoding(head, contentType)
	reader = transform.NewReader(reader, encoding.NewDecoder())

	title, err := extractTitleFromHTML(reader)
	if err != nil {
		return "", lastUrl, err
	}

	if len(title) > titleMaxAllowedLength {
		title = title[:titleMaxAllowedLength]
	}

	if r.StatusCode != 200 {
		return "", lastUrl, errors.New("[" + strconv.Itoa(r.StatusCode) + "] " + title)
	}

	log.Printf("Title for URL %s: %s", url, title)

	return title, lastUrl, nil
}

func extractTitleFromHTML(body io.Reader) (string, error) {
	node, err := html.Parse(body)
	if err != nil {
		return "", err
	}
	var title string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
		}
		if title != "" {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)
	return title, nil
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

func postTitle(parsed *irc.Message, title string, prefix string) {
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
