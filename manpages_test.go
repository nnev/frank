package main

import (
	"reflect"
	"testing"
)

func TestManpageExtract(t *testing.T) {
	tcs := []struct {
		Msg    string
		Expect []string
	}{
		{"foo", nil},
		{"foo bar", nil},
		{"frank: foo bar", nil},
		{"foo(1)", []string{"https://manpages.debian.org/foo.1"}},
		{"foo(3pl)", []string{"https://manpages.debian.org/foo.3pl"}},
		{"das kann man in foo(1) nachlesen", []string{"https://manpages.debian.org/foo.1"}},
		{"das kann man in foo(3pl) nachlesen", []string{"https://manpages.debian.org/foo.3pl"}},
		{"das kann man in foo(1) oder bar(3) nachlesen", []string{"https://manpages.debian.org/foo.1", "https://manpages.debian.org/bar.3"}},
		{"man foo", nil},
		{"frank: man foo", []string{"https://manpages.debian.org/foo"}},
	}

	for _, tc := range tcs {
		got := extractManpages(tc.Msg)
		if len(got) == 0 && len(tc.Expect) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, tc.Expect) {
			t.Errorf("extractManpages(%q) = %q, expected %q", tc.Msg, got, tc.Expect)
		}
	}
}
