package main

import (
	"log"
	"strings"

	"gopkg.in/sorcix/irc.v2"
)

func runnerInvite(parsed *irc.Message) error {
	if parsed.Command != "INVITE" {
		return nil
	}

	n := Nick(parsed)
	if !IsNickAdmin(parsed) && strings.ToLower(n) != "chanserv" {
		log.Printf("not reacting on invite from non-admin user: %s", n)
		return nil
	}

	if Target(parsed) != *nick {
		log.Printf("invite: Weird, target is not me? my nick=%s  target=%s", *nick, Target(parsed))
		return nil
	}

	channel := parsed.Trailing()
	log.Printf("Following invite for channel: %s", channel)
	Join(channel)

	return nil
}
