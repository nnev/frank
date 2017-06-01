package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/sorcix/irc.v2"
)

const karmaFile = "karma"

var (
	karmaMatcherRegex = regexp.MustCompile(`^([\d\pL]+)(\+\+|--)(?:$|\s#)`)
	karmaAnswerRegex  = regexp.MustCompile(`(?i)^karma:?\s+(?:for\s+)?([\d\pL]+)\??$`)
)

var defaultData = map[string]int{"frank": 9999}
var data = func() map[string]int {
	f, err := os.Open(karmaFile)
	if err != nil {
		log.Printf("could not open karma file %q: %v", karmaFile, err)
		return defaultData
	}
	defer f.Close()
	result := make(map[string]int)
	if err := gob.NewDecoder(f).Decode(&result); err != nil {
		log.Printf("could not read karma file %q: %v", karmaFile, err)
		return defaultData
	}
	return result
}()

func runnerKarma(msg *irc.Message) error {
	if msg.Command != irc.PRIVMSG {
		return nil
	}
	if err := match(msg); err != nil {
		return err
	}
	return answer(msg)
}

// reads the current line for karma-esque expressions and ups/dows the
// thing that was voted on. A user can’t vote on her/himself.
func match(msg *irc.Message) error {
	if len(msg.Params) < 1 || !strings.HasPrefix(msg.Params[0], "#") {
		// love/hate needs to be announced publicly to avoid skewing the
		// results
		return nil
	}

	matches := karmaMatcherRegex.FindStringSubmatch(msg.Trailing())
	if matches == nil {
		return nil
	}

	thing := strings.ToLower(matches[1])

	nick := msg.Prefix.Name
	if thing == strings.ToLower(nick) {
		log.Printf("User %s tried to karma her/himself. What a loser!", nick)
		Privmsg(nick, "[Karma] Voting on yourself is not supported")
		return nil
	}

	if matches[2] == "++" {
		data[thing] += 1
	} else {
		data[thing] -= 1
	}

	log.Printf("user %q changed karma (using the %s operator) for %q to %d", nick, matches[2], thing, data[thing])

	return writeAtomically(karmaFile, func(w io.Writer) error {
		return gob.NewEncoder(w).Encode(data)
	})
}

// answers a user with the current karma for a given thing
func answer(msg *irc.Message) error {
	if len(msg.Params) < 1 {
		return nil
	}
	target := msg.Params[0]
	if !strings.HasPrefix(target, "#") {
		target = msg.Prefix.Name
	}

	matches := karmaAnswerRegex.FindStringSubmatch(msg.Trailing())
	if matches == nil {
		return nil
	}
	thing := matches[1]
	Privmsg(target, fmt.Sprintf("[Karma] %s: %d", thing, data[strings.ToLower(thing)]))
	return nil
}
