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
