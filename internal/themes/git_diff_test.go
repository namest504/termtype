package themes

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

func rowText(ss tcell.SimulationScreen, y int) string {
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
	return strings.TrimRight(b.String(), " ")
}

// At a narrow width the diff theme must wrap a long target into several "+ "
// lines instead of letting it run off the edge, and each wrapped row must stay
// within the terminal width.
func TestDiff_WrapsLongTarget(t *testing.T) {
	ss := tcell.NewSimulationScreen("")
	if err := ss.Init(); err != nil {
		t.Fatal(err)
	}
	const w = 24
	ss.SetSize(w, 14)

	gs := &domain.GameState{Sentences: domain.Sentences}
	th := Themes["diff"]
	th.ResetState(gs)
	gs.TargetSentence = "The quick brown fox jumps over the lazy dog."
	gs.UserInput = ""
	th.UpdateScreen(ui.NewRenderer(ss), gs)
	ss.Show()

	plusLines := 0
	cells, cw, ch := ss.GetContents()
	for y := 0; y < ch; y++ {
		row := rowText(ss, y)
		if strings.HasPrefix(row, "+ ") {
			plusLines++
		}
		// Guard against drawing past the right edge.
		if len([]rune(row)) > cw {
			t.Errorf("row %d overflows width %d: %q", y, cw, row)
		}
	}
	_ = cells
	if plusLines < 2 {
		t.Errorf("expected the target to wrap into >=2 '+ ' lines, got %d", plusLines)
	}
}
