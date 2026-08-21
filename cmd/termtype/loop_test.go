package main

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/store"
)

// feedKeys returns a buffered event channel pre-loaded with the given key
// events. The loops under test consume them in order; the channel is left
// open (loops exit via their own key handling, not channel close).
func feedKeys(keys ...*tcell.EventKey) chan tcell.Event {
	ch := make(chan tcell.Event, len(keys))
	for _, k := range keys {
		ch <- k
	}
	return ch
}

// mustTS parses an RFC3339 timestamp, failing the test on error.
func mustTS(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return ts
}

func TestRunMenuQuitsOnEsc(t *testing.T) {
	s := newSimScreen(t, 80, 24)
	cfg := store.Config{}
	if _, err := runMenu(s, feedKeys(key(tcell.KeyEscape)), &cfg, store.New(t.TempDir())); err == nil {
		t.Fatal("Esc should return an error (menu cancelled), got nil")
	}
}

func TestRunMenuStartReturnsSelection(t *testing.T) {
	s := newSimScreen(t, 80, 24)
	cfg := store.Config{Theme: "log", Mode: "ta15", Graph: "off"}
	sel, err := runMenu(s, feedKeys(key(tcell.KeyEnter)), &cfg, store.New(t.TempDir()))
	if err != nil {
		t.Fatalf("Enter should start, got error %v", err)
	}
	if sel.themeName != "log" {
		t.Errorf("themeName = %q, want log", sel.themeName)
	}
	if got := store.ModeString(sel.limit); got != "ta15" {
		t.Errorf("mode = %q, want ta15", got)
	}
	if sel.graphOn {
		t.Error("graphOn = true, want false (config graph off)")
	}
}

func TestRunMenuCarouselPicksTheme(t *testing.T) {
	s := newSimScreen(t, 80, 24)
	cfg := store.Config{Theme: "cozy"}
	// ↓ expand, ↓ move to second theme, Enter select (collapse), Enter start.
	sel, err := runMenu(s, feedKeys(
		key(tcell.KeyDown), key(tcell.KeyDown), key(tcell.KeyEnter), key(tcell.KeyEnter),
	), &cfg, store.New(t.TempDir()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := sortedThemeNames()[1]; sel.themeName != want {
		t.Errorf("themeName = %q, want %q", sel.themeName, want)
	}
}

func TestRunMenuSettingsRoundTrip(t *testing.T) {
	s := newSimScreen(t, 80, 24)
	dir := t.TempDir()
	st := store.New(dir)
	cfg := store.Config{}
	// s → settings, → (mode to ta15), Esc → back to menu, Esc → quit.
	_, err := runMenu(s, feedKeys(
		rkey('s'), key(tcell.KeyRight), key(tcell.KeyEscape), key(tcell.KeyEscape),
	), &cfg, st)
	if err == nil {
		t.Fatal("final Esc should cancel the menu")
	}
	if cfg.Mode != "ta15" {
		t.Errorf("cfg.Mode = %q, want ta15 (settings change must mutate cfg)", cfg.Mode)
	}
	if saved := st.LoadConfig(); saved.Mode != "ta15" {
		t.Errorf("saved config Mode = %q, want ta15 (change must persist immediately)", saved.Mode)
	}
}

func TestRunSettingsSavesEachChange(t *testing.T) {
	s := newSimScreen(t, 80, 24)
	dir := t.TempDir()
	st := store.New(dir)
	cfg := store.Config{}
	runSettings(s, feedKeys(
		key(tcell.KeyDown), key(tcell.KeyDown), key(tcell.KeyDown), key(tcell.KeyDown), // row → Style
		key(tcell.KeyRight), // braille2 → braille3
		key(tcell.KeyEscape),
	), &cfg, st)
	if cfg.Style != "braille3" {
		t.Errorf("cfg.Style = %q, want braille3", cfg.Style)
	}
	if saved := st.LoadConfig(); saved.Style != "braille3" {
		t.Errorf("saved Style = %q, want braille3", saved.Style)
	}
}

func TestShowHistory(t *testing.T) {
	rounds := []store.Round{
		{TS: mustTS(t, "2026-08-19T10:00:00Z"), Theme: "cozy", Mode: "normal", Lang: "en", Source: "builtin", WPM: 60, Acc: 97, DurS: 20, WPMSeries: []float64{50, 60}},
		{TS: mustTS(t, "2026-08-20T10:00:00Z"), Theme: "log", Mode: "ta15", Lang: "en", Source: "words", WPM: 70, Acc: 99, DurS: 15},
	}
	t.Run("empty history escapes cleanly", func(t *testing.T) {
		s := newSimScreen(t, 80, 24)
		showHistory(s, feedKeys(key(tcell.KeyEscape)), nil)
	})
	t.Run("detail and back", func(t *testing.T) {
		s := newSimScreen(t, 80, 24)
		showHistory(s, feedKeys(
			key(tcell.KeyDown), key(tcell.KeyEnter), // open detail of older round
			key(tcell.KeyEscape), key(tcell.KeyEscape), // back to list, then out
		), rounds)
	})
}
