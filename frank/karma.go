package frank

import (
	"bytes"
	"encoding/gob"
	irc "github.com/fluffle/goirc/client"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const karmaFile = "karma"

// regex that matches karma additions
var karmaMatcherRegex = regexp.MustCompile(`^([\d\pL]+)(\+\+|--)(?:$|\s#)`)

// regex that matches karma info requests
var karmaAnswerRegex = regexp.MustCompile(`(?i)^karma\s+(?:for\s+)?([\d\pL]+)\??$`)

// create default and try to read saved file in immediately
var defaultData = map[string]int{"frank": 9999}
var data = readData()

func Karma(conn *irc.Conn, line *irc.Line) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MEGA-WTF:pkg: %v", r)
		}
	}()

	match(conn, line)
	answer(conn, line)
}

// reads the current line for karma-esque expressions and ups/dows the
// thing that was voted on. A user can’t vote on her/himself.
func match(conn *irc.Conn, line *irc.Line) {
	tgt := line.Args[0]
	msg := line.Args[1]

	if tgt[0:1] != "#" {
		// love/hate needs to be announced publicly to avoid skewing the
		// results
		return
	}

	if !karmaMatcherRegex.MatchString(msg) {
		return
	}

	match := karmaMatcherRegex.FindStringSubmatch(msg)

	if len(match) < 3 {
		log.Printf("WTF: regex match didn’t have enough parts")
		return
	}

	thing := strings.ToLower(match[1])

	if thing == strings.ToLower(line.Nick) {
		log.Printf("User %s tried to karma her/himself. What a loser!", line.Nick)
		conn.Notice(line.Nick, "[Karma] Voting on yourself is not supported")
		return
	}

	if match[2] == "++" {
		data[thing] += 1
	} else {
		data[thing] -= 1
	}

	log.Printf("%s karma for: %s  (total: %v)", thing, match[1], data[match[1]])
	writeData()
}

// answers a user with the current karma for a given thing
func answer(conn *irc.Conn, line *irc.Line) {
	tgt := line.Args[0]
	msg := line.Args[1]

	if !karmaAnswerRegex.MatchString(msg) {
		return
	}

	match := karmaAnswerRegex.FindStringSubmatch(msg)

	if len(match) != 2 || match[1] == "" {
		log.Printf("WTF: karma answer regex somehow failed and produced invalid results")
		return
	}

	score := strconv.Itoa(data[strings.ToLower(match[1])])

	// if we were the target, it was a private message. Answer user instead
	if tgt == conn.Me().Nick {
		tgt = line.Nick
	}
	conn.Notice(tgt, "[Karma] "+match[1]+": "+score)
}

// via http://golang.worleyspace.com/2011/10/blog-post.html
func writeData() {
	//initialize a *bytes.Buffer
	m := new(bytes.Buffer)
	//the *bytes.Buffer satisfies the io.Writer interface and can
	//be used in gob.NewEncoder()
	enc := gob.NewEncoder(m)
	//gob.Encoder has method Encode that accepts data items as parameter
	enc.Encode(data)
	//the bytes.Buffer type has method Bytes() that returns type []byte,
	//and can be used as a parameter in ioutil.WriteFile()
	err := ioutil.WriteFile(karmaFile, m.Bytes(), 0600)
	if err != nil {
		log.Printf("WTF: Couldn’t write %v: %v\n", karmaFile, err)
		return
	}
	log.Printf("just saved gob with\n")
}

// via http://golang.worleyspace.com/2011/10/blog-post.html
func readData() map[string]int {
	//read the file that was just written, n is []byte
	n, err := ioutil.ReadFile(karmaFile)
	if err != nil {
		log.Printf("WTF: Couldn’t read %v: %v\n", karmaFile, err)
		return defaultData
	}
	//create a bytes.Buffer type with n, type []byte
	p := bytes.NewBuffer(n)
	//bytes.Buffer satisfies the interface for io.Writer and can be used
	//in gob.NewDecoder()
	dec := gob.NewDecoder(p)
	data := map[string]int{}
	//we must decode into a pointer, so we'll take the address of data
	err = dec.Decode(&data)
	if err != nil {
		log.Printf("WTF: Couldn’t parse %v: %v\n", karmaFile, err)
		return defaultData
	}
	log.Printf("just read gob from file and it's showing: %v\n", data)
	return data
}
