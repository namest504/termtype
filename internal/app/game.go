package app

import (
	"fmt"
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
	showGraph  bool          // finished round: the wpm graph view is up
}

// Create a new game. timeLimit > 0 enables time-attack mode. sentences is the
// pool the chosen theme draws targets from; an empty pool falls back to the
// default English set. targetGen, when non-nil, replaces the pool entirely
// (the "words" source). meta labels recorded rounds; st may be nil to disable
// persistence.
func NewGame(s tcell.Screen, theme domain.Theme, timeLimit time.Duration, sentences []string, targetGen func() string, meta RoundMeta, st *store.Store) (*Game, error) {
	if len(sentences) == 0 {
		sentences = domain.Sentences
	}
	state := &domain.GameState{Sentences: sentences, TargetGen: targetGen, TimeLimit: timeLimit}
	theme.ResetState(state)

	return &Game{screen: s, renderer: ui.NewRenderer(s), state: state, theme: theme,
		meta: meta, store: st, history: st.LoadHistory()}, nil
}

// Run plays rounds until the player leaves: Esc returns false (back to the
// menu), Ctrl-C returns true (quit the program). events is the shared
// screen-event channel owned by main.
func (g *Game) Run(events <-chan tcell.Event) (quit bool) {
	ticker := time.NewTicker(1 * time.Second) // Tick every second
	defer ticker.Stop()

	g.render()

	for {
		select {
		case ev := <-events:
			switch ev := ev.(type) {
			case nil:
				return true // screen torn down
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
					// Only the leave keys work while the screen is too small.
					if ev.Key() == tcell.KeyEscape {
						return false
					}
					if ev.Key() == tcell.KeyCtrlC {
						return true
					}
				} else {
					if back, quit := g.handleKeyEvent(ev); back {
						return quit
					}
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
						g.state.SampleWPM()
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
		TS:        time.Now(),
		Theme:     g.meta.Theme,
		Mode:      store.ModeString(g.state.TimeLimit),
		Lang:      g.meta.Lang,
		WPM:       g.state.Wpm,
		Acc:       g.state.Accuracy,
		DurS:      g.state.FinalDurS,
		Source:    g.meta.Source,
		RawWPM:    g.state.FinalRawWPM,
		CPM:       g.state.FinalCPM,
		WPMSeries: g.state.WPMSamples,
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
// and pause banner) on top — or the wpm graph view when it is toggled up.
func (g *Game) render() {
	if g.state.IsFinished && g.showGraph {
		g.drawGraphView()
		g.screen.Show()
		return
	}
	g.theme.UpdateScreen(g.renderer, g.state)
	g.drawOverlay()
	g.screen.Show()
}

// drawGraphView is the full-screen result graph: the round's WPM over time
// with the accuracy/raw/cpm/time summary underneath.
func (g *Game) drawGraphView() {
	g.renderer.Clear()
	w, h := g.screen.Size()

	centered := func(y int, style tcell.Style, text string) {
		x := (w - runewidth.StringWidth(text)) / 2
		if x < 0 {
			x = 0
		}
		g.renderer.DrawText(x, y, style, ui.Truncate(text, w))
	}

	chartW := w - 8
	if chartW > 64 {
		chartW = 64
	}
	chartH := h - 8
	if chartH > 10 {
		chartH = 10
	}
	if chartH < 2 {
		chartH = 2
	}
	top := (h - (chartH + 6)) / 2
	if top < 0 {
		top = 0
	}

	centered(top, tcell.StyleDefault.Foreground(tcell.ColorGreen), fmt.Sprintf("wpm: %.0f", g.state.Wpm))
	ui.DrawLineChart(g.renderer, (w-chartW)/2, top+2, chartW, chartH, g.state.WPMSamples,
		tcell.StyleDefault.Foreground(tcell.ColorGray),
		tcell.StyleDefault.Foreground(tcell.ColorYellow))
	centered(top+chartH+3, tcell.StyleDefault.Foreground(tcell.ColorTeal),
		fmt.Sprintf("accuracy: %.1f  raw: %.0f  cpm: %.0f  time: %.0fs",
			g.state.Accuracy, g.state.FinalRawWPM, g.state.FinalCPM, g.state.FinalDurS))
	centered(top+chartH+5, tcell.StyleDefault.Foreground(tcell.ColorGray), "g back  enter next  esc menu")
	g.renderer.HideCursor()
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

	// Point at the graph view once a round has enough samples to draw one.
	if g.state.IsFinished && len(g.state.WPMSamples) >= 2 {
		g.drawRightAligned(" g graph ", 1, w, tcell.StyleDefault.Foreground(tcell.ColorGray))
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

// handleKeyEvent processes one key. back means leave the game loop; quit
// distinguishes Ctrl-C (quit the program) from Esc (back to the menu).
func (g *Game) handleKeyEvent(ev *tcell.EventKey) (back, quit bool) {
	if ev.Key() == tcell.KeyEscape {
		return true, false
	}
	if ev.Key() == tcell.KeyCtrlC {
		return true, true
	}

	// Ctrl-P toggles pause; while paused, all other input is ignored.
	if ev.Key() == tcell.KeyCtrlP {
		g.state.TogglePause()
		return false, false
	}
	if g.state.Paused {
		return false, false
	}

	if g.state.IsFinished {
		switch {
		case ev.Key() == tcell.KeyEnter:
			g.showGraph = false
			g.theme.ResetState(g.state)
		case ev.Key() == tcell.KeyRune && (ev.Rune() == 'g' || ev.Rune() == 'G'):
			g.showGraph = !g.showGraph
		}
		return false, false
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
		g.state.TypedRunes++
	}

	// Finalize once the whole sentence has been typed (rune count, not bytes).
	if len([]rune(g.state.UserInput)) >= len([]rune(g.state.TargetSentence)) {
		g.finalizeRound()
	}
	return false, false
}
