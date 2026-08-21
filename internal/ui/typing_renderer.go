package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
)

// TypingRendererOptions holds rendering options for the typing area.
type TypingRendererOptions struct {
	// StartY is the starting Y coordinate for drawing.
	StartY int
	// Width is the available width for text wrapping.
	Width int
	// PrefixWidth is the prefix width (e.g. log timestamp).
	PrefixWidth int
	// CenterText controls whether the text is center-aligned.
	CenterText bool
	// MaxLines caps how many wrapped lines are on screen at once; 0 draws
	// them all. When the target wraps to more lines (e.g. a words stream),
	// a MaxLines-tall window follows the cursor.
	MaxLines int
}

// TypingRenderer holds the shared logic for drawing the typing area and cursor.
type TypingRenderer struct{}

// LineOfRune returns the wrapped line the rune at idx falls on.
func LineOfRune(wrapped []string, idx int) int {
	offset := 0
	for i, line := range wrapped {
		n := len([]rune(line))
		if idx < offset+n {
			return i
		}
		offset += n
	}
	if len(wrapped) == 0 {
		return 0
	}
	return len(wrapped) - 1
}

// Draw renders the target sentence, user input, and cursor. It returns the
// number of text rows drawn, so callers can lay out content underneath.
func (tr *TypingRenderer) Draw(renderer domain.Renderer, gs *domain.GameState, opts TypingRendererOptions) int {
	// Calculate the available text width.
	availableWidth := opts.Width
	if !opts.CenterText {
		availableWidth -= opts.PrefixWidth
	}

	// Wrap the text.
	// When not centered, reserve right padding (2~3 cells).
	if !opts.CenterText {
		availableWidth -= 3
	}
	wrappedTarget := WrapText(gs.TargetSentence, availableWidth)
	inputRunes := []rune(gs.UserInput)

	// Window the wrapped lines around the cursor when they exceed MaxLines.
	winStart := 0
	lines := wrappedTarget
	if opts.MaxLines > 0 && len(wrappedTarget) > opts.MaxLines {
		winStart = WindowStart(len(wrappedTarget), LineOfRune(wrappedTarget, len(inputRunes)), opts.MaxLines)
		lines = wrappedTarget[winStart : winStart+opts.MaxLines]
	}

	inputOffset := 0
	for i := 0; i < winStart; i++ {
		inputOffset += len([]rune(wrappedTarget[i]))
	}

	// Draw the text.
	for lineIdx, line := range lines {
		lineRunes := []rune(line)

		// Calculate the starting X coordinate for the current line.
		startX := 1 + opts.PrefixWidth
		if opts.CenterText {
			startX = (opts.Width - runewidth.StringWidth(line)) / 2
		}

		currentX := startX
		for charIdx, r := range lineRunes {
			currentInputIdx := inputOffset + charIdx

			// Determine the style.
			style := tcell.StyleDefault
			if !opts.CenterText {
				style = tcell.StyleDefault.Foreground(tcell.ColorGray)
			}

			if currentInputIdx < len(inputRunes) {
				if inputRunes[currentInputIdx] == r {
					if opts.CenterText {
						style = tcell.StyleDefault.Foreground(tcell.ColorLawnGreen)
					} else {
						style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
					}
				} else {
					style = tcell.StyleDefault.Foreground(tcell.ColorRed)
				}
			}

			width := renderer.DrawRune(currentX, opts.StartY+lineIdx, r, style)
			currentX += width
		}
		inputOffset += len(lineRunes)
	}

	// Draw the cursor.
	tr.drawCursor(renderer, wrappedTarget, winStart, len(lines), inputRunes, opts)
	return len(lines)
}

func (tr *TypingRenderer) drawCursor(renderer domain.Renderer, wrappedTarget []string, winStart, winLen int, inputRunes []rune, opts TypingRendererOptions) {
	cursorLineIdx := 0
	cursorX := 1 + opts.PrefixWidth
	if opts.CenterText {
		cursorX = 0
	}

	currentOffset := 0
	foundCursor := false

	for i, line := range wrappedTarget {
		lineRunes := []rune(line)
		lineLen := len(lineRunes)

		if len(inputRunes) >= currentOffset && len(inputRunes) < currentOffset+lineLen {
			cursorLineIdx = i
			cursorRelIdx := len(inputRunes) - currentOffset

			startX := 1 + opts.PrefixWidth
			if opts.CenterText {
				startX = (opts.Width - runewidth.StringWidth(line)) / 2
			}

			cursorX = startX + runewidth.StringWidth(string(lineRunes[:cursorRelIdx]))
			foundCursor = true
			break
		}
		currentOffset += lineLen
	}

	if !foundCursor && len(wrappedTarget) > 0 {
		cursorLineIdx = len(wrappedTarget) - 1
		lastLine := wrappedTarget[len(wrappedTarget)-1]

		startX := 1 + opts.PrefixWidth
		if opts.CenterText {
			startX = (opts.Width - runewidth.StringWidth(lastLine)) / 2
		}

		cursorX = startX + runewidth.StringWidth(lastLine)
	}

	// Translate the cursor line into the visible window.
	row := cursorLineIdx - winStart
	if row < 0 {
		row = 0
	}
	if winLen > 0 && row >= winLen {
		row = winLen - 1
	}
	renderer.ShowCursor(cursorX, opts.StartY+row)
}
