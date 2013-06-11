package main

import (
	"flag"
	"strings"
	"github.com/breunigs/frank/frank"
	irc "github.com/fluffle/goirc/client"
	"log"
)

//~ const instaJoin = "#chaos-hd"
const instaJoin = "#test"

const nickServPass = ""

const ircServer = "irc.twice-irc.de"

func main() {
	flag.Parse() // parses the logging flags. TODO

	c := irc.SimpleClient("frank", "frank", "Frank Böterrich der Zweite")
	c.SSL = true

	// connect
	c.AddHandler(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			log.Printf("Connected as: %s\n", conn.Me.Nick)
			conn.Privmsg("nickserv", "identify " + nickServPass)
			for _, cn := range(strings.Split(instaJoin, " ")) {
				if cn != "" {
					conn.Join(cn)
				}
			}
		})

	// react
	c.AddHandler("PRIVMSG",
		func(conn *irc.Conn, line *irc.Line) {
			//~ tgt := line.Args[0]
			//~ msg := line.Args[1]

			go func() { frank.RaumBang(conn, line) }()
			go func() { frank.UriFind(conn, line) }()

			//~ log.Printf("      Debug: tgt: %s, msg: %s\n", tgt, msg)
		})

	// auto follow invites
	c.AddHandler("INVITE",
		func(conn *irc.Conn, line *irc.Line) {
			tgt := line.Args[0]
			cnnl := line.Args[1]
			if conn.Me.Nick != tgt {
				log.Printf("WTF: received invite for %s but target was %s\n")
				return
			}

			log.Printf("Following invite for channel: %s\n", cnnl)
			conn.Join(cnnl)
		})

	// auto deop frank
	c.AddHandler("MODE",
		func(conn *irc.Conn, line *irc.Line) {
			if len(line.Args) != 3 {
				// mode statement cannot be not in a channel, so ignore
				return
			}

			if line.Args[2] != conn.Me.Nick {
				// not referring to us
				return
			}

			if line.Args[1] != "+o" {
				// not relevant
				return
			}

			cn := line.Args[0]
			conn.Mode(cn, "+v", conn.Me.Nick)
			conn.Mode(cn, "-o", conn.Me.Nick)
			conn.Privmsg(cn, line.Nick + ": SKYNET® Protection activated")
		})

	// disconnect
	quit := make(chan bool)
	c.AddHandler(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// go go GO!
	if err := c.Connect(ircServer); err != nil {
		log.Printf("Connection error: %s\n", err)
	}

	log.Printf("Frank has booted\n")

	// Wait for disconnect
	<-quit
}
