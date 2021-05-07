package main

import (
	"database/sql"
	"testing"
	"time"
)

func TestInsertNextEvent(t *testing.T) {
	// overwrite DB query function to return locally defined event
	evt := event{}
	getNextEvent = func() (*event, error) {
		return &evt, nil
	}

	// setup different possibilities and expected results
	date := time.Date(2014, 4, 23, 18, 12, 0, 0, time.UTC)
	evtTreff := event{
		stammtisch: false,
		override:   "",
		location:   sql.NullString{String: "garbage", Valid: false},
		date:       date,
		topic:      sql.NullString{String: "Testing", Valid: true},
		speaker:    sql.NullString{String: "Test-Speaker", Valid: true},
	}
	strTreff := RobotBlockIdentifier + " 2014-04-23: c¼h: Testing von Test-Speaker"

	evtStammtisch := event{
		stammtisch: true,
		override:   "",
		location:   sql.NullString{String: "Mr. Woot", Valid: true},
		date:       date,
		topic:      sql.NullString{String: "GARBAGE", Valid: false},
		speaker:    sql.NullString{String: "GaRbAgE", Valid: false},
	}
	strStammtisch := RobotBlockIdentifier + " 2014-04-23: Stammtisch @ Mr. Woot https://www.noname-ev.de/yarpnarp.html bitte zu/absagen"

	now := time.Now()
	evtSpecial := event{
		stammtisch: false,
		override:   "RGB2R",
		location:   sql.NullString{String: "gArBaGe", Valid: false},
		date:       now,
		topic:      sql.NullString{String: "GArbAGe", Valid: false},
		speaker:    sql.NullString{String: "gaRBagE", Valid: false},
	}
	strSpecial := RobotBlockIdentifier + " HEUTE (" + now.Format("02.Jan") + "): Ausnahmsweise: RGB2R"

	strOld := RobotBlockIdentifier + " Derp"

	// Test if replacement works correctly
	evt = evtTreff

	var tests = map[event]map[string]string{
		evtTreff: map[string]string{
			"NoName":                         "NoName | " + strTreff,
			"NoName | " + strOld:             "NoName | " + strTreff,
			"NoName | " + strOld + " | Derp": "NoName | " + strTreff + " | Derp",
		},
		evtStammtisch: map[string]string{
			"NoName": "NoName | " + strStammtisch,
		},
		evtSpecial: map[string]string{
			"NoName": "NoName | " + strSpecial,
		},
	}

	for curEvt, topics := range tests {
		evt = curEvt
		for from, to := range topics {
			newTopic, err := replaceTopic(from)
			if err != nil {
				t.Fatal(err)
			}
			if newTopic != to {
				t.Errorf("insertNextEvent(%v)\n GOT: %q\nWANT: %q", from, newTopic, to)
			}
		}
	}
}
