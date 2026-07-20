package main

import (
	"strings"
	"testing"
	"time"

	"github.com/namest504/termtype/internal/store"
)

func TestHistoryLine(t *testing.T) {
	r := store.Round{
		TS:    time.Date(2026, 7, 20, 17, 30, 0, 0, time.UTC),
		Theme: "cozy", Mode: "ta15", Lang: "en", Source: "words",
		WPM: 76.4, Acc: 99.02,
	}
	line := historyLine(r)
	for _, want := range []string{"2026-07-20 17:30", "cozy", "ta15", "en", "words", "76 wpm", "99.0%"} {
		if !strings.Contains(line, want) {
			t.Errorf("historyLine = %q, missing %q", line, want)
		}
	}
}

func TestListTop(t *testing.T) {
	cases := []struct {
		top, sel, visible, want int
	}{
		{0, 0, 10, 0},
		{0, 9, 10, 0},   // selection still inside the window
		{0, 10, 10, 1},  // one past the window scrolls by one
		{5, 3, 10, 3},   // selection above the window scrolls up
		{5, 20, 10, 11}, // far below jumps the window down
	}
	for _, c := range cases {
		if got := listTop(c.top, c.sel, c.visible); got != c.want {
			t.Errorf("listTop(%d, %d, %d) = %d, want %d", c.top, c.sel, c.visible, got, c.want)
		}
	}
}
