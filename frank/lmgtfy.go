package frank

import (
	irc "github.com/fluffle/goirc/client"
	"log"
	"net/url"
	"regexp"
)

const googUrl = "http://googl.com/search?btnI=1&q="

// regex that matches lmgtfy requests
var lmgtfyMatcher = regexp.MustCompile(`^(?:[\d\pL._-]+: )?lmgtfy:? (.+)`)

func Lmgtfy(conn *irc.Conn, line *irc.Line) {
	tgt := line.Args[0]
	msg := line.Args[1]

	if tgt[0:1] != "#" {
		// only answer to this in channels
		return
	}

	post := extractPost(msg)

	if post == "" {
		return
	} else {
		conn.Privmsg(tgt, post)
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
