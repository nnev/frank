package frank

import (
	"net"
	"net/http"
	"time"
)

// abort HTTP requests if it takes longer than X seconds. Not sure, it’s
// definitely magic involved. Must be larger than 5.
const httpGetDeadline = 10

func HttpClientWithTimeout() http.Client {
	// via http://www.reddit.com/r/golang/comments/10awvj/timeout_on_httpget/c6bz49s
	return http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(time.Second * httpGetDeadline)
				c, err := net.DialTimeout(netw, addr, time.Second*(httpGetDeadline-5))
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
}
