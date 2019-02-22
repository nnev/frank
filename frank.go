package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robustirc/bridge/robustsession"
	"gopkg.in/sorcix/irc.v2"

	_ "net/http/pprof"
)

var (
	network   = flag.String("network", "", `DNS name to connect to (e.g. "robustirc.net"). The _robustirc._tcp SRV record must be present.`)
	tlsCAFile = flag.String("tls_ca_file", "", "Use the specified file as trusted CA instead of the system CAs. Useful for testing.")

	listenHttp = flag.String("listen_http", "", "[host]:port on which to serve debug handlers (if non-empty)")

	channels          = flag.String("channels", "", "channels the bot should join. Space separated.")
	nick              = flag.String("nick", "frank", "nickname of the bot")
	admins            = flag.String("admins", "xeen", "users who can control the bot. Space separated.")
	nickserv_password = flag.String("nickserv_password", "", "password used to identify with nickserv. No action is taken if password is blank or not set.")

	verbose = flag.Bool("verbose", false, "enable to get very detailed logs")
)

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
		log.Fatalf("Could not create RobustIRC session: %v", err)
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
		log.Fatalf("RobustIRC session error: %v", err)
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
	if *nickserv_password != "" {
		Post(fmt.Sprintf("PASS nickserv=%s", *nickserv_password))
	}
	Post(fmt.Sprintf("NICK %s", *nick))
	Post(fmt.Sprintf("USER bot 0 * :%s von Bötterich", *nick))
	setupJoinChannels()
}

func main() {
	setupFlags()

	if *listenHttp != "" {
		go func() {
			log.Fatal(http.ListenAndServe(*listenHttp, nil))
		}()
	}

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
	ListenerAdd("greeter", runnerGreet)
	ListenerAdd("manpages", runnerManpages)
	// Keep this last, so that other runners can access the name lists
	ListenerAdd("updateMembers", runnerMembers)

	ListenerAdd("topicchanger", runnerTopicChanger)

	if *verbose {
		ListenerAdd("verbose debugger", func(parsed *irc.Message) error {
			log.Printf("< PREFIX=%s COMMAND=%s PARAMS=%s TRAILING=%s", parsed.Prefix, parsed.Command, parsed.Params, parsed.Trailing())
			return nil
		})
	}

	ListenerAdd("nickname checker", func(parsed *irc.Message) error {
		if parsed.Command == irc.ERR_NICKNAMEINUSE {
			log.Printf("Nickname is already in use. Sleeping for a minute before restarting.")
			time.Sleep(time.Minute)
			log.Printf("Killing now due to nickname being in use")
			kill()
		}
		return nil
	})

	for raw := range session.Messages {
		msg := irc.ParseMessage(raw)
		if msg == nil {
			continue // message could not be parsed
		}

		if msg.Command == irc.PONG {
			continue
		}

		if err := listenersRun(msg); err != nil {
			log.Printf("error processing %q (%#v): %v", raw, msg, err)
		}
	}
}
