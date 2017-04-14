package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"gopkg.in/sorcix/irc.v2"
)

var highlightRe = regexp.MustCompile(`^(?:highpub|high)(?:\s+(.{1,70}))*`)

func runnerHighlight(msg *irc.Message) error {
	// Only accept highlight request via queries (as opposed to in
	// channels).
	if msg.Command != irc.PRIVMSG ||
		len(msg.Params) < 1 ||
		strings.HasPrefix(msg.Params[0], "#") {
		return nil
	}

	matches := highlightRe.FindStringSubmatch(msg.Trailing())
	if matches == nil {
		return nil // no highlight request
	}

	nick := msg.Prefix.Name // for convenience
	log.Printf("received highlighting request from %s: %#v", nick, matches)
	highlight := nick
	if matches[1] != "" {
		highlight = matches[1]
	}

	Privmsg(nick, "will highlight you in 5 seconds")

	// allow for 100ms round trip time to highlight on time
	time.Sleep(4900 * time.Millisecond)

	if strings.HasPrefix(msg.Trailing(), "highpub") {
		log.Printf("highlighting %s publicly for: %s", nick, highlight)
		Privmsg("#test", "highlight test: "+highlight)
	} else {
		log.Printf("highlighting %s privately for: %s", nick, highlight)
		Privmsg(nick, highlight)
	}

	return nil
}
