package app

import (
	"fmt"
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

// Create a new game. timeLimit > 0 enables time-attack mode. sentences is the
// pool the chosen theme draws targets from; an empty pool falls back to the
// default English set.
func NewGame(s tcell.Screen, theme domain.Theme, timeLimit time.Duration, sentences []string) (*Game, error) {
	if len(sentences) == 0 {
		sentences = domain.Sentences
	}
	state := &domain.GameState{Sentences: sentences, TimeLimit: timeLimit}
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

	g.render()

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
					g.render()
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
					g.render()
				}
			}
		case <-ticker.C:
			w, _ := g.screen.Size()
			if w >= 40 && !g.state.IsFinished {
				if !g.state.Paused {
					if g.state.TimeLimit > 0 && g.state.TimerStarted && g.state.Remaining() <= 0 {
						g.state.TimedOut = true
						g.state.Finalize()
					} else {
						g.theme.OnTick(g.state)
					}
				}
				g.render()
			}
		}
	}
}

// render draws the current theme plus the mode overlays (time-attack countdown
// and pause banner) on top.
func (g *Game) render() {
	g.theme.UpdateScreen(g.renderer, g.state)
	g.drawOverlay()
	g.screen.Show()
}

func (g *Game) drawOverlay() {
	w, h := g.screen.Size()

	if g.state.TimeLimit > 0 {
		var label string
		var style tcell.Style
		switch {
		case g.state.IsFinished && g.state.TimedOut:
			label = " ⏱ TIME UP "
			style = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorMaroon)
		case g.state.IsFinished:
			label = " ⏱ DONE "
			style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen)
		default:
			secs := int(g.state.Remaining().Seconds() + 0.999)
			label = fmt.Sprintf(" ⏱ %d:%02d ", secs/60, secs%60)
			if secs <= 5 {
				style = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed)
			} else {
				style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
			}
		}
		x := w - len([]rune(label))
		if x < 0 {
			x = 0
		}
		g.renderer.DrawText(x, 0, style, label)
	}

	if g.state.Paused {
		msg := " ⏸  PAUSED — Ctrl-P to resume "
		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
		x := (w - len([]rune(msg))) / 2
		if x < 0 {
			x = 0
		}
		g.renderer.DrawText(x, h/2, style, msg)
	}
}

// Handle key events
func (g *Game) handleKeyEvent(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		g.screen.Fini()
		os.Exit(0)
	}

	// Ctrl-P toggles pause; while paused, all other input is ignored.
	if ev.Key() == tcell.KeyCtrlP {
		g.state.TogglePause()
		return
	}
	if g.state.Paused {
		return
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

	// Finalize once the whole sentence has been typed (rune count, not bytes).
	if len([]rune(g.state.UserInput)) >= len([]rune(g.state.TargetSentence)) {
		g.state.Finalize()
	}
}
