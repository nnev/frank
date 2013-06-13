//Encoding a map to a gob. Save the gob to disk. Read the gob from disk. Decode the gob into another map.
package frank

import (
	irc "github.com/fluffle/goirc/client"
	"log"
	"time"
)

var lastHelps = map[string]time.Time{}

func Help(conn *irc.Conn, line *irc.Line) {
	if line.Args[0] != conn.Me.Nick {
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

	conn.Privmsg(line.Nick, "It’s a game to find out what "+conn.Me.Nick+" can do.")
	conn.Privmsg(line.Nick, "1. Most likely I can find out the <title> of an URL, if:")
	conn.Privmsg(line.Nick, "  – I am in the channel where it is posted")
	conn.Privmsg(line.Nick, "  – you sent it in a query to me")
	conn.Privmsg(line.Nick, "  I’m going to cache that URL for a certain amount of time.")
	conn.Privmsg(line.Nick, "2. There’s a karma system. You can’t vote on yourself.")
	conn.Privmsg(line.Nick, "  – thing++ # optional comment")
	conn.Privmsg(line.Nick, "  – thing-- # thing may be alphanumerical, Unicode is supported")
	conn.Privmsg(line.Nick, "  – karma for thing  //  karma thing  //  karma thing?")
	conn.Privmsg(line.Nick, "3. I’ll answer to !raum in certain channels.")
	conn.Privmsg(line.Nick, "If you need more details, please look at my source:")
	conn.Privmsg(line.Nick, "https://github.com/breunigs/frank")

}
