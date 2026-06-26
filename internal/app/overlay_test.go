package app

import (
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

type fakeTheme struct{}

func (fakeTheme) ResetState(*domain.GameState)                        {}
func (fakeTheme) UpdateScreen(r domain.Renderer, _ *domain.GameState) { r.Clear(); r.Show() }
func (fakeTheme) OnTick(*domain.GameState)                            {}

func rowString(ss tcell.SimulationScreen, y int) string {
	cells, w, _ := ss.GetContents()
	var b strings.Builder
	for x := 0; x < w; x++ {
		c := cells[y*w+x]
		if len(c.Runes) > 0 && c.Runes[0] != 0 {
			b.WriteRune(c.Runes[0])
		} else {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func newGame(ss tcell.SimulationScreen, st *domain.GameState) *Game {
	return &Game{screen: ss, renderer: ui.NewRenderer(ss), theme: fakeTheme{}, state: st}
}

// The time-attack countdown is drawn on the top row and the pause banner in the
// middle of the screen.
func TestOverlay_TimeAttackAndPause(t *testing.T) {
	ss := tcell.NewSimulationScreen("")
	if err := ss.Init(); err != nil {
		t.Fatal(err)
	}
	ss.SetSize(60, 12)

	st := &domain.GameState{
		TimeLimit:    60 * time.Second,
		TimerStarted: true,
		StartTime:    time.Now().Add(-15 * time.Second),
	}
	st.TogglePause() // pause (~45s should remain, frozen)
	g := newGame(ss, st)
	g.render()

	if top := rowString(ss, 0); !strings.Contains(top, "⏱") {
		t.Errorf("top row missing countdown clock: %q", top)
	}
	if mid := rowString(ss, 6); !strings.Contains(mid, "PAUSED") {
		t.Errorf("middle row missing pause banner: %q", mid)
	}
}

// Normal mode draws no countdown overlay.
func TestOverlay_NormalHasNoClock(t *testing.T) {
	ss := tcell.NewSimulationScreen("")
	if err := ss.Init(); err != nil {
		t.Fatal(err)
	}
	ss.SetSize(60, 12)

	g := newGame(ss, &domain.GameState{TimerStarted: true, StartTime: time.Now()})
	g.render()

	if top := rowString(ss, 0); strings.Contains(top, "⏱") {
		t.Errorf("normal mode should have no clock, got: %q", top)
	}
}
