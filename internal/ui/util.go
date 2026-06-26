package ui

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

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
