package app

import (
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

// Struct that manages the entire game
type Game struct {
	screen   tcell.Screen
	renderer *ui.Renderer
	state    *domain.GameState
	theme    domain.Theme
}

// Create a new game
func NewGame(s tcell.Screen, theme domain.Theme) (*Game, error) {
	state := &domain.GameState{Sentences: domain.Sentences}
	theme.ResetState(state)

	return &Game{screen: s, renderer: ui.NewRenderer(s), state: state, theme: theme}, nil
}

// Run the game (with a real-time Ticker)
func (g *Game) Run() {
	ticker := time.NewTicker(1 * time.Second) // Tick every second
	defer ticker.Stop()

	eventChan := make(chan tcell.Event)
	go func() {
		for {
			eventChan <- g.screen.PollEvent()
		}
	}()

	g.theme.UpdateScreen(g.renderer, g.state)

	for {
		select {
		case ev := <-eventChan:
			switch ev := ev.(type) {
			case *tcell.EventPaste:
				// Pasting is intentionally ignored — type the text yourself.
			case *tcell.EventResize:
				g.screen.Sync()
				w, _ := g.screen.Size()
				if w < 40 {
					g.screen.Clear()
					g.renderer.DrawText(1, 1, tcell.StyleDefault.Foreground(tcell.ColorRed), "Terminal too small (min width: 40)")
					g.screen.Show()
				} else {
					g.theme.UpdateScreen(g.renderer, g.state)
				}
			case *tcell.EventKey:
				w, _ := g.screen.Size()
				if w < 40 {
					// Do not process keys if screen is too small, except quit keys
					if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
						g.screen.Fini()
						os.Exit(0)
					}
				} else {
					g.handleKeyEvent(ev)
					g.theme.UpdateScreen(g.renderer, g.state)
				}
			}
		case <-ticker.C:
			w, _ := g.screen.Size()
			if w >= 40 {
				if !g.state.IsFinished {
					g.theme.OnTick(g.state)
					g.theme.UpdateScreen(g.renderer, g.state)
				}
			}
		}
	}
}

// Handle key events
func (g *Game) handleKeyEvent(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		g.screen.Fini()
		os.Exit(0)
	}

	if g.state.IsFinished {
		if ev.Key() == tcell.KeyEnter {
			g.theme.ResetState(g.state)
		}
		return
	}

	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(g.state.UserInput) > 0 {
			runes := []rune(g.state.UserInput)
			g.state.UserInput = string(runes[:len(runes)-1])
		}
	case tcell.KeyRune:
		if !g.state.TimerStarted {
			g.state.StartTime = time.Now()
			g.state.TimerStarted = true
		}
		g.state.UserInput += string(ev.Rune())
	}

	// Check typing completion (compare by runes, not bytes)
	inputRunes := []rune(g.state.UserInput)
	targetRunes := []rune(g.state.TargetSentence)
	if len(inputRunes) >= len(targetRunes) {
		g.state.IsFinished = true
		duration := time.Since(g.state.StartTime).Minutes()

		// Trim input beyond the target length at rune boundaries
		if len(inputRunes) > len(targetRunes) {
			inputRunes = inputRunes[:len(targetRunes)]
			g.state.UserInput = string(inputRunes)
		}

		if duration > 0 {
			g.state.Wpm = (float64(len(inputRunes)) / 5.0) / duration
		}

		correctChars := 0
		for i, r := range targetRunes {
			if i < len(inputRunes) && inputRunes[i] == r {
				correctChars++
			}
		}
		if len(targetRunes) > 0 {
			g.state.Accuracy = (float64(correctChars) / float64(len(targetRunes))) * 100
		}
	}
}
