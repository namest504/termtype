package themes

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["simple"] = &SimpleTheme{}
}

// SimpleTheme is the default plain typing view: the target text with live
// progress coloring and no decoration.
type SimpleTheme struct{}

func (t *SimpleTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
}

func (t *SimpleTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	renderer.Clear()
	if !gs.IsFinished {
		t.drawTypingScreen(renderer, gs)
	} else {
		t.drawResultScreen(renderer, gs)
	}
	renderer.Show()
}

func (t *SimpleTheme) drawTypingScreen(renderer domain.Renderer, gs *domain.GameState) {
	w, h := renderer.Size()
	rows := len(ui.WrapText(gs.TargetSentence, w-2))
	maxLines := h - 4
	if maxLines < 1 {
		maxLines = 1
	}
	if rows > maxLines {
		rows = maxLines
	}
	startY := (h - rows) / 2
	if startY < 1 {
		startY = 1
	}
	tr := &ui.TypingRenderer{}
	tr.Draw(renderer, gs, ui.TypingRendererOptions{
		StartY: startY, Width: w - 2, PrefixWidth: 0, CenterText: true, MaxLines: maxLines,
	})
	hint := "esc menu"
	renderer.DrawText((w-runewidth.StringWidth(hint))/2, h-2, tcell.StyleDefault.Foreground(tcell.ColorGray), hint)
}

func (t *SimpleTheme) drawResultScreen(renderer domain.Renderer, gs *domain.GameState) {
	renderer.HideCursor()
	w, h := renderer.Size()
	gl := ui.Glyphs()
	center := func(y int, style tcell.Style, s string) {
		x := (w - runewidth.StringWidth(s)) / 2
		if x < 0 {
			x = 0
		}
		renderer.DrawText(x, y, style, ui.Truncate(s, w))
	}
	center(h/2-1, tcell.StyleDefault, ui.ResultText(gs))
	center(h/2+1, tcell.StyleDefault.Foreground(tcell.ColorGray), gl.Send+" next "+gl.Sep+" esc menu")
}

func (t *SimpleTheme) OnTick(gs *domain.GameState) { /* Do nothing */ }
