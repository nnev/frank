package main

import (
	"log"
	"regexp"
	"strings"
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

func IsPrivateQuery(p Message) bool {
	return p.Command() == "PRIVMSG" && Target(p) == *nick
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
	return strings.SplitN(p.Prefix(), "!", 2)[0]
}

func Target(p Message) string {
	return p.Params()[0]
}

func IsTargetAdmin(p Message) bool {
	nick := Nick(p)
	admins := regexp.MustCompile("\\s+").Split(*admins, -1)
	for _, admin := range admins {
		if nick == admin {
			return true
		}
	}
	return false
}
