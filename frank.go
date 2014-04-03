package main

import (
	"crypto/tls"
	"flag"
	frankconf "github.com/breunigs/frank/config"
	"github.com/breunigs/frank/frank"
	irc "github.com/fluffle/goirc/client"
	"log"
	"strings"
)

func main() {
	flag.Parse() // parses the logging flags. TODO

	cfg := irc.NewConfig(frankconf.BotNick, frankconf.BotNick, "Frank Böterrich der Zweite")
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{InsecureSkipVerify: true}
	cfg.Flood = true
	cfg.Server = frankconf.IrcServer
	cfg.NewNick = func(n string) string { return n + "_" }
	c := irc.Client(cfg)
	c.EnableStateTracking()

	// connect
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			log.Printf("Connected as: %s\n", conn.Me().Nick)
			conn.Privmsg("nickserv", "identify "+frankconf.NickServPass)

			var instaJoin string
			if frankconf.Production {
				instaJoin = frankconf.InstaJoinProduction
			} else {
				instaJoin = frankconf.InstaJoinDebug
			}

			log.Printf("AutoJoining: %s\n", instaJoin)

			for _, cn := range strings.Split(instaJoin, " ") {
				if cn != "" {
					conn.Join(cn)
				}
			}

			// handle RSS
			frank.Rss(conn)

			// watch topics and maybe change them on midnight
			go frank.TopicChanger(conn)
		})

	// react
	c.HandleFunc("PRIVMSG",
		func(conn *irc.Conn, line *irc.Line) {
			// ignore eicar, the bot we love to hate.
			// Also ignore i3-bot.
			if line.Nick == "eicar" || line.Nick == "i3" {
				return
			}

			go frank.RaumBang(conn, line)
			go frank.UriFind(conn, line)
			go frank.Lmgtfy(conn, line)
			go frank.Karma(conn, line)
			go frank.Help(conn, line)
			go frank.ItsAlive(conn, line)
			go frank.Highlight(conn, line)

			if frankconf.Debug {
				tgt := line.Args[0]
				msg := line.Args[1]
				log.Printf("Debug MSG: tgt: %s, msg: %s\n", tgt, msg)
			}
		})

	if frankconf.Debug {
		c.HandleFunc("NOTICE",
			func(conn *irc.Conn, line *irc.Line) {
				tgt := line.Args[0]
				msg := line.Args[1]
				log.Printf("Debug NOTICE: tgt: %s, msg: %s\n", tgt, msg)
			})
	}

	c.HandleFunc("INVITE",
		func(conn *irc.Conn, line *irc.Line) {
			tgt := line.Args[0]
			cnnl := line.Args[1]

			// auto follow invites only in debug mode or if asked by master
			if frankconf.Production && line.Nick != frankconf.Master {
				log.Printf("only following invites by %s in production\n", frankconf.Master)
				return
			}

			if conn.Me().Nick != tgt {
				log.Printf("WTF: received invite for %s but target was %s\n", conn.Me().Nick, tgt)
				return
			}

			log.Printf("Following invite for channel: %s\n", cnnl)
			conn.Join(cnnl)
		})

	// auto deop frank
	c.HandleFunc("MODE",
		func(conn *irc.Conn, line *irc.Line) {
			log.Printf("Mode change array length: %s", len(line.Args))
			log.Printf("Mode changes: %s", line.Args)

			if len(line.Args) < 3 {
				// mode statement cannot be not in a channel, so ignore
				return
			}

			var modeop bool // true => add mode, false => remove mode
			var nickIndex int = 2
			for i := 0; i < len(line.Args[1]); i++ {
				switch m := line.Args[1][i]; m {
				case '+':
					modeop = true
				case '-':
					modeop = false
				case 'o':
					if !modeop || line.Args[nickIndex] != conn.Me().Nick {
						nickIndex += 1
						break
					}
					channel := line.Args[0]

					if strings.Contains(" "+frankconf.OpOkIn+" ", " "+channel+" ") {
						if strings.ToLower(line.Nick) != "chanserv" {
							conn.Privmsg(channel, "Unbelievable "+line.Nick+", you… http://yrden.de/f1.ogg")
						}
					} else {
						conn.Mode(channel, "+v-o", conn.Me().Nick, conn.Me().Nick)
						conn.Privmsg(channel, line.Nick+": SKYNET® Protection activated")
					}
					return
				default:
					nickIndex += 1
				}
			}
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
