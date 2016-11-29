package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPDFTitleGet(t *testing.T) {
	var files = make(map[string]string)
	files["samples/nada.pdf"] = ""
	files["samples/yes.pdf"] = "TITLE by AUTHOR"

	for filepath, expected := range files {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf, _ := ioutil.ReadFile(filepath)
			fmt.Fprintln(w, string(buf))
		}))
		defer ts.Close()

		title := PDFTitleGet(ts.URL)
		if title != expected {
			t.Errorf("TestPDFTitleGet(%v)\n GOT: %v\nWANT: %v", "from", title, expected)
		}
	}
}

func TestExtract(t *testing.T) {
	var msgs = make(map[string][]string)
	msgs["Ich finde http://github.com/lol toll, aber http://heise.de besser"] = []string{"http://github.com/lol", "http://heise.de"}
	msgs["dort (http://deinemudda.de) gibts geile pics"] = []string{"http://deinemudda.de"}
	msgs["http://heise.de, letztens gefunden."] = []string{"http://heise.de"}
	msgs["http-rfc ist doof"] = []string{}
	msgs["http://http://foo.de, letztens gefunden."] = []string{"http://http://foo.de"}
	msgs["http://http://foo.de letztens gefunden"] = []string{"http://http://foo.de"}
	msgs["sECuRE: failed Dein Algo nicht auf https://maps.google.de/maps?q=Frankfurt+(Oder)&hl=de ?"] = []string{"https://maps.google.de/maps?q=Frankfurt+(Oder)&hl=de"}
	msgs["(nested parens http://en.wikipedia.org/wiki/Heuristic_(engineering))"] = []string{"http://en.wikipedia.org/wiki/Heuristic_(engineering)"}
	msgs["enclosed by parens: (http://en.wikipedia.org/wiki/Heuristic_(engineering))"] = []string{"http://en.wikipedia.org/wiki/Heuristic_(engineering)"}

	for from, to := range msgs {
		x := fmt.Sprintf("%v", extract(from))
		to := fmt.Sprintf("%v", to)

		if x != to {
			t.Errorf("extract(%v)\n GOT: %v\nWANT: %v", from, x, to)
		}
	}
}

