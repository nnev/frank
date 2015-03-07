package main

import (
	"log"
)

type Listener func(Message) bool

var listeners []Listener

func ListenerAdd(l Listener) {
	log.Printf("Adding Listener: %v", l)
	listeners = append(listeners, l)
}

func listenersReset() {
	log.Printf("Resetting listeners")
	listeners = []Listener{}
}

func listenersRun(parsed Message) {
	for idx, listener := range listeners {
		if listener == nil {
			continue
		}

		go func(i int, l Listener) {
			keep := l(parsed)
			if !keep {
				log.Printf("Removing Listener %d: %v", i, l)
				listeners[i] = nil
			}
		}(idx, listener)
	}
}
