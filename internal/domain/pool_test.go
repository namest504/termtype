package domain

import "testing"

func TestRandomPoolSentenceIgnoresTargetGen(t *testing.T) {
	gs := &GameState{
		Sentences: []string{"from the pool"},
		TargetGen: func() string { return "generated stream" },
	}
	if got := gs.RandomPoolSentence(); got != "from the pool" {
		t.Errorf("RandomPoolSentence = %q, want the pool sentence", got)
	}
	if got := gs.RandomSentence(); got != "generated stream" {
		t.Errorf("RandomSentence = %q, want the TargetGen output", got)
	}
}
