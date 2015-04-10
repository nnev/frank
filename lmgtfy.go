package main

import (
	"log"
	"net/url"
	"regexp"
	"strings"
)

const googUrl = "http://googl.com/search?btnI=1&q="

// regex that matches lmgtfy requests
var lmgtfyMatcher = regexp.MustCompile(`^(?:[\d\pL._-]+: )?lmgtfy:? (.+)`)

func runnerLmgtfy(parsed Message) {
	tgt := Target(parsed)
	msg := parsed.Trailing

	if !strings.HasPrefix(tgt, "#") {
		// only answer to this in channels
		return
	}

	post := extractPost(msg)

	if post != "" {
		Privmsg(tgt, post)
	}
}

// returns the String to be posted
func extractPost(msg string) string {
	if !lmgtfyMatcher.MatchString(msg) {
		return ""
	}

	match := lmgtfyMatcher.FindStringSubmatch(msg)

	if len(match) < 2 {
		log.Printf("WTF: lmgtfy regex match didn’t have enough parts")
		return ""
	}

	u := googUrl + url.QueryEscape(match[1])
	t, lastUrl, err := TitleGet(u)

	post := ""

	if err != nil {
		post = "[LMGTFY] " + lastUrl
	} else {
		post = "[LMGTFY] " + t + " @ " + lastUrl
	}

	return post
}
