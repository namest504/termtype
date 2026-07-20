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
}

// TypingRenderer holds the shared logic for drawing the typing area and cursor.
type TypingRenderer struct{}

// Draw renders the target sentence, user input, and cursor.
func (tr *TypingRenderer) Draw(renderer domain.Renderer, gs *domain.GameState, opts TypingRendererOptions) {
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
	inputOffset := 0

	// Draw the text.
	for lineIdx, line := range wrappedTarget {
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
			style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
			if opts.CenterText {
				style = style.Background(tcell.ColorBlack)
			} else {
				style = tcell.StyleDefault.Foreground(tcell.ColorGray)
			}

			if currentInputIdx < len(inputRunes) {
				if inputRunes[currentInputIdx] == r {
					if opts.CenterText {
						style = tcell.StyleDefault.Foreground(tcell.ColorLawnGreen).Background(tcell.ColorBlack)
					} else {
						style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
					}
				} else {
					if opts.CenterText {
						style = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack)
					} else {
						style = tcell.StyleDefault.Foreground(tcell.ColorRed)
					}
				}
			}

			width := renderer.DrawRune(currentX, opts.StartY+lineIdx, r, style)
			currentX += width
		}
		inputOffset += len(lineRunes)
	}

	// Draw the cursor.
	tr.drawCursor(renderer, wrappedTarget, inputRunes, opts)
}

func (tr *TypingRenderer) drawCursor(renderer domain.Renderer, wrappedTarget []string, inputRunes []rune, opts TypingRendererOptions) {
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

	renderer.ShowCursor(cursorX, opts.StartY+cursorLineIdx)
}
