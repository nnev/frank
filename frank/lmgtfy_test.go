package frank

import (
	"testing"
)

func TestExtractPost(t *testing.T) {
	var samples = make(map[string]string)
	samples["xeen: lmgtfy: xeens deine mudda nacktbilder"] = "[LMGTFY] xeens deine mudda nacktbilder - Google Search @ http://www.google.com/search?btnI=1&q=xeens+deine+mudda+nacktbilder" //taken from the channel
	samples["lmgtfy: google maps"] = "[LMGTFY] Google Maps @ https://maps.google.com/maps?output=classic&dg=brw"
	samples["lmgtfy: yrden my mail setup"] = "[LMGTFY] yrden my mail setup - Google Search @ http://www.google.com/search?btnI=1&q=yrden+my+mail+setup"
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
