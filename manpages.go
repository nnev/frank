package main

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/sorcix/irc.v2"
)

// manpagesMatcher finds words of the form "name(section)", where section is
// either a number, or a number followed by some characters. See tests for a
// list of examples.
var manpagesMatcher = regexp.MustCompile(`\b(\w+)\((\d[\da-z_-]*)\)(\W|$)`)

func runnerManpages(parsed *irc.Message) error {
	for _, l := range extractManpages(parsed.Trailing()) {
		Privmsg(Target(parsed), "[manpage] "+l)
	}
	return nil
}

func extractManpages(msg string) (links []string) {
	const prefix = "frank: man "
	if strings.HasPrefix(msg, prefix) {
		return []string{fmt.Sprintf("https://manpages.debian.org/%s", strings.Replace(msg[len(prefix):], " ", "/", -1))}
	}

	ms := manpagesMatcher.FindAllStringSubmatch(msg, -1)
	for _, m := range ms {
		links = append(links, fmt.Sprintf("https://manpages.debian.org/%s.%s", m[1], m[2]))
	}
	return links
}
