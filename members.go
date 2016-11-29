package main

import (
	"log"
	"strings"
	"sync"
)

type membersMap struct {
	m   map[string]map[string]bool
	mtx sync.RWMutex
}

func (m membersMap) initChannel(channel string) {
	if m.m[channel] == nil {
		m.m[channel] = make(map[string]bool)
	}
}

func (m membersMap) add(nick, channel string) {
	m.initChannel(channel)
	m.m[channel][nick] = true
}

func (m membersMap) remove(nick, channel string) {
	m.initChannel(channel)
	delete(m.m[channel], nick)
}

func (m membersMap) rename(from, to string) {
	for _, c := range m.m {
		if _, ok := c[from]; !ok {
			continue
		}
		delete(c, from)
		c[to] = true
	}
}

func (m membersMap) IsMember(nick, channel string) bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	c := m.m[channel]
	if c == nil {
		return false
	}
	return c[nick]
}

var members = membersMap{m: make(map[string]map[string]bool)}

func runnerMembers(parsed Message) {
	interesting := map[string]bool{
		"353":  true,
		"PART": true,
		"QUIT": true,
		"JOIN": true,
		"NICK": true,
	}
	if !interesting[parsed.Command] {
		return
	}

	members.mtx.Lock()
	defer members.mtx.Unlock()

	switch parsed.Command {
	case "353": // Names
		channel := parsed.Params[len(parsed.Params)-1]
		for _, n := range strings.Split(parsed.Trailing, " ") {
			n = strings.TrimSpace(strings.TrimLeft(n, "~&@%+"))
			if n != "" {
				members.add(n, channel)
			}
		}
	case "PART":
		channel := Target(parsed)
		nick := Nick(parsed)
		members.remove(nick, channel)
	case "QUIT":
		nick := Nick(parsed)
		for _, c := range members.m {
			delete(c, nick)
		}
	case "JOIN":
		channel := parsed.Trailing
		nick := Nick(parsed)
		members.add(nick, channel)
	case "NICK":
		from := Nick(parsed)
		to := parsed.Trailing
		members.rename(from, to)
	}
	log.Printf("Members is now %v", members.m)
}

func IsMember(nick, channel string) bool {
	return members.IsMember(nick, channel)
}
