package main

import (
	parser "github.com/husio/go-irc"
	"log"
	"strings"
	"time"
)

var lastHelps = map[string]time.Time{}

func listenerHelp(parsed parser.Message) bool {
	n := Nick(parsed)

	if parsed.Command() != "PRIVMSG" || Target(parsed) != *nick {
		// no private query, ignore
		return true
	}

	content := strings.ToLower(parsed.Trailing())

	if content != "help" && content != "!help" {
		// no help request, ignore
		return true
	}

	last := lastHelps[n]
	if time.Since(last).Minutes() <= 1 {
		log.Printf("User %s tried spamming for help, not answering (last request @ %v)", n, last)
		return true
	}

	lastHelps[n] = time.Now()

	Privmsg(n, "1. Test your IRC client’s highlighting:")
	Privmsg(n, "  – /msg "+*nick+" high")
	Privmsg(n, "  – /msg "+*nick+" high custom_text")
	Privmsg(n, "  – /msg "+*nick+" highpub custom_text")
	Privmsg(n, "“high” sends you a private message, “highpub” posts to #test.")
	Privmsg(n, "Your nick will be used unless custom_text is defined. Delay is always 5 seconds.")
	Privmsg(n, " ")

	Privmsg(n, "2. I won’t spoiler URLs if you add “no spoiler” to your message")
	Privmsg(n, " ")

	Privmsg(n, "3. There’s a karma system. You can’t vote on yourself.")
	Privmsg(n, "  – thing++ # optional comment")
	Privmsg(n, "  – thing-- # thing may be alphanumerical, Unicode is supported")
	Privmsg(n, "  – karma for thing  //  karma thing  //  karma thing?")
	Privmsg(n, " ")

	Privmsg(n, "4. I’ll answer to !raum in certain channels.")
	Privmsg(n, " ")

	Privmsg(n, "If you need more details, please look at my source:")
	Privmsg(n, "https://github.com/breunigs/frank")

	return true
}
