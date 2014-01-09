package frank

import (
	"testing"
	"time"
)

func TestUpdateTopicText(t *testing.T) {
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
		if x := updateTopicText(from); x != to {
			t.Errorf("updateTopicText(%v)\n GOT: %v\nWANT: %v", from, x, to)
		}
	}
}
