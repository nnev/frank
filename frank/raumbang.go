package frank

import (
	irc "github.com/fluffle/goirc/client"
	"log"
	"os/exec"
	"regexp"
	"time"
)

const hostToPing = "chaostreff.vpn.zekjur.net"

// only answer !raum from this channel
const bangRaumChannel = "#chaos-hd"

var bangRaumRegex = regexp.MustCompile(`/^!raum($|\s)/i`)
var bangRaumLast = time.Now().Add(time.Second * -10)

func RaumBang(conn *irc.Conn, line *irc.Line) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	tgt := line.Args[0]
	msg := line.Args[1]

	if tgt != bangRaumChannel || !bangRaumRegex.MatchString(msg) {
		return
	}

	dur := time.Since(bangRaumLast)

	if dur.Seconds() <= 10 {
		log.Printf("WTF: last room stat request was %v seconds ago, skipping", dur)
		return
	}

	log.Printf("Received room stat request from %s\n", line.Nick)
	bangRaumLast = time.Now()

	err := exec.Command("ping", "-q", "-l 3", "-c 3", "-w 1", hostToPing).Run()
	if err != nil {
		conn.Privmsg(tgt, "Raumstatus: Die Weltnetzanfrage wurde nicht erwidert.")
	} else {
		conn.Privmsg(tgt, "Raumstatus: Ein Gerät innerhalb des Raumes beantwortet Weltnetzanfragen.")
	}
}
