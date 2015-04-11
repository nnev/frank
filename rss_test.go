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
		t.Errorf("List should contain exactly two items, but contains: %v", x)
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
	// simple detect
	feedUpdated := time.Now().Add(time.Minute)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, atomPlanetSample(feedUpdated))
	}))
	defer ts.Close()
	postitems := parseAtomFeed(ts.URL).postableForIrc()

	if len(postitems) != 1 || postitems[0] != "TITLE http://blog.ezelo.de/ipod_shuffle_linux/" {
		t.Errorf("should contain the post as postable for irc")
	}

	// freshness
	feedUpdated = time.Now().Add(-(freshness + 1) * time.Minute)
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, atomPlanetSample(feedUpdated))
	}))
	defer ts.Close()
	postitems = parseAtomFeed(ts.URL).postableForIrc()

	if len(postitems) != 0 {
		t.Errorf("should not contain posts created before freshness period")
	}

	// more than one post
	feedUpdated = time.Now().Add(time.Minute)
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, atomGithubSample(feedUpdated))
	}))
	defer ts.Close()
	postitems = parseAtomFeed(ts.URL).postableForIrc()

	if len(postitems) != 2 {
		t.Errorf("should contain both items, but contains %s", postitems)
	}

	// boottime check
	feedUpdated = time.Now().Add(-time.Minute)
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, atomGithubSample(feedUpdated))
	}))
	defer ts.Close()
	postitems = parseAtomFeed(ts.URL).postableForIrc()

	if len(postitems) != 0 {
		t.Errorf("should not contain any items created before booting %s", postitems)
	}

}

func atomPlanetSample(updated time.Time) string {
	return `<?xml version="1.0"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:planet="http://planet.intertwingly.net/" xmlns:indexing="urn:atom-extension:indexing" indexing:index="no"><access:restriction xmlns:access="http://www.bloglines.com/about/specs/fac-1.0" relationship="deny"/>
  <title>Planet NoName e.V.</title>
  <updated>` + updated.Format("2006-01-02T15:04:05Z") + `</updated>
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
    <updated>` + updated.Format("2006-01-02T15:04:05Z") + `</updated>
    <source>
      <id>http://blog.ezelo.de/</id>
      <author>
        <name>koebi</name>
      </author>
      <link href="http://blog.ezelo.de/" rel="alternate" type="text/html"/>
      <link href="http://blog.ezelo.de/index.xml" rel="self" type="application/rss+xml"/>
      <title>on little indie little … little strange</title>
      <updated>` + updated.Format("2006-01-02T15:04:05Z") + `</updated>
    </source>
  </entry>
</feed>
`
}

func atomGithubSample(updated time.Time) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom" xmlns:media="http://search.yahoo.com/mrss/" xml:lang="en-US">
  <id>tag:github.com,2008:/breunigs/frank/commits/robust</id>
  <link type="text/html" rel="alternate" href="https://github.com/breunigs/frank/commits/robust"/>
  <link type="application/atom+xml" rel="self" href="https://github.com/breunigs/frank/commits/robust.atom"/>
  <title>Recent Commits to frank:robust</title>
  <updated>` + updated.Format("2006-01-02T15:04:05-07:00") + `</updated>
  <entry>
    <id>tag:github.com,2008:Grit::Commit/05c498f24c791776a5c100d099abd5e4976c4af7</id>
    <link type="text/html" rel="alternate" href="https://github.com/breunigs/frank/commit/05c498f24c791776a5c100d099abd5e4976c4af7"/>
    <title>
        fix appendIfMiss logic error. And actually run its tests.
    </title>
    <updated>` + updated.Format("2006-01-02T15:04:05-07:00") + `</updated>
    <media:thumbnail height="30" width="30" url="https://avatars3.githubusercontent.com/u/307954?v=3&amp;s=30"/>
    <author>
      <name>breunigs</name>
      <uri>https://github.com/breunigs</uri>
    </author>
    <content type="html">
      &lt;pre style='white-space:pre-wrap;width:81ex'>fix appendIfMiss logic error. And actually run its tests.&lt;/pre>
    </content>
  </entry>
  <entry>
    <id>tag:github.com,2008:Grit::Commit/cfdee2b55ffb898c3449065cdbc5d6314ec03555</id>
    <link type="text/html" rel="alternate" href="https://github.com/breunigs/frank/commit/cfdee2b55ffb898c3449065cdbc5d6314ec03555"/>
    <title>
        add some tests for RSS parsing to make lil&#39;sECuRE happy
    </title>
    <updated>` + updated.Add(-10*time.Second).Format("2006-01-02T15:04:05-07:00") + `</updated>
    <media:thumbnail height="30" width="30" url="https://avatars3.githubusercontent.com/u/307954?v=3&amp;s=30"/>
    <author>
      <name>breunigs</name>
      <uri>https://github.com/breunigs</uri>
    </author>
    <content type="html">
      &lt;pre style='white-space:pre-wrap;width:81ex'>add some tests for RSS parsing to make lil&#39;sECuRE happy&lt;/pre>
    </content>
  </entry>
</feed>`
}
