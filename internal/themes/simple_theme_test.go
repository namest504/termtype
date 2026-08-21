package themes

import (
	"strings"
	"testing"

	"github.com/namest504/termtype/internal/domain"
)

// TestSimpleCentered verifies the simple theme's typing screen centers the
// sentence vertically, and that the result screen is horizontally centered
// with no leftover "(Esc for menu)" hint.
func TestSimpleCentered(t *testing.T) {
	theme := &SimpleTheme{}
	w, h := 80, 24

	gs := &domain.GameState{Sentences: []string{"hello world"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hello world"

	r := newGridRenderer(w, h)
	theme.UpdateScreen(r, gs)

	sentenceY := -1
	for y := 0; y < h; y++ {
		line := string(r.grid[y])
		if strings.Contains(line, "hello world") {
			sentenceY = y
			break
		}
	}
	if sentenceY == -1 {
		t.Fatal("expected to find the target sentence somewhere on screen")
	}
	if d := sentenceY - h/2; d < -2 || d > 2 {
		t.Errorf("expected sentence row %d to be within ±2 of vertical center %d", sentenceY, h/2)
	}

	for y := 0; y < h; y++ {
		if strings.Contains(string(r.grid[y]), "(Esc for menu)") {
			t.Error("expected the old '(Esc for menu)' hint text to be gone")
		}
	}

	// Result screen: the result line should be centered, not left-aligned at x=1.
	gs.IsFinished = true
	gs.WPM = 61.2
	gs.Accuracy = 97.5
	gs.FinalDurS = 12

	r2 := newGridRenderer(w, h)
	theme.UpdateScreen(r2, gs)

	resultY := -1
	for y := 0; y < h; y++ {
		if strings.Contains(string(r2.grid[y]), "wpm") {
			resultY = y
			break
		}
	}
	if resultY == -1 {
		t.Fatal("expected to find the result text on the finished screen")
	}
	line := string(r2.grid[resultY])
	firstNonSpace := strings.IndexFunc(line, func(r rune) bool { return r != ' ' })
	if firstNonSpace <= 1 {
		t.Errorf("expected the result line to be centered (not left-aligned at x<=1), got leading content at x=%d", firstNonSpace)
	}
}
