package main

import (
	"log"
	"regexp"
	"strings"

	"gopkg.in/sorcix/irc.v2"
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

func IsPrivateQuery(p *irc.Message) bool {
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

func Nick(p *irc.Message) string {
	return p.Prefix.Name
}

func Hostmask(p *irc.Message) string {
	return p.Prefix.Host
}

func Target(parsed *irc.Message) string {
	p := parsed.Params
	if len(p) == 0 {
		return ""
	} else {
		return p[0]
	}
}

func IsNickAdmin(p *irc.Message) bool {
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
