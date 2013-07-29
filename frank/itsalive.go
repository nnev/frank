package frank

import (
	frankconf "github.com/breunigs/frank/config"
	irc "github.com/fluffle/goirc/client"
	"log"
	"strings"
)

func ItsAlive(conn *irc.Conn, line *irc.Line) {
	if line.Args[0] != conn.Me().Nick {
		// no private query, ignore
		return
	}

	msg := line.Args[1]
	if !strings.HasPrefix(msg, "msg #") {
		//~ log.Printf("unknown prefix for: %s", msg)
		// only accept valid commands
		return
	}

	if line.Nick != frankconf.Master {
		// only answer to master
		log.Printf("only answering to master %s, but was %s", frankconf.Master, line.Nick)
		return
	}

	cmd := strings.SplitN(msg, " ", 3)
	channel := cmd[1]
	msg = cmd[2]

	log.Printf("master command: posting “%s” to %s", msg, channel)
	conn.Privmsg(channel, msg)

}
