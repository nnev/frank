package main

import (
	"strconv"
	"testing"
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

func TestappendIfMiss(t *testing.T) {
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
