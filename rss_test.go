package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestRecent(t *testing.T) {
	for i := 0; i < 100; i += 1 {
		addRecentUrl(strconv.Itoa(i))
	}

	if !isRecentUrl("99") {
		t.Errorf("99 should be recent URL")
	}

	if isRecentUrl("1") {
		t.Errorf("1 shouldn’t be recent URL")
	}
}

func TestAppendIfMiss(t *testing.T) {
	x := []string{}

	x = appendIfMiss(x, "test")
	if len(x) != 1 {
		t.Errorf("List should contain exactly one item")
	}

	x = appendIfMiss(x, "test2")
	if len(x) != 2 {
		t.Errorf("List should contain exactly two items")
	}

	if x[0] != "test" {
		t.Errorf("appendIfMiss should append items")
	}

	x = appendIfMiss(x, "test")
	if len(x) != 2 {
		t.Errorf("should not add already present items")
	}
}

func TestLoadURL(t *testing.T) {
	longBody := make([]byte, 1024*1024+1)
	for i := range longBody {
		longBody[i] = 'x'
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(longBody))
	}))
	defer ts.Close()

	content := loadURL(ts.URL)
	if !bytes.Equal(content, longBody[:1024*1024]) {
		t.Errorf("should contain up to 1MB of provided server response")
	}
}

func TestPostableForIrc(t *testing.T) {
	feedUpdated := time.Now().Format("2006-01-02T15:04:05Z")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, atomPlanetSample(feedUpdated, feedUpdated))
	}))
	defer ts.Close()
	postitems := parseAtomFeed(ts.URL).postableForIrc()

	if len(postitems) != 1 || postitems[0] != "TITLE http://blog.ezelo.de/ipod_shuffle_linux/" {
		t.Errorf("should contain the post as postable for irc")
	}

	feedUpdated = time.Now().Add(-(freshness + 1) * time.Minute).Format("2006-01-02T15:04:05Z")
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, atomPlanetSample(feedUpdated, feedUpdated))
	}))
	defer ts.Close()
	postitems = parseAtomFeed(ts.URL).postableForIrc()

	if len(postitems) != 0 {
		t.Errorf("should not contain posts created before freshness period")
	}
}

func atomPlanetSample(feedUpdated string, postUpdated string) string {
	return `<?xml version="1.0"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:planet="http://planet.intertwingly.net/" xmlns:indexing="urn:atom-extension:indexing" indexing:index="no"><access:restriction xmlns:access="http://www.bloglines.com/about/specs/fac-1.0" relationship="deny"/>
  <title>Planet NoName e.V.</title>
  <updated>` + feedUpdated + `</updated>
  <generator uri="http://intertwingly.net/code/venus/">Venus</generator>
  <author>
    <name>NoName e.V. Planet Admin</name>
    <email>webmaster@eris.noname-ev.de</email>
  </author>
  <id>http://blogs.noname-ev.de/atom.xml</id>
  <link href="http://blogs.noname-ev.de/atom.xml" rel="self" type="application/atom+xml"/>
  <link href="http://blogs.noname-ev.de" rel="alternate"/>

  <entry xml:lang="en-us">
    <id>http://blog.ezelo.de/ipod_shuffle_linux/</id>
    <link href="http://blog.ezelo.de/ipod_shuffle_linux/" rel="alternate" type="text/html"/>
    <title>TITLE</title>
    <summary type="xhtml"><div xmlns="http://www.w3.org/1999/xhtml"><h2 id="backstory:5811c604c69e0b5d0428458301a5ab0b">Backstory</h2>

<p>Of course, now, that I had that thing, I wanted to use it. Only to stumble upon the problem of not being able to connect it, because it doesn’t use the “standard” iphone cable, but just a USB-to-3.5mm phone connector.</p>

<h2 id="connecting-an-ipod-to-linux:5811c604c69e0b5d0428458301a5ab0b">Connecting an iPod to linux…</h2>

<pre><code>$ dmesg
[843399.244310] usb 1-1.2: new high-speed USB device number 24 using ehci-pc
</code></pre>
</div>
    </summary>
    <updated>` + postUpdated + `</updated>
    <source>
      <id>http://blog.ezelo.de/</id>
      <author>
        <name>koebi</name>
      </author>
      <link href="http://blog.ezelo.de/" rel="alternate" type="text/html"/>
      <link href="http://blog.ezelo.de/index.xml" rel="self" type="application/rss+xml"/>
      <title>on little indie little … little strange</title>
      <updated>` + postUpdated + `</updated>
    </source>
  </entry>
</feed>
`
}
