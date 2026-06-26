package domain

import "github.com/gdamore/tcell/v2"

// Renderer interface abstracts screen drawing functionality.
type Renderer interface {
	DrawText(x, y int, style tcell.Style, text string)
	DrawRune(x, y int, runeVal rune, style tcell.Style) int
	Clear()
	Show()
	Size() (int, int)
	SetContent(x, y int, runeVal rune, style tcell.Style)
	HideCursor()
	ShowCursor(x, y int)
}

// Theme interface defines the methods every theme must implement.
type Theme interface {
	// ResetState resets the game state for a new round.
	ResetState(*GameState)
	// UpdateScreen draws the current game state to the screen.
	UpdateScreen(Renderer, *GameState)
	// OnTick is called when real-time updates are needed (e.g. animation).
	OnTick(*GameState)
}
