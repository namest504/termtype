package main

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/store"
)

// newSimScreen returns an initialized in-memory tcell screen.
func newSimScreen(t *testing.T, w, h int) tcell.SimulationScreen {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	s.SetSize(w, h)
	t.Cleanup(s.Fini)
	return s
}

// screenString flattens the screen contents into one searchable string,
// one row per line.
func screenString(s tcell.SimulationScreen) string {
	cells, w, h := s.GetContents()
	var b strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			if len(c.Runes) > 0 {
				b.WriteRune(c.Runes[0])
			} else {
				b.WriteRune(' ')
			}
		}
		b.WriteRune('\n')
	}
	return b.String()
}

func wantOnScreen(t *testing.T, s tcell.SimulationScreen, substrs ...string) {
	t.Helper()
	dump := screenString(s)
	for _, sub := range substrs {
		if !strings.Contains(dump, sub) {
			t.Errorf("screen missing %q; dump:\n%s", sub, dump)
		}
	}
}

func TestDrawMenu(t *testing.T) {
	t.Run("collapsed shows carousel and summary", func(t *testing.T) {
		s := newSimScreen(t, 80, 24)
		m := newMenuModel("cozy")
		drawMenu(s, m, "Normal · Sentences · English")
		wantOnScreen(t, s, "termtype", "cozy", "Normal · Sentences · English", "start")
	})
	t.Run("expanded lists every theme", func(t *testing.T) {
		s := newSimScreen(t, 80, 24)
		m := newMenuModel("cozy")
		m.handleKey(key(tcell.KeyDown))
		drawMenu(s, m, "Normal · Sentences · English")
		wantOnScreen(t, s, sortedThemeNames()...)
	})
	t.Run("narrow terminal does not panic", func(t *testing.T) {
		s := newSimScreen(t, 20, 10)
		m := newMenuModel("cozy")
		drawMenu(s, m, "Normal · Sentences · English")
		wantOnScreen(t, s, "termtype")
	})
}

func TestDrawSettings(t *testing.T) {
	t.Run("shows all five rows", func(t *testing.T) {
		s := newSimScreen(t, 80, 24)
		drawSettings(s, newSettingsModel(store.Config{}))
		wantOnScreen(t, s, "Settings", "Mode", "Text", "Language", "Graph", "Style", "braille")
	})
	t.Run("narrow terminal does not panic", func(t *testing.T) {
		s := newSimScreen(t, 20, 10)
		drawSettings(s, newSettingsModel(store.Config{}))
		wantOnScreen(t, s, "Settings")
	})
}

func TestDrawRoundDetail(t *testing.T) {
	round := store.Round{
		TS:    time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC),
		Theme: "cozy", Mode: "normal", Lang: "en", Source: "builtin",
		WPM: 72.4, Acc: 98.5, DurS: 30,
		WPMSeries: []float64{40, 55, 60, 72, 70},
	}
	t.Run("with series draws graph and summary", func(t *testing.T) {
		s := newSimScreen(t, 80, 24)
		drawRoundDetail(s, round, 80, 24)
		s.Show()
		wantOnScreen(t, s, "wpm: 72", "accuracy: 98.5")
	})
	t.Run("short terminal skips the chart", func(t *testing.T) {
		s := newSimScreen(t, 40, 8)
		drawRoundDetail(s, round, 40, 8)
		s.Show()
		wantOnScreen(t, s, "wpm: 72")
	})
}
