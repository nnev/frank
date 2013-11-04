package frank

import (
	irc "github.com/fluffle/goirc/client"
	"log"
	"time"
)

var lastHelps = map[string]time.Time{}

func Help(conn *irc.Conn, line *irc.Line) {
	if line.Args[0] != conn.Me().Nick {
		// no private query, ignore
		return
	}

	if line.Args[1] != "help" && line.Args[1] != "!help" {
		// no help request, ignore
		return
	}

	last := lastHelps[line.Nick]
	if time.Since(last).Minutes() <= 1 {
		log.Printf("User %s tried spamming for help, not answering (last request @ %v)", line.Nick, last)
		return
	}

	lastHelps[line.Nick] = time.Now()

	conn.Privmsg(line.Nick, "1. Test your IRC client’s highlighting:")
	conn.Privmsg(line.Nick, "  – /msg frank high")
	conn.Privmsg(line.Nick, "  – /msg frank high custom_text")
	conn.Privmsg(line.Nick, "  – /msg frank highpub custom_text")
	conn.Privmsg(line.Nick, "“high” sends you a private message, “highpub” posts to #test.")
	conn.Privmsg(line.Nick, "Your nick will be used unless custom_text is defined. Delay is always 5 seconds.")
	conn.Privmsg(line.Nick, " ")

	conn.Privmsg(line.Nick, "2. I won’t spoiler URLs if you add “no spoiler” to your message")
	conn.Privmsg(line.Nick, " ")

	conn.Privmsg(line.Nick, "3. There’s a karma system. You can’t vote on yourself.")
	conn.Privmsg(line.Nick, "  – thing++ # optional comment")
	conn.Privmsg(line.Nick, "  – thing-- # thing may be alphanumerical, Unicode is supported")
	conn.Privmsg(line.Nick, "  – karma for thing  //  karma thing  //  karma thing?")
	conn.Privmsg(line.Nick, " ")

	conn.Privmsg(line.Nick, "4. I’ll answer to !raum in certain channels.")
	conn.Privmsg(line.Nick, " ")

	conn.Privmsg(line.Nick, "If you need more details, please look at my source:")
	conn.Privmsg(line.Nick, "https://github.com/breunigs/frank")

}
