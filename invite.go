package main

import (
	"log"
	"strings"
)

func runnerInvite(parsed Message) {
	if parsed.Command != "INVITE" {
		return
	}

	n := Nick(parsed)
	if !IsNickAdmin(parsed) && strings.ToLower(n) != "chanserv" {
		log.Printf("not reacting on invite from non-admin user: %s", n)
		return
	}

	if Target(parsed) != *nick {
		log.Printf("invite: Weird, target is not me? my nick=%s  target=%s", *nick, Target(parsed))
		return
	}

	channel := parsed.Trailing
	log.Printf("Following invite for channel: %s", channel)
	Join(channel)
}
