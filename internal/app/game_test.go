package app

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/domain"
)

func typeRunes(g *Game, s string) {
	for _, r := range s {
		g.handleKeyEvent(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
}

// stubTheme is a minimal domain.Theme whose ResetState mirrors SimpleTheme:
// it resets the round and draws a fresh target from the pool, which the
// overlay_test.go fakeTheme (an empty no-op ResetState) does not do.
type stubTheme struct{}

func (stubTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
}
func (stubTheme) UpdateScreen(r domain.Renderer, gs *domain.GameState) {}
func (stubTheme) OnTick(gs *domain.GameState)                          {}

// newTestGame returns a Game built through NewGame (not the newGame test
// helper) so NewGame's own construction logic is exercised: sentence-pool
// fallback, autoGraph-vs-cozy, and initial ResetState via the theme.
func newTestGame(t *testing.T, autoGraph bool, themeName string) *Game {
	t.Helper()
	s := tcell.NewSimulationScreen("UTF-8")
	if err := s.Init(); err != nil {
		t.Fatalf("init sim screen: %v", err)
	}
	s.SetSize(80, 24)
	t.Cleanup(s.Fini)
	g, err := NewGame(s, stubTheme{}, 0, []string{"ab"}, nil,
		RoundMeta{Theme: themeName, Lang: "en", Source: "builtin"}, nil, autoGraph)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	return g
}

func TestNewGame(t *testing.T) {
	t.Run("empty sentence pool falls back to default", func(t *testing.T) {
		s := tcell.NewSimulationScreen("UTF-8")
		if err := s.Init(); err != nil {
			t.Fatalf("init: %v", err)
		}
		t.Cleanup(s.Fini)
		g, err := NewGame(s, stubTheme{}, 0, nil, nil, RoundMeta{Theme: "simple"}, nil, false)
		if err != nil {
			t.Fatalf("NewGame: %v", err)
		}
		if len(g.state.Sentences) == 0 {
			t.Error("sentences should fall back to the default English pool")
		}
	})
	t.Run("cozy theme disables the auto graph", func(t *testing.T) {
		if g := newTestGame(t, true, "cozy"); g.autoGraph {
			t.Error("autoGraph must be forced off on the cozy theme")
		}
	})
	t.Run("other themes keep the auto graph", func(t *testing.T) {
		if g := newTestGame(t, true, "simple"); !g.autoGraph {
			t.Error("autoGraph should stay on for non-cozy themes")
		}
	})
}

func TestHandleKeyEvent_GameLifecycle(t *testing.T) {
	t.Run("esc goes back, ctrl-c quits", func(t *testing.T) {
		g := newTestGame(t, false, "simple")
		if back, quit := g.handleKeyEvent(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)); !back || quit {
			t.Errorf("Esc = (%v,%v), want (true,false)", back, quit)
		}
		if back, quit := g.handleKeyEvent(tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModNone)); !back || !quit {
			t.Errorf("Ctrl-C = (%v,%v), want (true,true)", back, quit)
		}
	})
	t.Run("pause swallows input", func(t *testing.T) {
		g := newTestGame(t, false, "simple")
		typeRunes(g, "a") // start the timer first so pause is meaningful
		g.handleKeyEvent(tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone))
		before := g.state.UserInput
		typeRunes(g, "b")
		if g.state.UserInput != before {
			t.Errorf("input while paused changed UserInput to %q", g.state.UserInput)
		}
		g.handleKeyEvent(tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone))
		typeRunes(g, "b")
		if g.state.UserInput == before {
			t.Error("input after resume should register")
		}
	})
	t.Run("backspace removes the last rune", func(t *testing.T) {
		g := newTestGame(t, false, "simple")
		typeRunes(g, "a")
		g.handleKeyEvent(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
		if g.state.UserInput != "" {
			t.Errorf("UserInput = %q, want empty after backspace", g.state.UserInput)
		}
	})
	t.Run("typing the full target finishes the round", func(t *testing.T) {
		g := newTestGame(t, false, "simple")
		typeRunes(g, g.state.TargetSentence)
		if !g.state.IsFinished {
			t.Fatal("round should finalize when the target is fully typed")
		}
	})
	t.Run("g toggles the graph view after finishing", func(t *testing.T) {
		g := newTestGame(t, false, "simple")
		typeRunes(g, g.state.TargetSentence)
		g.handleKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))
		if !g.showGraph {
			t.Error("g should raise the graph view")
		}
		g.handleKeyEvent(tcell.NewEventKey(tcell.KeyRune, 'g', tcell.ModNone))
		if g.showGraph {
			t.Error("second g should dismiss the graph view")
		}
	})
	t.Run("enter after finishing starts a new round", func(t *testing.T) {
		g := newTestGame(t, false, "simple")
		typeRunes(g, g.state.TargetSentence)
		g.handleKeyEvent(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
		if g.state.IsFinished {
			t.Error("Enter should reset the round")
		}
	})
}

func TestDrawGraphViewSmoke(t *testing.T) {
	g := newTestGame(t, false, "simple")
	typeRunes(g, g.state.TargetSentence)
	g.state.WPMSamples = []float64{30, 45, 50, 48}
	g.showGraph = true
	g.render() // routes to drawGraphView when finished+showGraph
}

// BUG 5 regression: for multibyte sentences, completion and accuracy must be
// rune-based, not byte-based.
// "héllo" is 5 runes / 6 bytes. Typing it perfectly should give 100% accuracy
// (a byte-based denominator would give 83.3%).
func TestHandleKeyEvent_MultibyteAccuracy(t *testing.T) {
	g := &Game{state: &domain.GameState{TargetSentence: "héllo"}}
	typeRunes(g, "héllo")

	if !g.state.IsFinished {
		t.Fatalf("should be finished after typing all runes")
	}
	if g.state.Accuracy != 100 {
		t.Errorf("Accuracy = %.2f, want 100 (denominator should be rune-based)", g.state.Accuracy)
	}
}

// Rune-based completion: must not be finished before the 5th rune.
func TestHandleKeyEvent_NotFinishedEarly(t *testing.T) {
	g := &Game{state: &domain.GameState{TargetSentence: "héllo"}}
	typeRunes(g, "héll") // 4 runes (5 bytes) — under 6 bytes, so unfinished either way
	if g.state.IsFinished {
		t.Errorf("should be unfinished after typing 4 runes (rune-based)")
	}
}

// ASCII sentence accuracy: with one wrong char, the denominator must be the rune count.
func TestHandleKeyEvent_AsciiAccuracy(t *testing.T) {
	g := &Game{state: &domain.GameState{TargetSentence: "abcd"}}
	typeRunes(g, "abXd") // 3 of 4 chars correct
	if !g.state.IsFinished {
		t.Fatalf("should be finished")
	}
	if g.state.Accuracy != 75 {
		t.Errorf("Accuracy = %.2f, want 75", g.state.Accuracy)
	}
}
