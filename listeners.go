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

func listenersRun(parsed Message) {
	for idx, listener := range listeners {
		if listener == nil {
			continue
		}

		go func(idx int, listener Listener) {
			keep := listener(parsed)
			if !keep {
				log.Printf("Removing Listener %d: %v", idx, listener)
				listeners[idx] = nil
			}
		}(idx, listener)
	}
}
