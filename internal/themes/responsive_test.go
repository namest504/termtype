package themes

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"termtype/internal/domain"
)

// mockRenderer is a screenless renderer implementing domain.Renderer (only reports size).
type mockRenderer struct{ w, h int }

func (m *mockRenderer) DrawText(x, y int, style tcell.Style, text string) {}
func (m *mockRenderer) DrawRune(x, y int, r rune, style tcell.Style) int {
	return runewidth.RuneWidth(r)
}
func (m *mockRenderer) Clear()                                         {}
func (m *mockRenderer) Show()                                          {}
func (m *mockRenderer) Size() (int, int)                               { return m.w, m.h }
func (m *mockRenderer) SetContent(x, y int, r rune, style tcell.Style) {}
func (m *mockRenderer) HideCursor()                                    {}
func (m *mockRenderer) ShowCursor(x, y int)                            {}

// Every theme must render without panicking at any terminal size (including very small ones).
// Also guards regressions like the log theme crashing at height<=4 and hex resize issues.
func TestThemes_NoPanicAcrossSizes(t *testing.T) {
	sizes := []struct{ w, h int }{
		{1, 1}, {2, 2}, {3, 3}, {10, 4}, {40, 1}, {40, 2},
		{40, 4}, {41, 5}, {80, 24}, {120, 40}, {200, 60},
	}
	inputs := []string{"", "Th", "partial input that is reasonably long", ""}

	for name, theme := range Themes {
		for _, sz := range sizes {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("theme %q panicked at %dx%d: %v", name, sz.w, sz.h, r)
					}
				}()
				r := &mockRenderer{w: sz.w, h: sz.h}
				gs := &domain.GameState{Sentences: domain.Sentences}
				theme.ResetState(gs)
				for _, in := range inputs {
					gs.UserInput = in
					theme.OnTick(gs)
					theme.UpdateScreen(r, gs)
				}
				// Also render the finished state
				gs.IsFinished = true
				gs.Wpm = 61.2
				gs.Accuracy = 97.5
				theme.UpdateScreen(r, gs)
			}()
		}
	}
}
