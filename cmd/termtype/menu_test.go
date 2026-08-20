package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// key and rkey live in settings_test.go, shared across this package's tests.

func TestSortedThemesCozyFirst(t *testing.T) {
	names := sortedThemeNames()
	if len(names) < 3 || names[0] != "cozy" || names[1] != "log" {
		t.Fatalf("sortedThemeNames() = %v, want cozy, log, ...", names)
	}
}

func TestCarouselWraps(t *testing.T) {
	t.Run("left from first wraps to last", func(t *testing.T) {
		m := newMenuModel("cozy")
		m.handleKey(key(tcell.KeyLeft))
		if got, want := m.idx, len(m.themes)-1; got != want {
			t.Errorf("idx = %d, want %d", got, want)
		}
	})
	t.Run("right from last wraps to first", func(t *testing.T) {
		m := newMenuModel("cozy")
		m.handleKey(key(tcell.KeyLeft)) // move to last first
		m.handleKey(key(tcell.KeyRight))
		if got, want := m.idx, 0; got != want {
			t.Errorf("idx = %d, want %d", got, want)
		}
	})
}

func TestExpandSelectCollapse(t *testing.T) {
	m := newMenuModel("cozy")
	t.Run("down expands with selection on current theme", func(t *testing.T) {
		m.handleKey(key(tcell.KeyDown))
		if !m.expanded || m.sel != m.idx {
			t.Errorf("expanded=%v sel=%d idx=%d, want expanded=true and sel==idx", m.expanded, m.sel, m.idx)
		}
	})
	t.Run("enter picks the selection and collapses", func(t *testing.T) {
		m.handleKey(key(tcell.KeyDown)) // move selection
		m.handleKey(key(tcell.KeyEnter))
		if m.expanded || m.idx != 1 {
			t.Errorf("expanded=%v idx=%d, want expanded=false and idx=1", m.expanded, m.idx)
		}
	})
}

func TestExpandedEscCollapsesWithoutQuit(t *testing.T) {
	m := newMenuModel("cozy")
	m.handleKey(key(tcell.KeyDown))
	t.Run("esc while expanded collapses without quitting", func(t *testing.T) {
		if act := m.handleKey(key(tcell.KeyEscape)); act != actNone || m.expanded {
			t.Errorf("act = %v, expanded = %v, want actNone and collapsed", act, m.expanded)
		}
	})
	t.Run("esc while collapsed quits", func(t *testing.T) {
		if act := m.handleKey(key(tcell.KeyEscape)); act != actQuit {
			t.Errorf("act = %v, want actQuit", act)
		}
	})
}

func TestMenuActions(t *testing.T) {
	cases := []struct {
		name string
		k    *tcell.EventKey
		want menuAction
	}{
		{"enter starts", key(tcell.KeyEnter), actStart},
		{"s opens settings", rkey('s'), actSettings},
		{"h opens history", rkey('h'), actHistory},
	}
	m := newMenuModel("cozy")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := m.handleKey(tc.k); got != tc.want {
				t.Errorf("handleKey() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRestoresSavedTheme(t *testing.T) {
	names := sortedThemeNames()
	m := newMenuModel(names[len(names)-1])
	if got, want := m.idx, len(names)-1; got != want {
		t.Fatalf("idx = %d, want %d (saved theme not restored)", got, want)
	}
}
