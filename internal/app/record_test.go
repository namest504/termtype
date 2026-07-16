package app

import (
	"strings"
	"testing"
	"time"

	"termtype/internal/domain"
	"termtype/internal/store"
)

// newRecordingGame returns a Game whose round is fully typed and ~10s old,
// ready for finalizeRound, persisting to a store rooted at dir.
func newRecordingGame(t *testing.T, dir string) *Game {
	t.Helper()
	ss := newSimScreen(t, 60, 12)
	st := &domain.GameState{TargetSentence: "hello world", UserInput: "hello world",
		TimerStarted: true, StartTime: time.Now().Add(-10 * time.Second)}
	g := newGame(ss, st)
	g.meta = RoundMeta{Theme: "simple", Lang: "en", Source: "builtin"}
	g.store = store.New(dir)
	g.history = g.store.LoadHistory()
	return g
}

func TestFinalizeRoundRecordsHistory(t *testing.T) {
	dir := t.TempDir()
	g := newRecordingGame(t, dir)
	g.finalizeRound()

	rounds := store.New(dir).LoadHistory()
	if len(rounds) != 1 {
		t.Fatalf("history has %d rounds, want 1", len(rounds))
	}
	r := rounds[0]
	if r.Theme != "simple" || r.Lang != "en" || r.Mode != "normal" || r.Source != "builtin" {
		t.Errorf("recorded round meta = %+v", r)
	}
	if r.WPM <= 0 || r.DurS < 9 {
		t.Errorf("recorded round stats = %+v, want positive wpm and ~10s duration", r)
	}
	if !g.state.IsFinished {
		t.Error("finalizeRound did not finalize the game state")
	}
}

func TestFinalizeRoundFirstRoundIsNewBest(t *testing.T) {
	g := newRecordingGame(t, t.TempDir())
	g.finalizeRound()
	if !strings.Contains(g.resultLine, "NEW BEST") {
		t.Errorf("resultLine = %q, want NEW BEST on first eligible round", g.resultLine)
	}
}

func TestFinalizeRoundShowsBestAndSparkline(t *testing.T) {
	dir := t.TempDir()
	seed := store.New(dir)
	for i := 0; i < 3; i++ {
		seed.AppendRound(store.Round{TS: time.Now(), Theme: "simple", Mode: "normal",
			Lang: "en", WPM: 200, Acc: 99, DurS: 10, Source: "builtin"})
	}
	g := newRecordingGame(t, dir)
	g.finalizeRound()
	if !strings.Contains(g.resultLine, "best 200") {
		t.Errorf("resultLine = %q, want existing best 200 shown", g.resultLine)
	}
	if strings.Contains(g.resultLine, "NEW BEST") {
		t.Errorf("resultLine = %q must not claim a new best below 200 wpm", g.resultLine)
	}
}

func TestFinalizeRoundNilStore(t *testing.T) {
	ss := newSimScreen(t, 60, 12)
	st := &domain.GameState{TargetSentence: "hi", UserInput: "hi",
		TimerStarted: true, StartTime: time.Now().Add(-6 * time.Second)}
	g := newGame(ss, st)
	g.finalizeRound() // must not panic with no store/meta configured
	if !g.state.IsFinished {
		t.Error("finalizeRound did not finalize with a nil store")
	}
}
