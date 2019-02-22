package main

import (
	"fmt"
	"io"
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

const simpleTitleBody = `<!DOCTYPE html>
<html>
<head>
<title>simple title</title>
</head>
`

const truncatedBody = `<!DOCTYPE html>
<html>
<head>
<title>title from truncated body</title>
</he
`

const emptyTitleBody = `<!DOCTYPE html>
<html>
<head>
<title></title>
</head>
`

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type stringDoer string

func (sd *stringDoer) Do(r *http.Request) (*http.Response, error) {
	return &http.Response{
		Request:    r,
		StatusCode: http.StatusOK,
		Body:       nopCloser{strings.NewReader(string(*sd))},
	}, nil
}

// TODO: non-200 responses
// TODO: long title truncation

func TestTitleGet(t *testing.T) {
	const irrelevantURL = "http://localhost" // irrelevant but valid
	for _, want := range []struct {
		body  string
		title string
	}{
		{
			body:  simpleTitleBody,
			title: "simple title",
		},
		{
			body:  truncatedBody,
			title: "title from truncated body",
		},
		{
			body:  emptyTitleBody,
			title: "",
		},
	} {
		want := want // capture
		t.Run(want.title, func(t *testing.T) {
			t.Parallel()
			sd := stringDoer(want.body)
			if got, _, _ := TitleGet(&sd, irrelevantURL); !strings.HasSuffix(got, want.title) {
				t.Errorf("unexpected title: got %q, want %q suffix", got, want.title)
			}
		})
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
		t.Errorf("Cache did not calculate correct amount of seconds since post. Got: %v, Expected: 1s", secs)
	}
}
