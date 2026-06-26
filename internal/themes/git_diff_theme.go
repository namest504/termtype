package themes

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

func init() {
	Themes["diff"] = &DiffTheme{}
}

// DiffTheme mimics a git diff UI.
type DiffTheme struct{}

type DiffThemeState struct {
	ContextLines []string
}

var fakeCode = []string{
	" func main() {",
	" ",
	" ",
	" ",
	" }",
}

func (t *DiffTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()

	state := &DiffThemeState{}
	state.ContextLines = make([]string, 5)
	copy(state.ContextLines, fakeCode)
	gs.CustomState = state
}

func (t *DiffTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	state, ok := gs.CustomState.(*DiffThemeState)
	if !ok {
		return
	}
	renderer.Clear()
	w, _ := renderer.Size()

	// Draw context lines
	renderer.DrawText(0, 0, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "diff --git a/main.go b/main.go")
	renderer.DrawText(0, 1, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "--- a/main.go")
	renderer.DrawText(0, 2, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "+++ b/main.go")
	renderer.DrawText(0, 3, tcell.StyleDefault.Foreground(tcell.ColorBlue), "@@ -1,5 +1,5 @@")

	y := 4
	for i, line := range state.ContextLines {
		if i == 2 { // where the sentence goes (may span several wrapped rows)
			y = t.drawTarget(renderer, gs, y, w)
		} else {
			renderer.DrawText(0, y, tcell.StyleDefault, " "+line)
			y++
		}
	}

	if gs.IsFinished {
		renderer.HideCursor()
		renderer.DrawText(0, y+1, tcell.StyleDefault, ui.Truncate(ui.ResultText(gs), w))
	}

	renderer.Show()
}

// drawTarget renders the target as one or more "+ " diff lines, wrapping to the
// terminal width and coloring each typed rune. It returns the next free row and
// positions the cursor (unless the round is finished). Glyph placement advances
// by display width so wide runes (e.g. Hangul) align with the cursor.
func (t *DiffTheme) drawTarget(renderer domain.Renderer, gs *domain.GameState, startY, w int) int {
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorDarkRed)

	inputRunes := []rune(gs.UserInput)
	wrapWidth := w - 3 // "+ " prefix plus a cell of right padding
	if wrapWidth < 1 {
		wrapWidth = 1
	}
	lines := ui.WrapText(gs.TargetSentence, wrapWidth)

	offset := 0
	cursorX, cursorY := 2, startY
	foundCursor := false
	y := startY
	for _, line := range lines {
		lineRunes := []rune(line)
		renderer.DrawText(0, y, green, "+ ")
		col := 2
		for ci, r := range lineRunes {
			idx := offset + ci
			style := green
			if idx < len(inputRunes) && inputRunes[idx] != r {
				style = red
			}
			renderer.SetContent(col, y, r, style)
			col += runewidth.RuneWidth(r)
		}
		if !foundCursor && len(inputRunes) >= offset && len(inputRunes) < offset+len(lineRunes) {
			rel := len(inputRunes) - offset
			cursorX = 2 + runewidth.StringWidth(string(lineRunes[:rel]))
			cursorY = y
			foundCursor = true
		}
		offset += len(lineRunes)
		y++
	}
	if !foundCursor && len(lines) > 0 {
		last := []rune(lines[len(lines)-1])
		cursorX = 2 + runewidth.StringWidth(string(last))
		cursorY = startY + len(lines) - 1
	}
	if !gs.IsFinished {
		renderer.ShowCursor(cursorX, cursorY)
	}
	return y
}

func (t *DiffTheme) OnTick(gs *domain.GameState) {}
