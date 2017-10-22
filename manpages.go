package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/sorcix/irc.v2"
)

// manpagesMatcher finds words of the form "name(section)", where section is
// either a number, or a number followed by some characters. See tests for a
// list of examples.
var manpagesMatcher = regexp.MustCompile(`\b([\w-]+)\((\d[\da-z_-]*)\)(\W|$)`)

func runnerManpages(parsed *irc.Message) error {
	for _, l := range extractManpages(parsed.Trailing()) {
		l := l // copy
		go func() {
			req, err := http.NewRequest("HEAD", l, nil)
			if err != nil {
				log.Printf("manpage: %v", err)
				return
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("manpage: %s: %v", l, err)
				return
			}
			// for keepalive
			ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if got, want := resp.StatusCode, http.StatusOK; got != want {
				log.Printf("manpage: not replying: %s: unexpected HTTP status code: got %d, want %d", l, got, want)
				return
			}
			Privmsg(Target(parsed), "[manpage] "+l)
		}()
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
