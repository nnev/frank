package main

import (
	"flag"
	"fmt"
	parser "github.com/husio/go-irc"
	"github.com/robustirc/bridge/robustsession"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	network   = flag.String("network", "", `DNS name to connect to (e.g. "robustirc.net"). The _robustirc._tcp SRV record must be present.`)
	tlsCAFile = flag.String("tls_ca_file", "", "Use the specified file as trusted CA instead of the system CAs. Useful for testing.")

	channels          = flag.String("channels", "", "channels the bot should join. Space separated.")
	nick              = flag.String("nick", "frank", "nickname of the bot")
	admins            = flag.String("admins", "xeen", "users who can control the bot. Space separated.")
	nickserv_password = flag.String("nickserv_password", "", "password used to identify with nickserv. No action is taken if password is blank or not set.")

	verbose = flag.Bool("verbose", false, "enable to get very detailed logs")
)

type Message parser.Message

var session *robustsession.RobustSession

func setupFlags() {
	flag.Parse()

	if *network == "" {
		log.Fatal("You must specify -network")
	}
}

func setupSession() {
	var err error
	session, err = robustsession.Create(*network, *tlsCAFile)
	if err != nil {
		log.Fatal("Could not create RobustIRC session: %v", err)
	}

	log.Printf("Created RobustSession for %s. Session id: %s", *nick, session.SessionId())
}

func setupKeepalive() {
	// TODO: only if no other traffic
	go func() {
		keepaliveToNetwork := time.After(1 * time.Minute)
		for {
			<-keepaliveToNetwork
			session.PostMessage("PING keepalive")
			keepaliveToNetwork = time.After(1 * time.Minute)
		}
	}()
}

func setupSignalHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-signalChan
		log.Printf("Exiting due to signal %q\n", sig)
		kill()
	}()
}

func setupSessionErrorHandler() {
	go func() {
		err := <-session.Errors
		log.Fatal("RobustIRC session error: %v", err)
	}()
}

func kill() {
	log.Printf("Deleting Session. Goodbye.")

	if err := session.Delete(*nick + " says goodbye"); err != nil {
		log.Fatalf("Could not properly delete RobustIRC session: %v", err)
	}

	os.Exit(int(syscall.SIGTERM) | 0x80)
}

func boot() {
	Post(fmt.Sprintf("NICK %s", *nick))
	Post(fmt.Sprintf("USER bot 0 * :%s von Bötterich", *nick))

	nickserv := make(chan bool)
	if *nickserv_password != "" {
		ListenerAdd(func(parsed Message) bool {
			from_nickserv := strings.ToLower(Nick(parsed)) == "nickserv"

			if parsed.Command() == "NOTICE" && from_nickserv {
				nickserv <- true
				return false
			}

			return true
		})

		log.Printf("Authenticating with NickServ")
		Privmsg("nickserv", "identify "+*nickserv_password)
	} else {
		nickserv <- false
	}

	go func() {
		select {
		case <-nickserv:
		case <-time.After(10 * time.Second):
			log.Printf("No response from nickserv, joining channels anyway")
		}

		for _, channel := range strings.Split(*channels, " ") {
			Join(channel)
		}
	}()
}

func parse(msg string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("parser broken: %v\nMessage that caused this: %s", r, msg)
		}
	}()

	parsed, err := parser.ParseLine(msg)
	if err != nil {
		log.Fatal("Could not parse IRC message: %v", err)
		return
	}

	if parsed.Command() == "PONG" {
		return
	}

	listenersRun(parsed)
}

func main() {
	listeners = []Listener{}
	setupFlags()
	setupSession()
	setupSignalHandler()
	setupKeepalive()
	setupSessionErrorHandler()
	boot()

	go TopicChanger()
	go Rss()

	ListenerAdd(listenerHelp)
	ListenerAdd(listenerAdmin)
	ListenerAdd(listenerHighlight)
	ListenerAdd(listenerKarma)
	ListenerAdd(listenerInvite)
	ListenerAdd(listenerLmgtfy)
	ListenerAdd(listenerUrifind)
	ListenerAdd(listenerRaumbang)

	if *verbose {
		ListenerAdd(func(parsed Message) bool {
			log.Printf("< PREFIX=%s COMMAND=%s PARAMS=%s TRAILING=%s", parsed.Prefix(), parsed.Command(), parsed.Params(), parsed.Trailing())
			return true
		})
	}

	ListenerAdd(func(parsed Message) bool {
		if parsed.Command() == ERR_NICKNAMEINUSE {
			log.Printf("Nickname is already in use. Sleeping for a minute before restarting.")
			listenersReset()
			time.Sleep(time.Minute)
			log.Printf("Killing now due to nickname being in use")
			kill()
			return false
		}
		return true
	})

	for {
		msg := <-session.Messages
		parse(msg)
	}
}