func TestTitleGet(t *testing.T) {
	var samples = make(map[string]string)
	samples["https://twitter.com/meganfinger/status/444586462076346368"] = "Megan Finger on Twitter: THANK YOU Central for my awesome email address and username...... Like really😒 http://t.co/G6ab4Ikoik"
	samples["https://twitter.com/dave_tucker/status/400269131255390210"] = "Dave Tucker on Twitter: This morning the wife asked “Why is your phone issuing you death threats?”. Me: “Oh it’s just my new alarm clock” /cc @CARROT_app"
	samples["http://twitter.com/dave_tucker/status/400269131255390210"] = "Dave Tucker on Twitter: This morning the wife asked “Why is your phone issuing you death threats?”. Me: “Oh it’s just my new alarm clock” /cc @CARROT_app"
	samples["https://twitter.com/Perspective_pic/status/400356645504831489/photo/1"] = "Destroying Stuff on Twitter: Sorry but this without a doubt the greatest thing ever seen on an air duct http://t.co/Om5qq4HLBu"
	samples["https://twitter.com/Perspective_pic/status/400356645504831489"] = "Destroying Stuff on Twitter: Sorry but this without a doubt the greatest thing ever seen on an air duct http://t.co/Om5qq4HLBu"
	samples["https://twitter.com/quityourjrob/status/405438033853313025/photo/1"] = "Joanna Robinson on Twitter: How to tell if a toy is for boys or girls. http://t.co/4MTdubGZo1"
	samples["https://twitter.com/rechelon/status/431242278221275137"] = "William Gillis ⚑ on Twitter: @SebastosPublius @jfsmith23 Yep. Godesky had gathered a large following back then and was more sane than Zerzan & less terrible than Jensen."
	samples["https://twitter.com/thejeremyvine/status/433607774375649280"] = "Jeremy Vine on Twitter: The internet was invented so someone could ask this question - and get an answer: http://t.co/MRJUGFqFMr"
	samples["http://twitter.com/thejeremyvine/status/433607774375649280"] = "Jeremy Vine on Twitter: The internet was invented so someone could ask this question - and get an answer: http://t.co/MRJUGFqFMr"
	samples["https://twitter.com/bhalp1/status/578925947245633536"] = "Ben Halpern on Twitter: Sometimes when I'm writing Javascript I want to throw up my hands and say \"this is bullshit!\" but I can never remember what \"this\" refers to"
	samples["http://www.spiegel.de/schulspiegel/abi/abitur-schueler-beantragt-klausur-nach-informationsfreiheitsgesetz-a-1027298.html"] = "Abitur: Schüler beantragt Klausur nach Informationsfreiheitsgesetz - SPIEGEL ONLINE"
	samples["https://github.com/breunigs/frank"] = "Frank is an IRC-Bot written in Go. It’s my pet project to learn Go and specifically tailored to my needs."
	samples["https://github.com/breunigs/python-librtmp-debian"] = "GitHub - breunigs/python-librtmp-debian"
	samples["http://forum.xda-developers.com/xposed/modules/mod-rootcloak-completely-hide-root-t2574647"] = "[MOD][XPOSED][4.0+] RootCloak - Completely H… | Xposed General"
	samples["https://code.facebook.com/posts/1433093613662262/-under-the-hood-facebook-s-cold-storage-system-"] = "Under the hood: Facebook’s cold storage system | Engineering Blog | Facebook Code | Facebook"
	samples["http://genius.cat-v.org/rob-pike/"] = "Rob Pike"

	for url, title := range samples {
		x, _, _ := TitleGet(url)
		if !strings.HasSuffix(x, title) {
			t.Errorf("TitleGet(%v)\n GOT: ||%v||\nWANT: ||%v||", url, x, title)
		}
	}
}

func TestClean(t *testing.T) {
	if x := clean("x‏‎​   x‏"); x != "x x" {
		t.Errorf("clean does not remove all whitespace/non-printable chars (got: %v)", x)
	}

	if x := clean(" trim "); x != "trim" {
		t.Errorf("clean does not trim properly (got: %v)", x)
	}
}

func TestCache(t *testing.T) {
	if cc := cacheGetByUrl("fakeurl"); cc != nil {
		t.Errorf("Empty Cache should return nil pointer")
	}

	cacheAdd("realurl", "some title")

	if cc := cacheGetByUrl("fakeurl"); cc != nil {
		t.Errorf("Cache should return nil pointer when URL not cached")
	}

	cc := cacheGetByUrl("realurl")

	if cc == nil {
		t.Errorf("Cache should find cached URL")
	}

	if cc.title != "some title" {
		t.Errorf("Cache did not return expected title (returned: %#v)", cc)
	}

	if ago := cacheGetTimeAgo(cc); ago != "0m" {
		t.Errorf("Cache did not produce expected time ago value. Expected: 0m. Returned: %s", ago)
	}

	tmp, _ := time.ParseDuration("-1h1m")
	cc.date = time.Now().Add(tmp)
	if ago := cacheGetTimeAgo(cc); ago != "1h" {
		t.Errorf("Cache did not produce expected time ago value. Expected: 1h. Returned: %s", ago)
	}

	tmp, _ = time.ParseDuration("-1h31m")
	cc.date = time.Now().Add(tmp)
	if ago := cacheGetTimeAgo(cc); ago != "2h" {
		t.Errorf("Cache did not produce expected time ago value. Expected: 2h. Returned: %s", ago)
	}

	cacheAdd("secondsAgoTestUrl", "another title")
	time.Sleep(time.Second)
	if secs := cacheGetSecondsToLastPost("another title"); secs != 1 {
		t.Errorf("Cache did not calculate correct amount of seconds since post. Got: %s, Expected: 1s", secs)
	}
}
