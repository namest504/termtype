package app

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/store"
	"github.com/namest504/termtype/internal/ui"
)

// minWidth is the narrowest terminal the game will play in. Below this the
// typing area is too cramped to be usable, so we show a short notice instead.
const minWidth = 20

// RoundMeta labels the rounds a game produces for the history file.
type RoundMeta struct {
	Theme  string
	Lang   string // "en" | "ko" | "-" for custom text
	Source string // "builtin" | "custom"
}

// Struct that manages the entire game
type Game struct {
	screen   tcell.Screen
	renderer *ui.Renderer
	state    *domain.GameState
	theme    domain.Theme

	meta       RoundMeta
	store      *store.Store
	history    []store.Round // loaded once at startup, appended in memory
	resultLine string        // PB/recent line for the finished round ("" = hide)
	resultBest bool          // true when resultLine announces a new PB
}

// Create a new game. timeLimit > 0 enables time-attack mode. sentences is the
// pool the chosen theme draws targets from; an empty pool falls back to the
// default English set. meta labels recorded rounds; st may be nil to disable
// persistence.
func NewGame(s tcell.Screen, theme domain.Theme, timeLimit time.Duration, sentences []string, meta RoundMeta, st *store.Store) (*Game, error) {
	if len(sentences) == 0 {
		sentences = domain.Sentences
	}
	state := &domain.GameState{Sentences: sentences, TimeLimit: timeLimit}
	theme.ResetState(state)

	return &Game{screen: s, renderer: ui.NewRenderer(s), state: state, theme: theme,
		meta: meta, store: st, history: st.LoadHistory()}, nil
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
				if w < minWidth {
					g.drawTooSmall(w)
				} else {
					g.render()
				}
			case *tcell.EventKey:
				w, _ := g.screen.Size()
				if w < minWidth {
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
			if w >= minWidth && !g.state.IsFinished {
				if !g.state.Paused {
					if g.state.TimeLimit > 0 && g.state.TimerStarted && g.state.Remaining() <= 0 {
						g.state.TimedOut = true
						g.finalizeRound()
					} else {
						g.theme.OnTick(g.state)
					}
				}
				g.render()
			}
		}
	}
}

// finalizeRound ends the round, appends it to the history, and prepares the
// result line the overlay shows: a NEW BEST banner, or the current best with
// a sparkline of the last 10 rounds in the same (mode, lang, source) bucket.
func (g *Game) finalizeRound() {
	g.state.Finalize()

	r := store.Round{
		TS:     time.Now(),
		Theme:  g.meta.Theme,
		Mode:   store.ModeString(g.state.TimeLimit),
		Lang:   g.meta.Lang,
		WPM:    g.state.Wpm,
		Acc:    g.state.Accuracy,
		DurS:   g.state.Elapsed().Seconds(),
		Source: g.meta.Source,
	}
	k := store.KeyOf(r)
	prevBest, hadBest := store.Best(g.history, k)
	g.history = append(g.history, r)
	g.store.AppendRound(r)

	switch {
	case store.PBEligible(r) && r.WPM > 0 && (!hadBest || r.WPM > prevBest):
		g.resultLine = fmt.Sprintf(" NEW BEST! %.0f wpm ", r.WPM)
		g.resultBest = true
	case hadBest:
		spark := ui.Sparkline(store.RecentWPMs(g.history, k, 10))
		g.resultLine = fmt.Sprintf(" best %.0f %s recent %s ", prevBest, ui.Glyphs().Sep, spark)
		g.resultBest = false
	default:
		g.resultLine = ""
	}
}

// render draws the current theme plus the mode overlays (time-attack countdown
// and pause banner) on top.
func (g *Game) render() {
	g.theme.UpdateScreen(g.renderer, g.state)
	g.drawOverlay()
	g.screen.Show()
}

// drawTooSmall shows a notice sized to fit the (very narrow) terminal.
func (g *Game) drawTooSmall(w int) {
	g.screen.Clear()
	style := tcell.StyleDefault.Foreground(tcell.ColorRed)
	msg := fmt.Sprintf("Too narrow (min %d cols)", minWidth)
	if w < len([]rune(msg)) {
		msg = fmt.Sprintf("min %d", minWidth)
	}
	g.renderer.DrawText(0, 0, style, ui.Truncate(msg, w))
	g.screen.Show()
}

// drawRightAligned draws text so it ends at column `right` (exclusive). It
// returns the start column, so HUD fields can be chained leftward.
func (g *Game) drawRightAligned(text string, y, right int, style tcell.Style) int {
	x := right - runewidth.StringWidth(text)
	if x < 0 {
		x = 0
	}
	g.renderer.DrawText(x, y, style, text)
	return x
}

func (g *Game) drawOverlay() {
	w, h := g.screen.Size()
	gl := ui.Glyphs()

	// Top-right HUD: a countdown badge (time-attack) with the live WPM/accuracy
	// to its left while typing. Both are theme-independent so every theme gets a
	// consistent heads-up display.
	right := w
	if g.state.TimeLimit > 0 {
		var label string
		var style tcell.Style
		switch {
		case g.state.IsFinished && g.state.TimedOut:
			label = fmt.Sprintf(" %s TIME UP ", gl.Clock)
			style = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorMaroon)
		case g.state.IsFinished:
			label = fmt.Sprintf(" %s DONE ", gl.Clock)
			style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen)
		default:
			secs := int(g.state.Remaining().Seconds() + 0.999)
			label = fmt.Sprintf(" %s %d:%02d ", gl.Clock, secs/60, secs%60)
			if secs <= 5 {
				style = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRed)
			} else {
				style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
			}
		}
		right = g.drawRightAligned(label, 0, right, style)
	}

	if g.state.TimerStarted && !g.state.IsFinished && g.state.Elapsed() >= time.Second {
		wpm, acc := g.state.LiveStats()
		stats := fmt.Sprintf(" %.0f wpm  %.0f%% ", wpm, acc)
		// Only show the live stats if they fit without colliding with the badge.
		if runewidth.StringWidth(stats) <= right {
			g.drawRightAligned(stats, 0, right, tcell.StyleDefault.Foreground(tcell.ColorTeal))
		}
	}

	// Once the round is finished the live stats disappear, freeing the top-right
	// corner for the personal-best line.
	if g.state.IsFinished && g.resultLine != "" {
		style := tcell.StyleDefault.Foreground(tcell.ColorGray)
		if g.resultBest {
			style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen)
		}
		if runewidth.StringWidth(g.resultLine) <= right {
			g.drawRightAligned(g.resultLine, 0, right, style)
		}
	}

	if g.state.Paused {
		msg := fmt.Sprintf(" %s PAUSED - Ctrl-P to resume ", gl.Pause)
		if runewidth.StringWidth(msg) > w {
			msg = fmt.Sprintf(" %s PAUSED ", gl.Pause)
		}
		style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
		x := (w - runewidth.StringWidth(msg)) / 2
		if x < 0 {
			x = 0
		}
		g.renderer.DrawText(x, h/2, style, ui.Truncate(msg, w))
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
		g.finalizeRound()
	}
}
