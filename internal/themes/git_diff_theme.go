package themes

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["diff"] = &DiffTheme{}
}

// DiffTheme mimics a git diff UI: the target is the added hunk, typed over
// a plausible bit of code with a couple of removed lines above it.
type DiffTheme struct{}

// diffLine is one non-target row of the fake hunk.
type diffLine struct {
	mark rune // ' ' unchanged context, '-' removed
	text string
}

var diffLead = []diffLine{
	{' ', "func handleRequest(w http.ResponseWriter, r *http.Request) {"},
	{' ', "    defer r.Body.Close()"},
	{'-', "    data, _ := io.ReadAll(r.Body)"},
	{'-', "    process(data)"},
}

var diffTail = []diffLine{
	{' ', "    w.WriteHeader(http.StatusOK)"},
	{' ', "}"},
}

func (t *DiffTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
}

func (t *DiffTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	renderer.Clear()
	w, h := renderer.Size()

	dim := tcell.StyleDefault.Foreground(tcell.ColorDimGray)
	renderer.DrawText(0, 0, dim, "diff --git a/server/handler.go b/server/handler.go")
	renderer.DrawText(0, 1, dim, "--- a/server/handler.go")
	renderer.DrawText(0, 2, dim, "+++ b/server/handler.go")
	renderer.DrawText(0, 3, tcell.StyleDefault.Foreground(tcell.ColorBlue), "@@ -12,8 +12,9 @@")

	y := t.drawContext(renderer, diffLead, 4, w)
	y = t.drawTarget(renderer, gs, y, w, h)
	y = t.drawContext(renderer, diffTail, y, w)

	if gs.IsFinished {
		renderer.HideCursor()
		renderer.DrawText(0, y+1, tcell.StyleDefault, ui.Truncate(ui.ResultText(gs), w))
	}

	renderer.Show()
}

// drawContext renders unchanged and removed hunk lines and returns the next
// free row.
func (t *DiffTheme) drawContext(renderer domain.Renderer, lines []diffLine, y, w int) int {
	removed := tcell.StyleDefault.Foreground(tcell.ColorRed)
	for _, ln := range lines {
		style, mark := tcell.StyleDefault, " "
		if ln.mark == '-' {
			style, mark = removed, "-"
		}
		renderer.DrawText(0, y, style, ui.Truncate(mark+" "+ln.text, w))
		y++
	}
	return y
}

// drawTarget renders the target as one or more "+ " diff lines, wrapping to the
// terminal width and coloring each typed rune. Long targets (the words stream)
// scroll inside a window that follows the cursor. It returns the next free row
// and positions the cursor (unless the round is finished). Glyph placement
// advances by display width so wide runes (e.g. Hangul) align with the cursor.
func (t *DiffTheme) drawTarget(renderer domain.Renderer, gs *domain.GameState, startY, w, h int) int {
	green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorDarkRed)

	inputRunes := []rune(gs.UserInput)
	wrapWidth := w - 3 // "+ " prefix plus a cell of right padding
	if wrapWidth < 1 {
		wrapWidth = 1
	}
	allLines := ui.WrapText(gs.TargetSentence, wrapWidth)

	// Window the "+" hunk so it never runs off the screen.
	visible := h - startY - 4 // leave room for the trailing context and result
	if visible > 8 {
		visible = 8
	}
	if visible < 1 {
		visible = 1
	}
	winStart := 0
	lines := allLines
	if len(allLines) > visible {
		winStart = ui.WindowStart(len(allLines), ui.LineOfRune(allLines, len(inputRunes)), visible)
		lines = allLines[winStart : winStart+visible]
	}

	offset := 0
	for i := 0; i < winStart; i++ {
		offset += len([]rune(allLines[i]))
	}

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
		// The cursor sits past the window's last rune (end of the target,
		// or the input has consumed the whole visible window).
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
