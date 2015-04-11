package main

import (
	"log"
	"os/exec"
	"regexp"
	"time"
)

const hostToPing = "chaostreff.vpn.zekjur.net"

// only answer !raum from this channel
const bangRaumChannel = "#chaos-hd"

var bangRaumRegex = regexp.MustCompile(`(?i)^!raum($|\s)`)
var bangRaumLast = time.Now().Add(time.Second * -10)

func runnerRaumbang(parsed Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	tgt := Target(parsed)
	msg := parsed.Trailing

	if tgt != bangRaumChannel || !bangRaumRegex.MatchString(msg) {
		return
	}

	dur := time.Since(bangRaumLast)

	if dur.Seconds() <= 10 {
		log.Printf("WTF: last room stat request was %v seconds ago, skipping", dur)
		return
	}

	log.Printf("Received room stat request from %s", Nick(parsed))
	bangRaumLast = time.Now()

	err := exec.Command("ping", "-q", "-l 3", "-c 3", "-w 1", hostToPing).Run()
	if err != nil {
		Privmsg(tgt, "Raumstatus: Die Weltnetzanfrage wurde nicht erwidert.")
	} else {
		Privmsg(tgt, "Raumstatus: Ein Gerät innerhalb des Raumes beantwortet Weltnetzanfragen.")
	}
}
