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

	// Draw context lines
	renderer.DrawText(0, 0, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "diff --git a/main.go b/main.go")
	renderer.DrawText(0, 1, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "--- a/main.go")
	renderer.DrawText(0, 2, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "+++ b/main.go")
	renderer.DrawText(0, 3, tcell.StyleDefault.Foreground(tcell.ColorBlue), "@@ -1,5 +1,5 @@")

	y := 4
	for i, line := range state.ContextLines {
		if i == 2 { // where the sentence goes
			plusStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
			renderer.DrawText(0, y, plusStyle, "+ "+gs.TargetSentence)

			// User input feedback
			targetRunes := []rune(gs.TargetSentence)
			inputRunes := []rune(gs.UserInput)

			// Advance by display width so wide (e.g. Hangul) runes line up with
			// the "+ " prefix and the cursor, which also use display width.
			col := 2
			for i := 0; i < len(targetRunes); i++ {
				style := tcell.StyleDefault.Foreground(tcell.ColorGreen)
				if i < len(inputRunes) && inputRunes[i] != targetRunes[i] {
					// Mark only typed-but-wrong characters in red
					style = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorDarkRed)
				}
				// i >= len(inputRunes): not yet typed → keep default green
				renderer.SetContent(col, y, targetRunes[i], style)
				col += runewidth.RuneWidth(targetRunes[i])
			}
		} else {
			renderer.DrawText(0, y, tcell.StyleDefault, " "+line)
		}
		y++
	}

	if gs.IsFinished {
		renderer.HideCursor()
		resultText := ui.ResultText(gs)
		renderer.DrawText(0, y+2, tcell.StyleDefault, resultText)
	} else {
		cursorX := 2 + runewidth.StringWidth(gs.UserInput)
		renderer.ShowCursor(cursorX, 6)
	}

	renderer.Show()
}

func (t *DiffTheme) OnTick(gs *domain.GameState) {}
