package main

import (
	"log"
)

func listenerInvite(parsed Message) bool {
	if parsed.Command() != "INVITE" {
		return true
	}

	if !IsNickAdmin(parsed) {
		return true
	}

	if Target(parsed) != *nick {
		log.Printf("invite: Weird, target is not me? my nick=%s  target=%s", *nick, Target(parsed))
		return true
	}

	channel := parsed.Trailing()
	log.Printf("Following invite for channel: %s", channel)
	Join(channel)

	return true
}
