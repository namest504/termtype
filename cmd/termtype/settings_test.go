package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/store"
)

func key(k tcell.Key) *tcell.EventKey { return tcell.NewEventKey(k, 0, tcell.ModNone) }

func TestSettingsModelFromConfig(t *testing.T) {
	m := newSettingsModel(store.Config{Mode: "ta30", Source: "words", Lang: "ko", Graph: "off", Style: "box"})
	if gameModes[m.modeIdx].name != "Time Attack (30s)" {
		t.Fatalf("mode idx wrong: %s", gameModes[m.modeIdx].name)
	}
	if textSources[m.srcIdx].code != "words" || languages[m.langIdx].code != "ko" {
		t.Fatal("source/lang not restored")
	}
	if m.graphOn || chartStyles[m.styleIdx].code != "box" {
		t.Fatal("graph/style not restored")
	}
}

func TestSettingsCycleAndApply(t *testing.T) {
	m := newSettingsModel(store.Config{})
	// row 0 = Mode: Right → Time Attack (15s)
	if changed, _ := m.handleKey(key(tcell.KeyRight)); !changed {
		t.Fatal("right on mode row should report a change")
	}
	// row 4 = Style: Left twice (braille2 → braille1 → box)
	m.row = 4
	m.handleKey(key(tcell.KeyLeft))
	m.handleKey(key(tcell.KeyLeft))
	cfg := m.apply(store.Config{Theme: "cozy"})
	if cfg.Mode != "ta15" || cfg.Style != "box" || cfg.Theme != "cozy" {
		t.Fatalf("apply produced %+v", cfg)
	}
}

func TestSettingsFreshConfigDefaultsToBraille2(t *testing.T) {
	m := newSettingsModel(store.Config{})
	// Fresh config should show braille2 (not braille1)
	if chartStyles[m.styleIdx].code != "braille2" {
		t.Fatalf("fresh config should default to braille2, got %s", chartStyles[m.styleIdx].code)
	}
	// Changing an unrelated row should not downgrade the style
	m.row = 0 // Mode
	m.handleKey(key(tcell.KeyRight))
	cfg := m.apply(store.Config{})
	if cfg.Style != "braille2" {
		t.Fatalf("changing unrelated row should preserve braille2, got %s", cfg.Style)
	}
}

func TestSettingsLanguagePinnedForWords(t *testing.T) {
	m := newSettingsModel(store.Config{Source: "words"})
	m.row = 2 // Language
	if changed, _ := m.handleKey(key(tcell.KeyRight)); changed {
		t.Fatal("language must not cycle while Words is selected")
	}
}

func TestSettingsEscDone(t *testing.T) {
	m := newSettingsModel(store.Config{})
	if _, done := m.handleKey(key(tcell.KeyEscape)); !done {
		t.Fatal("esc should finish the screen")
	}
}

func TestSettingsRowNavigationClamps(t *testing.T) {
	m := newSettingsModel(store.Config{})
	m.handleKey(key(tcell.KeyUp)) // already at top
	if m.row != 0 {
		t.Fatal("up at top should clamp")
	}
	for i := 0; i < 10; i++ {
		m.handleKey(key(tcell.KeyDown))
	}
	if m.row != settingsRows-1 {
		t.Fatalf("down should clamp at %d, got %d", settingsRows-1, m.row)
	}
}
