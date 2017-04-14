package main

import (
	"log"
	"time"

	"gopkg.in/sorcix/irc.v2"

	"golang.org/x/sync/errgroup"
)

type Runner func(*irc.Message) error

type Listener struct {
	desc    string
	created string
	runner  Runner
}

var listeners []*Listener

func ListenerAdd(desc string, r Runner) {
	log.Printf("Adding Listener for: %s", desc)
	listeners = append(listeners, &Listener{
		runner:  r,
		desc:    desc,
		created: time.Now().Format("2006-01-02 15:04:05 -0700"),
	})
}

func listenersRun(msg *irc.Message) error {
	var wg errgroup.Group
	for _, l := range listeners {
		l := l // copy
		wg.Go(func() error {
			return l.runner(msg)
		})
	}
	return wg.Wait()
}
