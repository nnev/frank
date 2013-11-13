package frank

import (
	"testing"
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

	for to, from := range topics {
		if x := updateTopicText(to); x != from {
			t.Errorf("updateTopicText(%v) = %v, want %v", to, x, from)
		}
	}
}
