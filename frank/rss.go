package frank

import (
	irc "github.com/fluffle/goirc/client"
	rss "github.com/jteeuwen/go-pkg-rss"
	"log"
	"strconv"
	"time"
)

// how often to check the feeds (in minutes)
const checkEvery = 3

// ignore all posts that are older than X minutes
const freshness = 10

// if there’s an error reading a feed, retry after X minutes
const retryAfter = 9

// how many items to show if there have been many updates in an interval
const maxItems = 2

// reference time: Mon Jan 2 15:04:05 -0700 MST 2006
const timeFormat1 = "Mon, 02 Jan 2006 15:04:05 -0700"
const timeFormat2 = "2006-01-02T15:04:05Z"

var conn *irc.Conn

var ignoreBefore = time.Now()

func Rss(connection *irc.Conn) {
	conn = connection
	// this feels wrong, the missing alignment making it hard to read.
	// Does anybody have a suggestion how to make this nice in go?
	//~ go pollFeed("#i3-test", "i3", timeFormat2, "http://code.stapelberg.de/git/i3/atom/?h=next")
	go pollFeed("#i3-test", "i3lock", timeFormat2, "http://code.stapelberg.de/git/i3lock/atom/?h=master")
	go pollFeed("#i3-test", "i3status", timeFormat2, "http://code.stapelberg.de/git/i3status/atom/?h=master")
	go pollFeed("#i3-test", "i3website", timeFormat2, "http://code.stapelberg.de/git/i3-website/atom/?h=master")
	go pollFeed("#i3-test", "i3-faq", timeFormat1, "https://faq.i3wm.org/feeds/rss/")

	go pollFeed("#chaos-hd", "nn-wiki", timeFormat2, "https://www.noname-ev.de/wiki/index.php?title=Special:RecentChanges&feed=atom")
	go pollFeed("#chaos-hd", "nn-planet", timeFormat2, "http://blogs.noname-ev.de/atom.xml")
}

func pollFeed(channel string, feedName string, timeFormat string, uri string) {
	// this will process all incoming new feed items and discard all that
	// are somehow erroneous or older than the threshold. It will directly
	// post any updates.
	itemHandler := func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
		log.Printf("RSS: %d new item(s) in %s\n", len(newitems), feedName)

		postitems := []string{}

		for _, item := range newitems {
			pubdate, err := time.Parse(timeFormat, item.PubDate)
			// ignore items with unreadable date format
			if err != nil {
				log.Printf("RSS: WTF @ reading date for %s: %s (err: %v)", feedName, item.PubDate, err)
				continue
			}

			// ignore items that were posted before frank booted or are older
			// than “freshness” minutes
			if ignoreBefore.After(pubdate) || time.Since(pubdate) >= freshness*time.Minute {
				log.Printf("RSS: skipping old post for %s (posted at %s)", feedName, pubdate)
				continue
			}

			url := ""
			if len(item.Links) > 0 {
				url = item.Links[0].Href
			}
			author := item.Author.Name

			if author == "" {
				postitems = append(postitems, "::"+feedName+":: "+item.Title+" @ "+url)
			} else {
				postitems = append(postitems, "::"+feedName+":: "+item.Title+" @ "+url+" (by "+author+")")
			}
		}

		cnt := len(postitems)

		// hide updates if they exceed the maxItems counter. If there’s only
		// one more item in the list than specified in maxItems, all of the
		// items will be printed – otherwise that item would be replaced by
		// a useless message that it has been hidden.
		if cnt > maxItems+1 {
			cntS := strconv.Itoa(cnt)
			maxS := strconv.Itoa(maxItems)
			msg := "::" + feedName + ":: had " + cntS + " updates, showing the latest " + maxS
			conn.Privmsg(channel, msg)
			postitems = postitems[cnt-maxItems : cnt]
		}

		// newer items appear first in feeds, so reverse them here to keep
		// the order in line with how IRC wprks
		for i := len(postitems) - 1; i >= 0; i -= 1 {
			conn.Privmsg(channel, postitems[i])
			log.Printf("RSS-post: %s", postitems[i])
		}
	}

	// create the feed listener/updater
	feed := rss.New(checkEvery, true, chanHandler, itemHandler)

	// check for updates infinite loop
	for {
		log.Printf("RSS: updating %s", feedName)
		if err := feed.Fetch(uri, nil); err != nil {
			log.Printf("RSS: [e] %s: %s", uri, err)
			time.Sleep(retryAfter * time.Minute)
			continue
		}

		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
	}
}

// unused default handler
func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	log.Printf("RSS: %d new channel(s) in %s\n", len(newchannels), feed.Url)
}
