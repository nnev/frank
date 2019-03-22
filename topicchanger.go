package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/sorcix/irc.v2"

	_ "github.com/lib/pq"
)

// I would prefer 🕖, but it’s not available in most fonts
const RobotBlockIdentifier = "ꜰ"

func TopicChanger() {
	for {
		Post("TOPIC #chaos-hd") // TODO: configurable channel
		time.Sleep(5 * time.Minute)
	}
}

func runnerTopicChanger(msg *irc.Message) error {
	var topic string

	channel := "#chaos-hd" // TODO: configurable channel

	switch msg.Command {
	// A user changed the topic
	case irc.TOPIC:
		if len(msg.Params) < 1 || msg.Params[0] != channel {
			return nil
		}
		return updateTopic(channel, msg.Trailing())

	// We received a reply to our periodic TOPIC command.
	case irc.RPL_TOPIC:
		topic = msg.Trailing()
		fallthrough
	case irc.RPL_NOTOPIC:
		if len(msg.Params) < 2 || msg.Params[1] != channel {
			return nil
		}
		return updateTopic(channel, topic)
	}

	return nil
}

const separator = "|"

func replaceTopic(current string) (string, error) {
	nextEvent, err := getNextEvent()
	if err != nil {
		return "", err
	}

	nextEventPart := fmt.Sprintf(" %s %v ", RobotBlockIdentifier, nextEvent)
	parts := strings.Split(current+" ", separator)
	for idx, part := range parts {
		if strings.Contains(part, RobotBlockIdentifier) {
			parts[idx] = nextEventPart
		}
	}
	if !strings.Contains(current, RobotBlockIdentifier) {
		parts = append(parts, nextEventPart)
	}
	return strings.TrimSpace(strings.Join(parts, separator)), nil
}

func updateTopic(channel, currentTopic string) error {
	newTopic, err := replaceTopic(currentTopic)
	if err != nil {
		return err
	}
	if strings.TrimSpace(currentTopic) != strings.TrimSpace(newTopic) {
		log.Printf("updating topic from %q to %q", currentTopic, newTopic)
		Post("TOPIC " + channel + " :" + newTopic)
	}
	return nil
}

type event struct {
	stammtisch bool
	override   string
	location   string
	date       time.Time
	topic      string
	speaker    string
}

func (evt *event) String() string {
	t := ""

	date := evt.date.Format("2006-01-02")
	dateToday := time.Now().Format("2006-01-02")
	dateTomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	switch date {
	case dateToday:
		t += "HEUTE (" + evt.date.Format("02.Jan") + ")"
	case dateTomorrow:
		t += "MORGEN (" + evt.date.Format("02.Jan") + ")"
	default:
		t += date
	}

	t += ": "

	if evt.override != "" {
		t += "Ausnahmsweise: " + evt.override
	} else if evt.stammtisch {
		t += "Stammtisch @ " + evt.location
		t += " https://www.noname-ev.de/yarpnarp.html"
		t += " bitte zu/absagen"
	} else {
		t += "c¼h: " + evt.topic
		if evt.speaker != "" {
			t += " von " + evt.speaker
		}
	}

	return strings.TrimSpace(t)
}

var getNextEvent = func() (*event, error) {
	const nextEventQuery = `
SELECT
  stammtisch,
  override,
  CASE WHEN location = '' OR location IS NULL THEN 'TBA' ELSE location END,
  termine.date,
  CASE WHEN topic = '' OR topic IS NULL THEN 'noch keine ◉︵◉' ELSE topic END,
  speaker
FROM termine
LEFT JOIN vortraege
ON termine.date = vortraege.date
WHERE termine.date >= NOW()::date
ORDER BY termine.date ASC
LIMIT 1
`
	db, err := sql.Open("postgres", "dbname=nnev user=anon host=/var/run/postgresql sslmode=disable")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var e event
	if err := db.QueryRow(nextEventQuery).Scan(
		&e.stammtisch,
		&e.override,
		&e.location,
		&e.date,
		&e.topic,
		&e.speaker); err != nil {
		return nil, err
	}

	if *verbose {
		log.Printf("event from SQL: %#v", e)
	}

	return &e, nil
}
