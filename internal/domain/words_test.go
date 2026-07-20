package domain

import (
	"strings"
	"testing"
)

func TestCommonWordsClean(t *testing.T) {
	if len(CommonWords) < 150 {
		t.Fatalf("CommonWords has %d words, want at least 150", len(CommonWords))
	}
	seen := map[string]bool{}
	for _, w := range CommonWords {
		if w == "" || w != strings.ToLower(w) || strings.ContainsAny(w, " \t.,!?'") {
			t.Errorf("word %q is not a clean lowercase word", w)
		}
		if seen[w] {
			t.Errorf("word %q appears twice in the pool", w)
		}
		seen[w] = true
	}
}

func TestRandomWordsCount(t *testing.T) {
	for _, n := range []int{1, 25, 250} {
		got := strings.Fields(RandomWords(n))
		if len(got) != n {
			t.Errorf("RandomWords(%d) has %d words", n, len(got))
		}
		for _, w := range got {
			if !contains(CommonWords, w) {
				t.Errorf("RandomWords produced %q, not in CommonWords", w)
			}
		}
	}
	if RandomWords(0) != "" {
		t.Error("RandomWords(0) should be empty")
	}
}

func TestRandomWordsNoImmediateRepeat(t *testing.T) {
	words := strings.Fields(RandomWords(500))
	for i := 1; i < len(words); i++ {
		if words[i] == words[i-1] {
			t.Fatalf("immediate repeat %q at %d", words[i], i)
		}
	}
}

func TestTargetGenOverridesPool(t *testing.T) {
	gs := &GameState{
		Sentences: []string{"from the pool"},
		TargetGen: func() string { return "generated stream" },
	}
	if got := gs.RandomSentence(); got != "generated stream" {
		t.Errorf("RandomSentence = %q, want the TargetGen output", got)
	}
	gs.TargetGen = nil
	if got := gs.RandomSentence(); got != "from the pool" {
		t.Errorf("RandomSentence = %q, want the pool sentence", got)
	}
}

func contains(pool []string, w string) bool {
	for _, p := range pool {
		if p == w {
			return true
		}
	}
	return false
}
