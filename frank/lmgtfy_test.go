package frank

import (
	"testing"
)

func TestExtractPost(t *testing.T) {
	var samples = make(map[string]string)
    samples["xeen: lmgtfy: xeens deine mudda nacktbilder"] = "[LMGTFY] xeens deine mudda nacktbilder - Google Search @ http://www.google.com/search?btnI=1&q=xeens+deine+mudda+nacktbilder" //taken from the channel
    samples["lmgtfy: google"] = "[LMGTFY] Google Maps @ https://maps.google.com/"
    samples["lmgtfy: schach"] = "[LMGTFY] Schach – Wikipedia @ http://de.wikipedia.org/wiki/Schach"
    samples["buaitrnosups"] = ""
    samples["warum funktioniert lmgtfy nicht?"] = ""
    samples["lmgtfy lmgtfy"] = "[LMGTFY] Let me google that for you @ http://lmgtfy.com/"

	for msg, post := range samples {
		x := extractPost(msg)
		if x != post {
			t.Errorf("extractPost(%v)\n GOT: ||%v||\nWANT: ||%v||", msg, x, post)
		}
	}
}
