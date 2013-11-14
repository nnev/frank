package frank

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
