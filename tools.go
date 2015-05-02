package main

import (
	"errors"
	"log"
	"regexp"
	"strings"
	"time"
)

func Post(msg string) {
	log.Printf(">>> %s", msg)

	if err := session.PostMessage(msg); err != nil {
		log.Fatalf("Could not post message to RobustIRC: %v", err)
	}
}

func Privmsg(user string, msg string) {
	Post("PRIVMSG " + user + " :" + msg)
}

func Topic(channel string, topic string) {
	Post("TOPIC " + channel + " :" + topic)
}

func TopicGet(channel string) (string, error) {
	received := make(chan string)

	topicGetRunner := func(parsed Message) {
		// Example Topic:
		// PREFIX=robustirc.net COMMAND=332 PARAMS=[frank #test]

		p := parsed.Params

		if len(p) < 2 || p[1] != channel {
			// not the channel we're interested in
			return
		}

		if parsed.Command == RPL_TOPIC {
			received <- parsed.Trailing
		}

		if parsed.Command == RPL_NOTOPIC {
			received <- ""
		}

	}

	l := ListenerAdd("topic getter", topicGetRunner)
	Post("TOPIC " + channel)

	select {
	case topic := <-received:
		l.Remove()
		return topic, nil
	case <-time.After(60 * time.Second):
		l.Remove()
	}
	return "", errors.New("failed to get topic: no reply within 60 seconds")
}

func IsPrivateQuery(p Message) bool {
	return p.Command == "PRIVMSG" && Target(p) == *nick
}

func Join(channel string) {
	channel = strings.TrimSpace(channel)
	channel = strings.TrimPrefix(channel, "#")

	if channel == "" {
		return
	}

	log.Printf("joining #%s", channel)
	if *nickserv_password != "" {
		Privmsg("chanserv", "invite #"+channel)
	}
	Post("JOIN #" + channel)
}

func Nick(p Message) string {
	return strings.SplitN(p.Prefix, "!", 2)[0]
}

func Target(parsed Message) string {
	p := parsed.Params
	if len(p) == 0 {
		return ""
	} else {
		return p[0]
	}
}

func IsNickAdmin(p Message) bool {
	nick := Nick(p)
	admins := regexp.MustCompile("\\s+").Split(*admins, -1)

	for _, admin := range admins {
		if *verbose {
			log.Printf("debug admin: checking if |%s|==|%s| (=%v)", nick, admin, nick == admin)
		}
		if nick == admin {
			return true
		}
	}
	return false
}

func IsIn(needle string, haystack []string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func clean(text string) string {
	text = whitespaceRegex.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}
