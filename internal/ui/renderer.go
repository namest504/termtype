package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Renderer is a helper struct responsible for drawing to the screen.
type Renderer struct {
	screen tcell.Screen
}

// NewRenderer creates a new Renderer instance.
func NewRenderer(s tcell.Screen) *Renderer {
	return &Renderer{screen: s}
}

// DrawText draws text on the screen.
func (r *Renderer) DrawText(x, y int, style tcell.Style, text string) {
	currentX := x
	for _, runeVal := range []rune(text) {
		width := r.DrawRune(currentX, y, runeVal, style)
		currentX += width
	}
}

// DrawRune draws a single character on the screen and returns its width.
func (r *Renderer) DrawRune(x, y int, runeVal rune, style tcell.Style) int {
	r.screen.SetContent(x, y, runeVal, nil, style)
	return runewidth.RuneWidth(runeVal)
}

// Clear clears the screen.
func (r *Renderer) Clear() {
	r.screen.Clear()
}

// ShowCursor shows the cursor.
func (r *Renderer) ShowCursor(x, y int) {
	r.screen.ShowCursor(x, y)
}

// HideCursor hides the cursor.
func (r *Renderer) HideCursor() {
	r.screen.HideCursor()
}

// Show updates the screen.
func (r *Renderer) Show() {
	r.screen.Show()
}

// Size returns the screen size.
func (r *Renderer) Size() (int, int) {
	return r.screen.Size()
}

// SetContent sets the character at a specific position on the screen.
func (r *Renderer) SetContent(x, y int, runeVal rune, style tcell.Style) {
	r.screen.SetContent(x, y, runeVal, nil, style)
}
