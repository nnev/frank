package main

import (
	"database/sql"
	"testing"
	"time"
)

func TestInsertNextEvent(t *testing.T) {
	// overwrite DB query function to return locally defined event
	evt := event{}
	getNextEvent = func() *event {
		return &evt
	}

	// setup different possibilities and expected results
	date := time.Date(2014, 4, 23, 18, 12, 0, 0, time.UTC)
	evtTreff := event{
		Stammtisch: false,
		Override:   toNullString(""),
		Location:   toNullString(""),
		Date:       date,
		Topic:      toNullString("Testing"),
	}
	strTreff := ROBOT_BLOCK_IDENTIFIER + " 2014-04-23: c¼h: Testing"

	evtStammtisch := event{
		Stammtisch: true,
		Override:   toNullString(""),
		Location:   toNullString("Mr. Woot"),
		Date:       date,
		Topic:      toNullString(""),
	}
	strStammtisch := ROBOT_BLOCK_IDENTIFIER + " 2014-04-23: Stammtisch @ Mr. Woot https://www.noname-ev.de/yarpnarp.html bitte zu/absagen"

	evtSpecial := event{
		Stammtisch: false,
		Override:   toNullString("RGB2R"),
		Location:   toNullString(""),
		Date:       time.Now(),
		Topic:      toNullString(""),
	}
	strSpecial := ROBOT_BLOCK_IDENTIFIER + " HEUTE (" + time.Now().Format("02.Jan") + "): Ausnahmsweise: RGB2R"

	strOld := ROBOT_BLOCK_IDENTIFIER + " Derp"

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
			if x := insertNextEvent(from); x != to {
				t.Errorf("insertNextEvent(%v)\n GOT: %v\nWANT: %v", from, x, to)
			}
		}
	}
}

func toNullString(s string) sql.NullString {
	return sql.NullString{Valid: true, String: s}
}
