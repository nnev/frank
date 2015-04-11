package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// how often to check the feeds (in minutes)
const checkEvery = 3

// ignore all posts that are older than X minutes
const freshness = 90

// if there’s an error reading a feed, retry after X minutes
const retryAfter = 9

// how many items to show if there have been many updates in an interval
const maxItems = 3

var bootTimestamp = time.Now()

var rssHttpClient = http.Client{Timeout: 10 * time.Second}

func Rss() {
	// this feels wrong, the missing alignment making it hard to read.
	// Does anybody have a suggestion how to make this nice in go?
	// go pollFeed("#i3", "i3faq", timeFormat1, "https://faq.i3wm.org/feeds/rss/")

	go pollFeed("#chaos-hd", "nn-web", "https://www.noname-ev.de/gitcommits.atom")
	go pollFeed("#chaos-hd", "nn-wiki", "https://www.noname-ev.de/wiki/index.php?title=Special:RecentChanges&feed=atom")
	go pollFeed("#chaos-hd", "nn-planet", "http://blogs.noname-ev.de/atom.xml")
	go pollFeed("#chaos-hd", "frank", "https://github.com/breunigs/frank/commits/robust.atom")
}

type Feed struct {
	// XMLName Name      `xml:"http://www.w3.org/2005/Atom feed"`
	TitleRaw string    `xml:"title"`
	Id       string    `xml:"id"`
	Link     string    `xml:"link"`
	Updated  time.Time `xml:"updated,attr"`
	Author   string    `xml:"author"`
	Entry    []Entry   `xml:"entry"`
}

func (f Feed) postableForIrc() []string {
	oneLiners := []string{}

	for _, entry := range f.Entry {
		if !entry.RecentlyPublished() {
			if *verbose {
				// log.Printf("RSS: skipping non-recent entry. published @ %s :: %s %s", entry.Updated, f.Title(), entry.Title())
			}
			continue
		}

		if isRecentUrl(entry.Href()) {
			if *verbose {
				log.Printf("RSS: skipping already already posted :: %s %s", f.Title(), entry.Title())
			}
			continue
		}
		addRecentUrl(entry.Href())

		oneLiners = appendIfMiss(oneLiners, entry.OneLiner())
	}

	return oneLiners
}

func (f Feed) Title() string {
	return strings.TrimSpace(f.TitleRaw)
}

type Entry struct {
	TitleRaw string    `xml:"title"`
	Id       string    `xml:"id"`
	Link     []Link    `xml:"link"`
	Updated  time.Time `xml:"updated"`
	Author   string    `xml:"author>name"`
}

func (e Entry) Title() string {
	return strings.TrimSpace(e.TitleRaw)
}

func (e Entry) RecentlyPublished() bool {
	if bootTimestamp.After(e.Updated) {
		return false
	}

	return time.Since(e.Updated) < freshness*time.Minute
}

func (e Entry) Href() string {
	if len(e.Link) == 0 {
		return ""
	}

	return strings.TrimSpace(e.Link[0].Href)
}

func (e Entry) OneLiner() string {
	author := strings.TrimSpace(e.Author)
	if author != "" {
		author = " (by " + author + ")"
	}

	return e.Title() + author + " " + e.Href()
}

type Link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
}

func loadURL(url string) []byte {
	r, err := rssHttpClient.Get(url)

	if err != nil {
		log.Printf("RSS: could resolve URL %s: %s\n", url, err)
		return []byte{}
	}
	defer r.Body.Close()

	// read up to 1 MB
	limitedBody := io.LimitReader(r.Body, 1024*1024)
	body, err := ioutil.ReadAll(limitedBody)
	if err != nil {
		log.Printf("RSS: could read data from URL %s: %s\n", url, err)
		return []byte{}
	}

	return body
}

func parseAtomFeed(url string) Feed {
	f := Feed{}
	if err := xml.Unmarshal(loadURL(url), &f); err != nil {
		log.Printf("RSS: could not parse %s: %s\n", url, err)
	}

	return f
}

func pollFeed(channel string, feedName string, url string) {
	for {
		time.Sleep(checkEvery * time.Minute)
		if *verbose {
			log.Printf("RSS %s: checking", feedName)
		}
		pollFeedRunner(channel, feedName, url)
	}
}

func pollFeedRunner(channel string, feedName string, url string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg:RSS: %v\n", r)
			time.Sleep(retryAfter * time.Minute)
			return
		}
	}()

	postitems := parseAtomFeed(url).postableForIrc()
	cnt := len(postitems)

	// hide updates if they exceed the maxItems counter. If there’s only
	// one more item in the list than specified in maxItems, all of the
	// items will be printed – otherwise that item would be replaced by
	// a useless message that it has been hidden.
	if cnt > maxItems+1 {
		msg := fmt.Sprintf("::%s:: had %d updates, showing the latest %d", feedName, cnt, maxItems)
		Privmsg(channel, msg)
		postitems = postitems[cnt-maxItems : cnt]
	}

	// newer items appear first in feeds, so reverse them here to keep
	// the order in line with how IRC wprks
	for i := len(postitems) - 1; i >= 0; i -= 1 {
		Privmsg(channel, "::"+feedName+":: "+postitems[i])
		log.Printf("RSS %s: posting %s\n", feedName, postitems[i])
	}
}

// append string to slice only if it’s not already present.
func appendIfMiss(slice []string, s string) []string {
	for _, elm := range slice {
		if elm == s {
			return slice
		}
	}
	return append(slice, s)
}

// LIFO that stores the recent posted URLs. Used to avoid posting entries multiple times.
var recent []string = make([]string, 50)
var recentIndex = 0

func addRecentUrl(url string) {
	recent[recentIndex] = url
	recentIndex += 1
	if len(recent) == recentIndex {
		recentIndex = 0
	}
}

func isRecentUrl(url string) bool {
	for _, a := range recent {
		if url == a {
			return true
		}
	}
	return false
}
