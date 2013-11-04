package frank

import (
	irc "github.com/fluffle/goirc/client"
	"log"
	"regexp"
	"strings"
	"time"
)

var customTextRegex = regexp.MustCompile(`^(?:high|highpub)\s+(.{1,70})`)

func Highlight(conn *irc.Conn, line *irc.Line) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	if line.Args[0] != conn.Me().Nick {
		// no private query, ignore
		return
	}

	msg := line.Args[1]
	if !strings.HasPrefix(msg, "high") {
		// no highlight request, ignore
		return
	}

	log.Printf("received highlighting request from %s\n", line.Nick)

	var highlight string
	if customTextRegex.MatchString(msg) {
		match := customTextRegex.FindStringSubmatch(msg)
		highlight = match[1]
	} else {
		highlight = line.Nick
	}

	// allow for 100ms round trip time to highlight on time
	time.Sleep(4900 * time.Millisecond)

	if strings.HasPrefix(msg, "highpub") {
		log.Printf("highlighting %s publicly for: %s\n", line.Nick, highlight)
		conn.Privmsg("#test", "highlight test: "+highlight)
	} else {
		log.Printf("highlighting %s privately for: %s\n", line.Nick, highlight)
		conn.Privmsg(line.Nick, highlight)
	}
}
