package app

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// wrapText는 주어진 텍스트를 지정된 너비에 맞춰 여러 줄로 나눕니다.
func wrapText(text string, width int) []string {
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
		if runewidth.StringWidth(currentLine+" "+word) <= width {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}
