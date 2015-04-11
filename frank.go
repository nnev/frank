package main

import (
	"flag"
	"fmt"
	parser "github.com/husio/irc"
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

type Message *parser.Message

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
		log.Printf("Exiting due to signal %q", sig)
		kill()
	}()
}

func setupSessionErrorHandler() {
	go func() {
		err := <-session.Errors
		log.Fatal("RobustIRC session error: %v", err)
	}()
}

func setupJoinChannels() {
	for _, channel := range strings.Split(*channels, " ") {
		Join(channel)
	}
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

	if *nickserv_password == "" {
		setupJoinChannels()
		return
	}

	nickserv := make(chan bool, 1)
	listener := ListenerAdd("nickserv auth detector", func(parsed Message) {
		// PREFIX=services.robustirc.net COMMAND=MODE PARAMS=[frank2] TRAILING=+r
		is_me := Target(parsed) == *nick
		is_plus_r := strings.HasPrefix(parsed.Trailing, "+") && strings.Contains(parsed.Trailing, "r")

		if parsed.Command == "MODE" && is_me && is_plus_r {
			nickserv <- true
		}
	})

	log.Printf("NICKSERV: Authenticating…")
	Privmsg("nickserv", "identify "+*nickserv_password)

	go func() {
		select {
		case <-nickserv:
			log.Printf("NICKSERV: auth successful")

		case <-time.After(10 * time.Second):
			log.Printf("NICKSERV: auth failed. No response within 10s, joining channels anyway. Maybe check the password, i.e. “/msg frank msg nickserv identify <pass>” and watch the logs.")
		}

		listener.Remove()
		setupJoinChannels()
	}()
}

func parse(msg string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("parser broken: %v\nMessage that caused this: %s", r, msg)
		}
	}()

	if strings.TrimSpace(msg) == "" {
		return
	}

	parsed, err := parser.ParseLine(msg)
	if err != nil {
		log.Fatal("Could not parse IRC message: %v", err)
		return
	}

	if parsed.Command == "PONG" {
		return
	}

	listenersRun(parsed)
}

func main() {
	listenersReset()
	setupFlags()
	setupSession()
	setupSignalHandler()
	setupKeepalive()
	setupSessionErrorHandler()
	boot()

	go TopicChanger()
	go Rss()

	ListenerAdd("help", runnerHelp)
	ListenerAdd("admin", runnerAdmin)
	ListenerAdd("highlight", runnerHighlight)
	ListenerAdd("karma", runnerKarma)
	ListenerAdd("invite", runnerInvite)
	ListenerAdd("lmgtfy", runnerLmgtfy)
	ListenerAdd("urifind", runnerUrifind)
	ListenerAdd("raumbang", runnerRaumbang)

	if *verbose {
		ListenerAdd("verbose debugger", func(parsed Message) {
			log.Printf("< PREFIX=%s COMMAND=%s PARAMS=%s TRAILING=%s", parsed.Prefix, parsed.Command, parsed.Params, parsed.Trailing)
		})
	}

	ListenerAdd("nickname checker", func(parsed Message) {
		if parsed.Command == ERR_NICKNAMEINUSE {
			log.Printf("Nickname is already in use. Sleeping for a minute before restarting.")
			listenersReset()
			time.Sleep(time.Minute)
			log.Printf("Killing now due to nickname being in use")
			kill()
		}
	})

	for {
		msg := <-session.Messages
		parse(msg)
	}
}
