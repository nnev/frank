package main

import (
	"flag"
	"github.com/breunigs/frank/frank"
	irc "github.com/fluffle/goirc/client"
	"log"
	"strings"
)

const (
	//~ instaJoin = "#chaos-hd"
	instaJoin    = "#test"
	nickServPass = ""
	ircServer    = "irc.twice-irc.de"
	botNick      = "frank"
)

func main() {
	flag.Parse() // parses the logging flags. TODO

	cfg := irc.NewConfig(botNick, botNick, "Frank Böterrich der Zweite")
	cfg.SSL = true
	cfg.Server = ircServer
	cfg.NewNick = func(n string) string { return n + "_" }
	c := irc.Client(cfg)

	// connect
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			log.Printf("Connected as: %s\n", conn.Me().Nick)
			conn.Privmsg("nickserv", "identify "+nickServPass)
			for _, cn := range strings.Split(instaJoin, " ") {
				if cn != "" {
					conn.Join(cn)
				}
			}
		})

	// react
	c.HandleFunc("PRIVMSG",
		func(conn *irc.Conn, line *irc.Line) {
			//~ tgt := line.Args[0]
			//~ msg := line.Args[1]

			// ignore eicar, the bot we love to hate
			if line.Nick == "eicar" {
				return
			}

			go frank.RaumBang(conn, line)
			go frank.UriFind(conn, line)
			go frank.Karma(conn, line)
			go frank.Help(conn, line)

			//~ log.Printf("      Debug: tgt: %s, msg: %s\n", tgt, msg)
		})

	// auto follow invites
	c.HandleFunc("INVITE",
		func(conn *irc.Conn, line *irc.Line) {
			tgt := line.Args[0]
			cnnl := line.Args[1]
			if conn.Me().Nick != tgt {
				log.Printf("WTF: received invite for %s but target was %s\n")
				return
			}

			log.Printf("Following invite for channel: %s\n", cnnl)
			conn.Join(cnnl)
		})

	// auto deop frank
	c.HandleFunc("MODE",
		func(conn *irc.Conn, line *irc.Line) {
			if len(line.Args) != 3 {
				// mode statement cannot be not in a channel, so ignore
				return
			}

			if line.Args[2] != conn.Me().Nick {
				// not referring to us
				return
			}

			if line.Args[1] != "+o" {
				// not relevant
				return
			}

			cn := line.Args[0]
			conn.Mode(cn, "+v", conn.Me().Nick)
			conn.Mode(cn, "-o", conn.Me().Nick)
			conn.Privmsg(cn, line.Nick+": SKYNET® Protection activated")
		})

	// disconnect
	quit := make(chan bool)
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// go go GO!
	if err := c.Connect(); err != nil {
		log.Printf("Connection error: %s\n", err)
	}

	log.Printf("Frank has booted\n")

	// Wait for disconnect
	<-quit
}
