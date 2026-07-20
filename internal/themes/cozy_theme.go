package themes

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["cozy"] = &CozyTheme{}
}

// --- Cozy Theme --- //

// CozyTheme is a warm, quiet typing screen: a muted dark background, a
// three-line window over the target text centered on screen, typed text in
// cream, mistakes in red, upcoming text dimmed, and a small timer above.
type CozyTheme struct{}

var (
	cozyBg     = tcell.NewRGBColor(31, 29, 27)    // warm near-black
	cozyCream  = tcell.NewRGBColor(213, 196, 161) // typed, correct
	cozyDim    = tcell.NewRGBColor(110, 103, 92)  // not yet typed
	cozyRed    = tcell.NewRGBColor(204, 102, 102) // typed, wrong
	cozyAccent = tcell.NewRGBColor(184, 187, 38)  // timer and highlights
)

// cozyVisibleLines is the height of the window over the wrapped target.
const cozyVisibleLines = 3

func (t *CozyTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
}

func (t *CozyTheme) OnTick(gs *domain.GameState) {}

func (t *CozyTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	w, h := renderer.Size()
	bg := tcell.StyleDefault.Background(cozyBg)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			renderer.SetContent(x, y, ' ', bg)
		}
	}

	if gs.IsFinished {
		t.drawResult(renderer, gs, w, h)
	} else {
		t.drawTyping(renderer, gs, w, h)
	}
	renderer.Show()
}

// cozyColumnWidth returns the text column width for a terminal width: a
// comfortable reading measure that shrinks with the terminal.
func cozyColumnWidth(w int) int {
	col := w - 6
	if col > 60 {
		col = 60
	}
	if col < 1 {
		col = 1
	}
	return col
}

// cozyWindowStart returns the first visible line of the window so the cursor
// line stays on the middle row where possible.
func cozyWindowStart(totalLines, cursorLine int) int {
	start := cursorLine - 1
	if start > totalLines-cozyVisibleLines {
		start = totalLines - cozyVisibleLines
	}
	if start < 0 {
		start = 0
	}
	return start
}

// cursorLineOf returns the wrapped line the next rune to type falls on.
func cursorLineOf(wrapped []string, typedRunes int) int {
	offset := 0
	for i, line := range wrapped {
		n := len([]rune(line))
		if typedRunes < offset+n {
			return i
		}
		offset += n
	}
	return len(wrapped) - 1
}

func (t *CozyTheme) drawTyping(renderer domain.Renderer, gs *domain.GameState, w, h int) {
	col := cozyColumnWidth(w)
	wrapped := ui.WrapText(gs.TargetSentence, col)
	input := []rune(gs.UserInput)
	target := []rune(gs.TargetSentence)

	startX := (w - col) / 2
	textTop := h/2 - 1
	if textTop < 2 {
		textTop = 2
	}

	// Timer: the countdown in time attack, elapsed seconds otherwise.
	secs := int(gs.Elapsed().Seconds())
	if gs.TimeLimit > 0 {
		secs = int(gs.Remaining().Seconds() + 0.999)
	}
	renderer.DrawText(startX, textTop-2, tcell.StyleDefault.Foreground(cozyAccent).Background(cozyBg),
		fmt.Sprintf("%ds", secs))

	cursorLine := cursorLineOf(wrapped, len(input))
	winStart := cozyWindowStart(len(wrapped), cursorLine)

	// Rune offset of the first visible line.
	offset := 0
	for i := 0; i < winStart; i++ {
		offset += len([]rune(wrapped[i]))
	}

	for row := 0; row < cozyVisibleLines && winStart+row < len(wrapped); row++ {
		lineRunes := []rune(wrapped[winStart+row])
		x := startX
		for i, r := range lineRunes {
			idx := offset + i
			style := tcell.StyleDefault.Foreground(cozyDim).Background(cozyBg)
			switch {
			case idx < len(input) && idx < len(target) && input[idx] == target[idx]:
				style = tcell.StyleDefault.Foreground(cozyCream).Background(cozyBg)
			case idx < len(input):
				style = tcell.StyleDefault.Foreground(cozyRed).Background(cozyBg)
			case idx == len(input):
				// Block cursor on the next rune to type.
				style = tcell.StyleDefault.Foreground(cozyBg).Background(cozyCream)
			}
			x += renderer.DrawRune(x, textTop+row, r, style)
		}
		offset += len(lineRunes)
	}
	renderer.HideCursor()
}

func (t *CozyTheme) drawResult(renderer domain.Renderer, gs *domain.GameState, w, h int) {
	renderer.HideCursor()

	drawCentered := func(y int, style tcell.Style, text string) {
		x := (w - runewidth.StringWidth(text)) / 2
		if x < 0 {
			x = 0
		}
		renderer.DrawText(x, y, style, ui.Truncate(text, w))
	}

	// The wpm graph slots between the title and the summary when the round
	// has samples and the terminal has the rows for it.
	chartH := 0
	if len(gs.WPMSamples) >= 2 && h >= 12 {
		chartH = h - 10
		if chartH > 8 {
			chartH = 8
		}
	}

	top := (h - (6 + chartH)) / 2
	if top < 1 {
		top = 1
	}

	drawCentered(top, tcell.StyleDefault.Foreground(cozyAccent).Background(cozyBg),
		fmt.Sprintf("wpm: %.0f", gs.Wpm))
	y := top + 2
	if chartH > 0 {
		chartW := w - 10
		if chartW > 52 {
			chartW = 52
		}
		ui.DrawLineChart(renderer, (w-chartW)/2, y, chartW, chartH, gs.WPMSamples,
			tcell.StyleDefault.Foreground(cozyDim).Background(cozyBg),
			tcell.StyleDefault.Foreground(cozyCream).Background(cozyBg))
		y += chartH + 1
	}
	drawCentered(y, tcell.StyleDefault.Foreground(cozyCream).Background(cozyBg),
		fmt.Sprintf("accuracy: %.1f  raw: %.0f  cpm: %.0f  time: %.0fs",
			gs.Accuracy, gs.FinalRawWPM, gs.FinalCPM, gs.FinalDurS))
	drawCentered(y+2, tcell.StyleDefault.Foreground(cozyDim).Background(cozyBg),
		"enter next  esc quit")
}
