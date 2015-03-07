package main

import (
	parser "github.com/husio/go-irc"
	"log"
)

type Listener func(parser.Message) bool

var listeners []Listener

func ListenerAdd(l Listener) {
	log.Printf("Adding Listener: %v", l)
	listeners = append(listeners, l)
}

func listenersRun(parsed parser.Message) {
	for idx, listener := range listeners {
		if listener == nil {
			continue
		}

		go func(idx int, listener Listener) {
			if !listener(parsed) {
				log.Printf("Removing Listener %d: %v", idx, listener)
				listeners[idx] = nil
			}
		}(idx, listener)
	}
}
