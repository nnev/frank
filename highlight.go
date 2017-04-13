package main

import (
	"log"
	"regexp"
	"strings"
	"time"

	"gopkg.in/sorcix/irc.v2"
)

var customTextRegex = regexp.MustCompile(`^(?:high|highpub)\s+(.{1,70})`)

func runnerHighlight(parsed *irc.Message) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	if !IsPrivateQuery(parsed) {
		return nil
	}

	msg := parsed.Trailing()

	if !strings.HasPrefix(msg, "high") {
		// no highlight request, ignore
		return nil
	}

	n := Nick(parsed)

	log.Printf("received highlighting request from %s", n)

	highlight := n
	if customTextRegex.MatchString(msg) {
		match := customTextRegex.FindStringSubmatch(msg)
		highlight = match[1]
	}

	// allow for 100ms round trip time to highlight on time
	time.Sleep(4900 * time.Millisecond)

	if strings.HasPrefix(msg, "highpub") {
		log.Printf("highlighting %s publicly for: %s", n, highlight)
		Privmsg("#test", "highlight test: "+highlight)
	} else {
		log.Printf("highlighting %s privately for: %s", n, highlight)
		Privmsg(n, highlight)
	}

	return nil
}
