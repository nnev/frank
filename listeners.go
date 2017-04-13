package main

import (
	"log"
	"sync"
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

var listenersMutex sync.Mutex
var listeners []*Listener

func ListenerAdd(desc string, r Runner) *Listener {
	log.Printf("Adding Listener for: %s", desc)

	l := &Listener{runner: r, desc: desc, created: time.Now().Format("2006-01-02 15:04:05 -0700")}
	l.Add()

	return l
}

func (listener *Listener) Add() {
	listenersMutex.Lock()
	listeners = append(listeners, listener)
	listenersMutex.Unlock()
	listenersDebug()
}

func (listener *Listener) Remove() {
	go func() {
		listenersMutex.Lock()
		index := -1
		for idx, l := range listeners {
			if listener == l {
				index = idx
			}
		}

		if index >= 0 {
			log.Printf("Removing Listener: %s at index %d", listener, index)
			listeners[index] = listeners[len(listeners)-1]
			listeners = listeners[0 : len(listeners)-1]
		} else {
			log.Printf("Removing Listener: %s but was not found in list", listener)
		}

		listenersMutex.Unlock()
		listenersDebug()
	}()
}

func listenersDebug() {
	listenersMutex.Lock()
	s := ""
	for _, listener := range listeners {
		s += listener.desc + ", "
	}
	log.Printf("listeners #%d: %s", len(listeners), s)
	listenersMutex.Unlock()

}

func listenersReset() {
	listenersMutex.Lock()
	listeners = []*Listener{}
	log.Printf("# of listeners: 0")
	listenersMutex.Unlock()
}

func listenersRun(msg *irc.Message) error {
	listenersMutex.Lock()
	defer listenersMutex.Unlock()

	var wg errgroup.Group
	for _, l := range listeners {
		l := l // copy
		wg.Go(func() error {
			return l.runner(msg)
		})
	}
	return wg.Wait()
}
