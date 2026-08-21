package themes

import (
	"testing"

	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

// TestThemesHonorASCIIContract renders every registered theme, in both its
// typing and finished states, with ui.SetASCII(true) active, and asserts no
// rune above the ASCII range (127) ever lands on the grid. This is the
// enforcement test for the ASCII contract: any theme that draws a decorative
// glyph or em dash without checking ui.IsASCII() should fail it.
func TestThemesHonorASCIIContract(t *testing.T) {
	prevASCII := ui.IsASCII()
	ui.SetASCII(true)
	t.Cleanup(func() { ui.SetASCII(prevASCII) })

	sentences := []string{"the quick brown fox jumps over the lazy dog"}

	for name, theme := range Themes {
		for _, size := range []struct{ w, h int }{{100, 30}, {40, 16}} {
			t.Run(name, func(t *testing.T) {
				gs := &domain.GameState{Sentences: sentences}
				theme.ResetState(gs)
				gs.TargetSentence = sentences[0]

				r := newGridRenderer(size.w, size.h)

				// Typing state, partway through.
				gs.UserInput = "the quick brown"
				gs.IsFinished = false
				theme.UpdateScreen(r, gs)
				assertGridIsASCII(t, r, name, "typing")

				// Finished state.
				gs.UserInput = sentences[0]
				gs.IsFinished = true
				gs.WPM = 61.2
				gs.Accuracy = 97.5
				gs.FinalDurS = 12
				r2 := newGridRenderer(size.w, size.h)
				theme.UpdateScreen(r2, gs)
				assertGridIsASCII(t, r2, name, "finished")
			})
		}
	}
}

func assertGridIsASCII(t *testing.T, r *gridRenderer, theme, state string) {
	t.Helper()
	for y := 0; y < r.h; y++ {
		for x := 0; x < r.w; x++ {
			if ch := r.grid[y][x]; ch > 127 {
				t.Errorf("theme %q (%s state): non-ASCII rune %q at (%d,%d) with ui.SetASCII(true)", theme, state, ch, x, y)
			}
		}
	}
}
