package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// --- Helper Functions --- //

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	for _, r := range []rune(text) {
		s.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}

func getStyleForLogLevel(level string) tcell.Style {
	style := tcell.StyleDefault.Foreground(tcell.ColorGray)
	switch level {
	case "INFO":
		style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	case "WARN":
		style = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	case "ERROR":
		style = tcell.StyleDefault.Foreground(tcell.ColorRed)
	case "DEBUG":
		style = tcell.StyleDefault.Foreground(tcell.ColorBlue)
	}
	return style
}
