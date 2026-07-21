package ui

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// WindowStart returns the first visible line of a visible-line-tall window
// over total lines, keeping the cursor line near the middle row where
// possible and pinning the window to the edges of the text.
func WindowStart(total, cursor, visible int) int {
	start := cursor - (visible-1)/2
	if start > total-visible {
		start = total - visible
	}
	if start < 0 {
		start = 0
	}
	return start
}

// WrapText splits the given text into multiple lines fitting the specified width.
func WrapText(text string, width int) []string {
	var lines []string
	if width <= 0 {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	currentLine := ""
	for _, word := range words {
		if currentLine == "" {
			// Start a new line: keep a single word even if it exceeds the width (avoids empty lines)
			currentLine = word
		} else if runewidth.StringWidth(currentLine+" "+word) <= width {
			currentLine += " " + word
		} else {
			// Keep the inter-word space at the line end to preserve input offset alignment
			lines = append(lines, currentLine+" ")
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

// Truncate cuts a string to within the display width. If it exceeds the width, it returns a truncated copy.
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= width {
		return s
	}
	return runewidth.Truncate(s, width, "")
}
