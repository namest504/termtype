package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/store"
	"github.com/namest504/termtype/internal/ui"
)

// historyLine formats one round for the history list.
func historyLine(r store.Round) string {
	return fmt.Sprintf("%s  %-7s %-6s %-2s %-7s %4.0f wpm  %5.1f%%",
		r.TS.Format("2006-01-02 15:04"), r.Theme, r.Mode, r.Lang, r.Source, r.WPM, r.Acc)
}

// listTop scrolls the list window so the selection stays visible.
func listTop(top, sel, visible int) int {
	if sel < top {
		return sel
	}
	if sel >= top+visible {
		return sel - visible + 1
	}
	return top
}

// showHistory is the in-menu history browser: a newest-first list of rounds;
// Enter opens a round's detail (with the wpm graph when the round has a
// series), Esc goes back.
func showHistory(s tcell.Screen, rounds []store.Round) {
	// Newest first.
	rev := make([]store.Round, len(rounds))
	for i, r := range rounds {
		rev[len(rounds)-1-i] = r
	}

	sel, top := 0, 0
	detail := false

	for {
		s.Clear()
		w, h := s.Size()

		if detail {
			drawRoundDetail(s, rev[sel], w, h)
		} else if len(rev) == 0 {
			drawText(s, 2, 1, tcell.StyleDefault.Bold(true), "History")
			drawText(s, 2, 3, tcell.StyleDefault.Foreground(tcell.ColorGray), "No rounds yet.")
			drawText(s, 2, 5, tcell.StyleDefault.Foreground(tcell.ColorGray), "Esc back")
		} else {
			drawText(s, 2, 1, tcell.StyleDefault.Bold(true), "History")
			visible := h - 5
			if visible < 1 {
				visible = 1
			}
			top = listTop(top, sel, visible)
			for row := 0; row < visible && top+row < len(rev); row++ {
				style := tcell.StyleDefault
				if top+row == sel {
					style = style.Reverse(true)
				}
				drawText(s, 2, 3+row, style, ui.Truncate(historyLine(rev[top+row]), w-4))
			}
			gl := ui.Glyphs()
			help := gl.ArrowUD + " select " + gl.Sep + " " + gl.Enter + " detail " + gl.Sep + " Esc back"
			drawText(s, 2, h-1, tcell.StyleDefault.Foreground(tcell.ColorGray), ui.Truncate(help, w-4))
		}
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				if detail {
					detail = false
					continue
				}
				return
			case tcell.KeyUp:
				if !detail && sel > 0 {
					sel--
				}
			case tcell.KeyDown:
				if !detail && sel < len(rev)-1 {
					sel++
				}
			case tcell.KeyEnter:
				if !detail && len(rev) > 0 {
					detail = true
				}
			}
		}
	}
}

// drawRoundDetail draws one round: its wpm graph when a series was recorded,
// plus the summary and when/where it was played.
func drawRoundDetail(s tcell.Screen, r store.Round, w, h int) {
	renderer := ui.NewRenderer(s)

	centered := func(y int, style tcell.Style, text string) {
		x := (w - runewidth.StringWidth(text)) / 2
		if x < 0 {
			x = 0
		}
		drawText(s, x, y, style, ui.Truncate(text, w))
	}

	chartH := 0
	if len(r.WPMSeries) >= 2 && h >= 12 {
		chartH = h - 10
		if chartH > 10 {
			chartH = 10
		}
	}

	top := (h - (8 + chartH)) / 2
	if top < 1 {
		top = 1
	}

	centered(top, tcell.StyleDefault.Foreground(tcell.ColorGreen), fmt.Sprintf("wpm: %.0f", r.WPM))
	y := top + 2
	if chartH > 0 {
		chartW := w - 8
		if chartW > 64 {
			chartW = 64
		}
		ui.DrawLineChart(renderer, (w-chartW)/2, y, chartW, chartH, r.WPMSeries,
			tcell.StyleDefault.Foreground(tcell.ColorGray),
			tcell.StyleDefault.Foreground(tcell.ColorYellow))
		y += chartH + 1
	}

	summary := fmt.Sprintf("accuracy: %.1f", r.Acc)
	if r.RawWPM > 0 {
		summary += fmt.Sprintf("  raw: %.0f", r.RawWPM)
	}
	if r.CPM > 0 {
		summary += fmt.Sprintf("  cpm: %.0f", r.CPM)
	}
	summary += fmt.Sprintf("  time: %.0fs", r.DurS)
	centered(y, tcell.StyleDefault.Foreground(tcell.ColorTeal), summary)
	centered(y+2, tcell.StyleDefault.Foreground(tcell.ColorGray),
		fmt.Sprintf("%s  %s %s %s %s", r.TS.Format("2006-01-02 15:04"), r.Theme, r.Mode, r.Lang, r.Source))
	centered(y+4, tcell.StyleDefault.Foreground(tcell.ColorGray), "Esc back")
	s.HideCursor()
}
