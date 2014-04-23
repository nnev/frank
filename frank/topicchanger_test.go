package frank

import (
	"testing"
	"time"
)

func TestAdvanceDates(t *testing.T) {
	var topics = make(map[string]string)
	topics["NoName e.V. | Chaostreff Heidelberg | morgen: Suche nach cLFV bei LHCb"] = "NoName e.V. | Chaostreff Heidelberg | heute: Suche nach cLFV bei LHCb"
	topics["NoName e.V. | heute: Suche nach cLFV bei LHCb"] = "NoName e.V."
	topics["NoName e.V. | HEUTE: Suche nach cLFV bei LHCb"] = "NoName e.V."
	topics["MORGEN: derp"] = "HEUTE: derp"
	topics["HEUTE: derp"] = ""
	topics["Verein | 2b || !2b | morgen komische Topics"] = "Verein | 2b || !2b | heute komische Topics"
	topics["Verein | 2b || !2b | heute komische Topics"] = "Verein | 2b || !2b"

	dateYesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	dateToday := time.Now().Format("2006-01-02")
	dateTomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	dateDayAfterTomorrow := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

	topics[dateToday+": derp"] = "HEUTE: derp"
	topics[dateToday+" derp"] = "HEUTE derp"
	topics[dateYesterday] = dateYesterday
	topics[dateDayAfterTomorrow+" | derp"] = dateDayAfterTomorrow + " | derp"
	topics[dateTomorrow+" | derp"] = "MORGEN | derp"

	for from, to := range topics {
		if x := advanceDates(from); x != to {
			t.Errorf("advanceDates(%v)\n GOT: %v\nWANT: %v", from, x, to)
		}
	}
}

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
		Override:   "",
		Location:   "",
		Date:       date,
		Topic:      "Testing",
	}
	strTreff := ROBOT_BLOCK_IDENTIFIER + " 2014-04-23: c¼h: Testing"

	evtStammtisch := event{
		Stammtisch: true,
		Override:   "",
		Location:   "Mr. Woot",
		Date:       date,
		Topic:      "",
	}
	strStammtisch := ROBOT_BLOCK_IDENTIFIER + " 2014-04-23: Stammtisch @ Mr. Woot https://www.noname-ev.de/yarpnarp.html bitte zu/absagen"

	evtSpecial := event{
		Stammtisch: false,
		Override:   "RGB2R",
		Location:   "",
		Date:       date,
		Topic:      "",
	}
	strSpecial := ROBOT_BLOCK_IDENTIFIER + " 2014-04-23: Ausnahmsweise: RGB2R"

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
