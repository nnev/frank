package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"gopkg.in/sorcix/irc.v2"
)

var ErrNoResponse = errors.New("extractPost did not return a response")

func runnerLmgtfy(parsed *irc.Message) error {
	tgt := Target(parsed)
	msg := parsed.Trailing()

	if !strings.HasPrefix(tgt, "#") {
		// only answer to this in channels
		return nil
	}

	reply, err := lmgtfyReplyFor(msg)
	if err != nil {
		if err == ErrNoResponse {
			return nil
		}
		Privmsg(tgt, fmt.Sprintf("Error: %v", err))
		return nil
	}
	Privmsg(tgt, fmt.Sprintf("[LMGTFY] %s", reply))

	return nil
}

func googleLucky(ctx context.Context, query string) (*url.URL, error) {
	u, err := url.Parse("https://www.google.com/search?btnI=1&q=")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "frank IRC Bot")

	cl := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Only follow redirects within Google.
			if !strings.Contains(req.URL.Host, ".google.") {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := cl.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer func() {
		// Exhaust r.Body to prevent Keep-Alive breakage
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}()

	if got, want := resp.StatusCode, http.StatusFound; got != want {
		// get a few characters (to not exceed the RobustIRC message
		// length and IRC netiquette) of the body into the error
		// message
		body, _ := ioutil.ReadAll(&io.LimitedReader{R: resp.Body, N: 200})
		return nil, fmt.Errorf("unexpected HTTP status code for %q: got %d, want %d. body: %s", u.String(), got, want, string(body))
	}

	return resp.Location()
}

var lmgtfyMatcher = regexp.MustCompile(`^(?:[\d\pL._-]+: )?lmgtfy:? (.+)`)

func lmgtfyReplyFor(msg string) (string, error) {
	match := lmgtfyMatcher.FindStringSubmatch(msg)
	if match == nil {
		return "", ErrNoResponse
	}
	if len(match) < 2 {
		return "", fmt.Errorf("could not extract lmgtfy query from %q", msg)
	}
	query := match[1]

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()
	u, err := googleLucky(ctx, query)
	if err != nil {
		return "", fmt.Errorf("Error googling %q: %v", query, err)
	}
	result := u.String()

	c := &http.Client{Timeout: 10 * time.Second}
	if title, _, err := TitleGet(c, result); err == nil {
		return fmt.Sprintf("%s @ %s", title, result), nil
	}

	// Fall back to the result URL
	return result, nil
}
