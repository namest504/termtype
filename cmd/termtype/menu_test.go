package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func rkey(r rune) *tcell.EventKey { return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone) }

func TestSortedThemesCozyFirst(t *testing.T) {
	names := sortedThemeNames()
	if len(names) < 3 || names[0] != "cozy" || names[1] != "log" {
		t.Fatalf("theme order wrong: %v", names)
	}
}

func TestCarouselWraps(t *testing.T) {
	m := newMenuModel("cozy")
	m.handleKey(key(tcell.KeyLeft))
	if m.idx != len(m.themes)-1 {
		t.Fatalf("left from first should wrap to last, got %d", m.idx)
	}
	m.handleKey(key(tcell.KeyRight))
	if m.idx != 0 {
		t.Fatalf("right should wrap back to first, got %d", m.idx)
	}
}

func TestExpandSelectCollapse(t *testing.T) {
	m := newMenuModel("cozy")
	m.handleKey(key(tcell.KeyDown))
	if !m.expanded || m.sel != m.idx {
		t.Fatal("down should expand with selection on current theme")
	}
	m.handleKey(key(tcell.KeyDown)) // move selection
	m.handleKey(key(tcell.KeyEnter))
	if m.expanded || m.idx != 1 {
		t.Fatalf("enter should pick sel and collapse, idx=%d expanded=%v", m.idx, m.expanded)
	}
}

func TestExpandedEscCollapsesWithoutQuit(t *testing.T) {
	m := newMenuModel("cozy")
	m.handleKey(key(tcell.KeyDown))
	if act := m.handleKey(key(tcell.KeyEscape)); act != actNone || m.expanded {
		t.Fatalf("esc while expanded should just collapse, got act=%v", act)
	}
	if act := m.handleKey(key(tcell.KeyEscape)); act != actQuit {
		t.Fatalf("esc while collapsed should quit, got %v", act)
	}
}

func TestMenuActions(t *testing.T) {
	m := newMenuModel("cozy")
	if act := m.handleKey(key(tcell.KeyEnter)); act != actStart {
		t.Fatalf("enter → start, got %v", act)
	}
	if act := m.handleKey(rkey('s')); act != actSettings {
		t.Fatalf("s → settings, got %v", act)
	}
	if act := m.handleKey(rkey('h')); act != actHistory {
		t.Fatalf("h → history, got %v", act)
	}
}

func TestRestoresSavedTheme(t *testing.T) {
	names := sortedThemeNames()
	m := newMenuModel(names[len(names)-1])
	if m.idx != len(names)-1 {
		t.Fatalf("saved theme not restored, idx=%d", m.idx)
	}
}
