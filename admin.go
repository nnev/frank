package main

import (
	"log"
	"strings"
)

func listenerAdmin(parsed Message) bool {
	if !IsPrivateQuery(parsed) {
		return true
	}

	n := Nick(parsed)
	if !IsNickAdmin(parsed) {
		// log.Printf("Not executing admin command for normal user %s", n)
		return true
	}

	msg := parsed.Trailing

	if strings.HasPrefix(msg, "msg ") {
		cmd := strings.SplitN(msg, " ", 3)
		if len(cmd) >= 3 {
			channel := cmd[1]
			msg = cmd[2]

			log.Printf("ADMIN %s: posting “%s” to %s", n, msg, channel)
			Privmsg(channel, msg)
		}
	}

	if msg == "quit" || msg == "exit" {
		Privmsg(Nick(parsed), "If you really want "+*nick+" to exit, type: REALLY_QUIT")
	}

	if msg == "REALLY_QUIT" {
		Privmsg(n, "As you wish.")
		log.Printf("ADMIN %s: quitting", n)
		kill()
	}

	if strings.HasPrefix(msg, "settopic #") {
		cmd := strings.SplitN(msg, " ", 2)
		channel := cmd[1]
		setTopic(channel)
	}

	return true
}
