package themes

import (
	"strings"
	"testing"

	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

// TestMatrixTypingClearing verifies the sentence band sits in a readable
// clearing: the row just above the wrapped sentence and the row just below
// it must be entirely blank, even though the rain is animating underneath.
func TestMatrixTypingClearing(t *testing.T) {
	theme := &MatrixTheme{}
	w, h := 80, 24

	gs := &domain.GameState{Sentences: []string{"hello world"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hello world"

	r := newGridRenderer(w, h)
	// Rain drops initialize on the first UpdateScreen call, so the screen
	// size must already be set before that call. Then mirror the real game
	// loop's Update -> Tick -> Update order.
	theme.UpdateScreen(r, gs)
	theme.OnTick(gs)
	theme.UpdateScreen(r, gs)

	startY := h/2 - 2
	rows := len(ui.WrapText(gs.TargetSentence, w-4))
	if cap := matrixMaxLines(h, startY); rows > cap {
		rows = cap
	}

	top := startY - 1
	bottom := startY + rows
	for _, y := range []int{top, bottom} {
		if y < 0 || y >= h {
			continue
		}
		for x := 0; x < w; x++ {
			if r.grid[y][x] != ' ' {
				t.Errorf("row %d (clearing band edge) cell x=%d = %q, want blank", y, x, r.grid[y][x])
			}
		}
	}
}

// TestMatrixResultPanel verifies the finished-state panel shows "TRACE
// COMPLETE" and the result stats, and that the rain is still alive outside
// the cleared panel band.
func TestMatrixResultPanel(t *testing.T) {
	theme := &MatrixTheme{}
	w, h := 80, 24

	gs := &domain.GameState{Sentences: []string{"hello world"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hello world"
	gs.IsFinished = true
	gs.WPM = 61.2
	gs.Accuracy = 97.5
	gs.FinalDurS = 12

	r := newGridRenderer(w, h)
	theme.UpdateScreen(r, gs)
	theme.OnTick(gs)
	theme.UpdateScreen(r, gs)

	startY := h/2 - 2

	titleLine := string(r.grid[startY])
	if !strings.Contains(titleLine, "TRACE COMPLETE") {
		t.Errorf("row %d = %q, want it to contain %q", startY, titleLine, "TRACE COMPLETE")
	}

	statsLine := string(r.grid[startY+2])
	stats := ui.ResultText(gs)
	if !strings.Contains(statsLine, stats) {
		t.Errorf("row %d = %q, want it to contain %q", startY+2, statsLine, stats)
	}

	// Outside the panel band ([startY-1, startY+3]) the rain should still be
	// alive: at least one non-blank cell somewhere else on the screen.
	found := false
	for y := 0; y < h; y++ {
		if y >= startY-1 && y <= startY+3 {
			continue
		}
		for x := 0; x < w; x++ {
			if r.grid[y][x] != ' ' {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Errorf("expected non-blank rain cells outside the result panel band, found none")
	}
}
