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
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// how many URLs can the cache store
const cacheSize = 50

// how many hours an entry should be considered valid
const cacheValidHours = 12

// how many kilo bytes should be considered when looking for the title
// tag.
const httpReadKByte = 100

// abort HTTP requests if it takes longer than X seconds. Not sure, it’s
// definitely magic involved. Must be larger than 5.
const httpGetDeadline = 10

// new line replace regex
var newlineReplacer = regexp.MustCompile(`\s+`)

var ignoreDomainsRegex = regexp.MustCompile(`^http://p\.nnev\.de`)

func UriFind(conn *irc.Conn, line *irc.Line) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	msg := line.Args[1]

	urls := extract(msg)

	for _, url := range urls {
		if url == "" {
			continue
		}

		if title := findCache(url); title != "" {
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
			title, err := titleGet(url)
			if err != nil {
				postTitle(conn, line, err.Error(), "Error")
			} else if title != "" {
				postTitle(conn, line, title, "")
				addCache(url, title)
			}
		}(url)
	}
}

// regexing ////////////////////////////////////////////////////////////

func extract(msg string) []string {
	find := exec.Command("./urifind", "-u")
	pipe, err := find.StdinPipe()
	if err != nil {
		log.Printf("WTF: couldn’t open stdin pipe to urifind: %s", err)
		return nil
	}
	pipe.Write([]byte(msg))
	pipe.Close()
	out, err := find.Output()
	if err != nil {
		log.Printf("WTF: urlfind failed with: %s", err)
		return nil
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}

// http/html stuff /////////////////////////////////////////////////////

func titleGet(url string) (string, error) {
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
		return "", err
	}
	defer r.Body.Close()

	// TODO: r.Body → utf8?
	title := titleParseHtml(io.LimitReader(r.Body, 1024*httpReadKByte))
	title = newlineReplacer.ReplaceAllString(title, " ")
	log.Printf("Title for URL %s: %s\n", url, title)

	if r.StatusCode != 200 {
		return "", errors.New("[" + strconv.Itoa(r.StatusCode) + "] " + title)
	}

	return title, nil
}

func titleParseHtml(r io.Reader) string {
	doc, err := html.Parse(r)
	if err != nil {
		log.Printf("WTF: html parser blew up: %s\r\n", err)
		return ""
	}

	title := ""

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Title {
			title = ""
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type != html.TextNode {
					continue
				}
				title += c.Data
			}

		} else { // recurse down
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	f(doc)

	return strings.TrimSpace(title)
}

// Cache ///////////////////////////////////////////////////////////////
type Cache struct {
	url   string
	title string
	date  time.Time
}

var cache = [cacheSize]Cache{}
var cacheIndex = 0

func addCache(url string, title string) {
	if len(cache) == cacheIndex {
		cacheIndex = 0
	}
	cache[cacheIndex] = Cache{url, title, time.Now()}
	cacheIndex += 1
}

func findCache(url string) string {
	for _, cc := range cache {
		if cc.url == url && time.Since(cc.date).Hours() <= cacheValidHours {
			return cc.title
		}
	}
	return ""
}

// util ////////////////////////////////////////////////////////////////

func postTitle(conn *irc.Conn, line *irc.Line, title string, prefix string) {
	tgt := line.Args[0]

	log.Printf("nick=%s, target=%s, title=%s", line.Nick, tgt, title)
	// if target is our current nick, it was a private message.
	// Answer the users in this case.
	if tgt == conn.Me().Nick {
		tgt = line.Nick
	}
	if prefix == "" {
		prefix = "Link Info"
	} else {
		prefix = newlineReplacer.ReplaceAllString(prefix, " ")
	}
	title = newlineReplacer.ReplaceAllString(title, " ")
	// use notice instead of PrivMsg to avoid bots answering each other
	conn.Notice(tgt, "["+prefix+"] "+title)
}
