package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/chart"
	"github.com/namest504/termtype/internal/store"
)

// key and rkey build synthetic tcell key events for driving handleKey in
// this package's tests. Shared across settings_test.go and menu_test.go.
func key(k tcell.Key) *tcell.EventKey { return tcell.NewEventKey(k, 0, tcell.ModNone) }
func rkey(r rune) *tcell.EventKey     { return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone) }

func TestSettingsModelFromConfig(t *testing.T) {
	m := newSettingsModel(store.Config{Mode: "ta30", Source: "words", Lang: "ko", Graph: "off", Style: "box"})
	if got, want := gameModes[m.modeIdx].name, "Time Attack (30s)"; got != want {
		t.Errorf("mode = %q, want %q", got, want)
	}
	if got, want := textSources[m.srcIdx].code, "words"; got != want {
		t.Errorf("source = %q, want %q", got, want)
	}
	if got, want := languages[m.langIdx].code, "ko"; got != want {
		t.Errorf("lang = %q, want %q", got, want)
	}
	if m.graphOn {
		t.Error("graphOn = true, want false")
	}
	if got, want := chartStyles[m.styleIdx].code, "box"; got != want {
		t.Errorf("style = %q, want %q", got, want)
	}
}

func TestSettingsCycleAndApply(t *testing.T) {
	m := newSettingsModel(store.Config{})
	// row 0 = Mode: Right → Time Attack (15s)
	if changed, _ := m.handleKey(key(tcell.KeyRight)); !changed {
		t.Fatal("right on mode row: changed = false, want true")
	}
	// row 4 = Style: Left twice (braille2 → braille1 → box)
	m.row = 4
	m.handleKey(key(tcell.KeyLeft))
	m.handleKey(key(tcell.KeyLeft))
	cfg := m.apply(store.Config{Theme: "cozy"})
	if cfg.Mode != "ta15" || cfg.Style != "box" || cfg.Theme != "cozy" {
		t.Fatalf("apply() = %+v, want Mode=ta15 Style=box Theme=cozy", cfg)
	}
}

func TestSettingsFreshConfigDefaultsToBraille2(t *testing.T) {
	m := newSettingsModel(store.Config{})
	// Fresh config should show braille2 (not braille1)
	if got, want := chartStyles[m.styleIdx].code, "braille2"; got != want {
		t.Fatalf("fresh config style = %q, want %q", got, want)
	}
	// Changing an unrelated row should not downgrade the style
	m.row = 0 // Mode
	m.handleKey(key(tcell.KeyRight))
	cfg := m.apply(store.Config{})
	if got, want := cfg.Style, "braille2"; got != want {
		t.Fatalf("after unrelated change, style = %q, want %q (preserved)", got, want)
	}
}

// TestUnknownStyleFallsBackToBraille2Consistently guards against
// chartOptionsFor and newSettingsModel disagreeing on the fallback for an
// unknown/legacy style code: both must treat it as braille2, and an
// unrelated settings change must persist "braille2" rather than
// silently rewriting it to "braille1" (index-0 fallback).
func TestUnknownStyleFallsBackToBraille2Consistently(t *testing.T) {
	const unknown = "braille4"

	o := chartOptionsFor(unknown)
	want := chart.Options{Style: chart.StyleBraille, Interp: chart.InterpSmooth, Thickness: 2}
	if o != want {
		t.Fatalf("chartOptionsFor(%q) = %+v, want %+v", unknown, o, want)
	}

	m := newSettingsModel(store.Config{Style: unknown})
	if got, want := chartStyles[m.styleIdx].code, "braille2"; got != want {
		t.Fatalf("newSettingsModel(%q) style = %q, want %q", unknown, got, want)
	}

	m.row = 0 // Mode: unrelated to Style
	m.handleKey(key(tcell.KeyRight))
	cfg := m.apply(store.Config{Style: unknown})
	if got, want := cfg.Style, "braille2"; got != want {
		t.Fatalf("after unrelated change, style = %q, want %q", got, want)
	}
}

func TestSettingsLanguagePinnedForWords(t *testing.T) {
	m := newSettingsModel(store.Config{Source: "words"})
	m.row = 2 // Language
	if changed, _ := m.handleKey(key(tcell.KeyRight)); changed {
		t.Fatal("language changed = true while Words selected, want false")
	}
}

func TestSettingsEscDone(t *testing.T) {
	m := newSettingsModel(store.Config{})
	if _, done := m.handleKey(key(tcell.KeyEscape)); !done {
		t.Fatal("esc: done = false, want true")
	}
}

func TestSettingsRowNavigationClamps(t *testing.T) {
	t.Run("up at top clamps to row 0", func(t *testing.T) {
		m := newSettingsModel(store.Config{})
		m.handleKey(key(tcell.KeyUp))
		if got, want := m.row, 0; got != want {
			t.Errorf("row = %d, want %d", got, want)
		}
	})
	t.Run("down past bottom clamps to last row", func(t *testing.T) {
		m := newSettingsModel(store.Config{})
		for i := 0; i < 10; i++ {
			m.handleKey(key(tcell.KeyDown))
		}
		if got, want := m.row, settingsRows-1; got != want {
			t.Errorf("row = %d, want %d", got, want)
		}
	})
}
