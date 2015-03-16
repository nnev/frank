package main

import (
	"log"
	"strings"
)

func listenerInvite(parsed Message) bool {
	if parsed.Command != "INVITE" {
		return true
	}

	n := Nick(parsed)
	if !IsNickAdmin(parsed) && strings.ToLower(n) != "chanserv" {
		log.Printf("not reacting on invite from non-admin user: %s", n)
		return true
	}

	if Target(parsed) != *nick {
		log.Printf("invite: Weird, target is not me? my nick=%s  target=%s", *nick, Target(parsed))
		return true
	}

	channel := parsed.Trailing
	log.Printf("Following invite for channel: %s", channel)
	Join(channel)

	return true
}
