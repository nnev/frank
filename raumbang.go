package main

import (
	"log"
	"os/exec"
	"regexp"
	"time"
)

const hostToPing = "chaostreff.vpn.zekjur.net"

var bangRaumRegex = regexp.MustCompile(`(?i)^!raum($|\s)`)
var bangRaumLast = time.Now().Add(time.Second * -5)

func runnerRaumbang(parsed Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	if !IsPrivateQuery(parsed) {
		return
	}

	if !bangRaumRegex.MatchString(parsed.Trailing) {
		return
	}

	dur := time.Since(bangRaumLast)

	if dur.Seconds() <= 5 {
		log.Printf("WTF: last room stat request was %v seconds ago, skipping", dur)
		return
	}

	log.Printf("Received room stat request from %s", Nick(parsed))
	bangRaumLast = time.Now()

	n := Nick(parsed)

	err := exec.Command("ping", "-q", "-l 3", "-c 3", "-w 1", hostToPing).Run()
	if err != nil {
		Privmsg(n, "No reply, so room is probably not yet open.")
	} else {
		Privmsg(n, "Pluta replies, so the room is likley open \\o/")
	}
}
