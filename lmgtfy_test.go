package main

import (
	"strings"
	"testing"
)

func TestExtractPost(t *testing.T) {
	// Comment out when testing. Google changes results regularly, making this system test failing too often
	return

	var samples = make(map[string]string)
	samples["xeen: lmgtfy: xeens deine mudda nacktbilder"] = "[LMGTFY] frank/lmgtfy_test.go at master · breunigs/frank · GitHub @ https://github.com/breunigs/frank/blob/master/frank/lmgtfy_test.go" //taken from the channel
	samples["lmgtfy: google maps"] = "[LMGTFY] Google Maps @ https://"
	samples["lmgtfy: yrden my mail setup"] = "[LMGTFY] yrden my mail setup - Google Search @ http://www.google.com/search?btnI=1&q=yrden+my+mail+setup"
	samples["buaitrnosups"] = ""
	samples["warum funktioniert lmgtfy nicht?"] = ""
	samples["lmgtfy lmgtfy"] = "[LMGTFY] Let me google that for you @ http://lmgtfy.com/"

	for msg, post := range samples {
		x := extractPost(msg)
		if !strings.HasPrefix(x, post) {
			t.Errorf("extractPost(%v)\n GOT: ||%v||\nWANT: ||%v||", msg, x, post)
		}
	}
}
