package main

import (
	"log"
	"strings"
	"sync"

	"gopkg.in/sorcix/irc.v2"
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

func runnerMembers(parsed *irc.Message) error {
	interesting := map[string]bool{
		"353":  true,
		"PART": true,
		"QUIT": true,
		"JOIN": true,
		"NICK": true,
	}
	if !interesting[parsed.Command] {
		return nil
	}

	members.mtx.Lock()
	defer members.mtx.Unlock()

	switch parsed.Command {
	case "353": // Names
		if len(parsed.Params) < 4 {
			return nil
		}
		// parsed.Params[0] is my own nick
		// parsed.Params[1] is a sigil for the channel (“=” public, “@” secret, …)
		// parsed.Params[2] is the name of the channel
		// parsed.Params[3] (or parsed.Trailing()) are the space-separated nicknames
		channel := parsed.Params[2]
		for _, n := range strings.Split(parsed.Trailing(), " ") {
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
		channel := parsed.Trailing()
		nick := Nick(parsed)
		members.add(nick, channel)
	case "NICK":
		from := Nick(parsed)
		to := parsed.Trailing()
		members.rename(from, to)
	}
	log.Printf("Members is now %v", members.m)
	return nil
}

func IsMember(nick, channel string) bool {
	return members.IsMember(nick, channel)
}
