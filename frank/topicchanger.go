package frank

import (
	frankconf "github.com/breunigs/frank/config"
	irc "github.com/fluffle/goirc/client"
	"log"
	"regexp"
	"strings"
	"time"
)

const INTERVAL_PERIOD time.Duration = 24 * time.Hour
const HOUR_TO_TICK int = 0
const MINUTE_TO_TICK int = 0
const SECOND_TO_TICK int = 1

var regexTomorrow = regexp.MustCompile(`(?i)\smorgen:?\s`)
var regexToday = regexp.MustCompile(`(?i)\sheute:?\s`)

func TopicChanger(conn *irc.Conn) {
	ticker := updateTicker()
	for {
		<-ticker.C
		for _, cn := range strings.Split(frankconf.TopicChanger, " ") {
			setTopic(conn, cn)
		}
		ticker = updateTicker()
	}
}

// simple cron design by Daniele B. Thank you.
// http://stackoverflow.com/a/19549474/1684530
func updateTicker() *time.Ticker {
	tn := time.Now()
	nextTick := time.Date(tn.Year(), tn.Month(), tn.Day(), HOUR_TO_TICK, MINUTE_TO_TICK, SECOND_TO_TICK, 0, time.Local)
	if !nextTick.After(time.Now()) {
		nextTick = nextTick.Add(INTERVAL_PERIOD)
	}
	diff := nextTick.Sub(time.Now())
	log.Printf("next topic check in: %s", diff)
	return time.NewTicker(diff)
}

func setTopic(conn *irc.Conn, channel string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("most likely coding error: %v", r)
		}
	}()

	topic := conn.StateTracker().GetChannel(channel).Topic
	newtopic := updateTopicText(topic)

	if topic == newtopic {
		return
	}

	log.Printf("%s OLD TOPIC: %s", channel, topic)
	log.Printf("%s NEW TOPIC: %s", channel, newtopic)

	conn.Topic(channel, newtopic)
}

func updateTopicText(topic string) string {
	sep := "|"

	parts := strings.Split(" "+topic+" ", sep)
	new := []string{}

	dateToday := time.Now().Format("2006-01-02")
	dateTomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	for _, part := range parts {
		if strings.Contains(part, dateToday) {
			part = strings.Replace(part, dateToday, "HEUTE", -1)
			new = append(new, part)

		} else if strings.Contains(part, dateTomorrow) {
			part = strings.Replace(part, dateTomorrow, "MORGEN", -1)
			new = append(new, part)

		} else if regexTomorrow.MatchString(part) {
			// tomorrow → today
			match := regexTomorrow.FindStringSubmatch(part)[0]
			r := " heute"
			if strings.HasSuffix(match, ": ") {
				r += ":"
			}
			r += " "

			if strings.HasPrefix(match, " MOR") {
				r = strings.ToUpper(r)
			}

			n := regexTomorrow.ReplaceAllString(part, r)
			new = append(new, n)

		} else if regexToday.MatchString(part) {
			// today → (remove)

		} else {
			// keep
			new = append(new, part)
		}
	}
	return strings.TrimSpace(strings.Join(new, sep))
}
