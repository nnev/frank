package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"text/template"
	"time"
)

var lastSeenLimit = 30 * 24 * time.Hour
var lastSeenWriteThresh = time.Minute
var greeting *template.Template

// maps channel -> (nick -> last seen)
var lastSeen = struct {
	mtx sync.Mutex
	m   map[string]map[string]time.Time
}{}

func init() {
	readLastSeen()
	readGreeting()
}

func touchLastSeen(channel string, nick string) (absent time.Duration) {
	lastSeen.mtx.Lock()
	defer lastSeen.mtx.Unlock()

	c := lastSeen.m[channel]
	if c == nil {
		c = make(map[string]time.Time)
		lastSeen.m[channel] = c
	}
	t := c[nick]
	c[nick] = time.Now()
	return time.Now().Sub(t)
}

func runnerGreet(parsed Message) {
	botnick := *nick
	if botnick == Nick(parsed) {
		// we ignore ourselves
		return
	}
	nick := Nick(parsed)

	var channel string
	switch parsed.Command {
	case "JOIN":
		channel = parsed.Trailing
	case "PART":
		channel = Target(parsed)
	case "PRIVMSG":
		channel = Target(parsed)
	default:
		return
	}

	// TODO: Make channels configurable
	if channel != "#chaos-hd" {
		return
	}

	if channel == "" || channel[0] != '#' {
		log.Printf("Not greeting in non-channel %q", channel)
		return
	}

	// To handle renames of users correctly, we also save the hostmask. Only if
	// we've seen neither it's a genuinely new user.
	absentNick := touchLastSeen(channel, nick)
	absentHostmask := touchLastSeen(channel, Hostmask(parsed))
	seen := (absentNick <= lastSeenLimit) || (absentHostmask <= lastSeenLimit)
	if parsed.Command == "JOIN" && !seen {
		log.Printf("I have not seen %q in %q recently, so I'm greeting them", nick, channel)

		params := struct {
			Nick string
			Bot  string
		}{nick, botnick}

		var msg string

		msgBuf := new(bytes.Buffer)
		if greeting == nil {
			msg = fmt.Sprintf("Hey %s! o/", nick)
		} else if err := greeting.Execute(msgBuf, params); err != nil {
			log.Println("Could not render greeting:", err)
			msg = fmt.Sprintf("Hey %s! o/", nick)
		} else {
			msg = msgBuf.String()
		}
		Privmsg(channel, msg)
	}

	if absentNick > lastSeenWriteThresh || absentHostmask > lastSeenWriteThresh {
		writeLastSeen()
	}
}

func writeLastSeen() {
	lastSeen.mtx.Lock()
	defer lastSeen.mtx.Unlock()

	log.Println("Writing last-seen")

	// Take out the garbage
	for _, c := range lastSeen.m {
		for nick, last := range c {
			if time.Now().Sub(last) > lastSeenLimit {
				delete(c, nick)
			}
		}
	}

	// We write the file atomically to prevent corruption
	tmp, err := ioutil.TempFile(".", ".last-seen")
	if err != nil {
		log.Println("Could not create temporary file for last-seen:", err)
		return
	}
	defer os.Remove(tmp.Name())

	if err := gob.NewEncoder(tmp).Encode(lastSeen.m); err != nil {
		log.Println("Could not write last-seen:", err)
		return
	}

	if err := os.Rename(tmp.Name(), "last-seen"); err != nil {
		log.Println("Could not write last-seen:", err)
		return
	}
}

func readLastSeen() {
	lastSeen.mtx.Lock()
	defer lastSeen.mtx.Unlock()

	// Make sure, lastSeen is initialized to a sane, if empty, value.
	defer func() {
		if lastSeen.m == nil {
			lastSeen.m = make(map[string]map[string]time.Time)
		}
	}()

	f, err := os.Open("last-seen")
	if err != nil {
		log.Println("Could not read last-seen:", err)
		return
	}
	defer f.Close()

	var m map[string]map[string]time.Time
	if err := gob.NewDecoder(f).Decode(&m); err != nil {
		log.Println("Could not read last-seen:", err)
		return
	}
	lastSeen.m = m
}

func readGreeting() {
	t, err := template.ParseFiles("greeting.txt")
	if err != nil {
		log.Println("Could not parse greeting:", err)
		return
	}
	greeting = t
}
