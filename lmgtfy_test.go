package main

import (
	"context"
	"testing"
	"time"
)

func TestGoogleLucky(t *testing.T) {
	const query = "Apple"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	u, err := googleLucky(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	var okay bool
	possibilities := []string{
		"www.apple.com",
		"apple.com",
	}
	for _, p := range possibilities {
		if u.Host == p {
			okay = true
			break
		}
	}
	if !okay {
		t.Fatalf("unexpected Host field of query %q result %q: got %q, want one of %v", query, u.String(), u.Host, possibilities)
	}
}
