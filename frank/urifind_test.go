package frank

import (
	"fmt"
	"testing"
)

func TestExtract(t *testing.T) {
	var msgs = make(map[string][]string)
	msgs["Ich finde http://github.com/lol toll, aber http://heise.de besser"] = []string{"http://github.com/lol", "http://heise.de"}
	msgs["dort (http://deinemudda.de) gibts geile pics"] = []string{"http://deinemudda.de"}
	msgs["http://heise.de, letztens gefunden."] = []string{"http://heise.de"}
	msgs["http-rfc ist doof"] = []string{}
	msgs["http://http://foo.de, letztens gefunden."] = []string{"http://http://foo.de"}
	msgs["http://http://foo.de letztens gefunden"] = []string{"http://http://foo.de"}
	msgs["sECuRE: failed Dein Algo nicht auf https://maps.google.de/maps?q=Frankfurt+(Oder)&hl=de ?"] = []string{"https://maps.google.de/maps?q=Frankfurt+(Oder)&hl=de"}

	for from, to := range msgs {
		x := fmt.Sprintf("%v", extract(from))
		to := fmt.Sprintf("%v", to)

		if x != to {
			t.Errorf("extract(%v)\n GOT: %v\nWANT: %v", from, x, to)
		}
	}
}
