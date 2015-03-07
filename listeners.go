package main

import (
	"log"
	"sync"
)

type Listener func(Message) bool

var listenersMutex sync.Mutex
var listeners []Listener

func ListenerAdd(l Listener) {
	log.Printf("Adding Listener: %v", l)
	listenersMutex.Lock()
	listeners = append(listeners, l)
	listenersMutex.Unlock()
}

func listenerRemove(listener Listener) {
	log.Printf("Removing Listener: %v", listener)

	listenersMutex.Lock()
	index := -1
	for idx, l := range listeners {
		if &listener == &l {
			index = idx
		}
	}

	if index > 0 {
		listeners[index] = listeners[len(listeners)-1]
		listeners = listeners[0 : len(listeners)-1]
	}

	listenersMutex.Unlock()
}

func listenersReset() {
	log.Printf("Resetting listeners")

	listenersMutex.Lock()
	listeners = []Listener{}
	listenersMutex.Unlock()
}

func listenersRun(parsed Message) {
	listenersMutex.Lock()
	temp := make([]Listener, len(listeners))
	copy(temp, listeners)
	listenersMutex.Unlock()

	for _, listener := range temp {
		go func(l Listener) {
			keep := l(parsed)
			if !keep {
				listenerRemove(l)
			}
		}(listener)
	}
}
