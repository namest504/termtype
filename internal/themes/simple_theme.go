package themes

import (
	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["simple"] = &SimpleTheme{}
}

// --- Simple Theme --- //

type SimpleTheme struct{}

func (t *SimpleTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
}

func (t *SimpleTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	renderer.Clear()

	// Starting Y coordinate for drawing text
	startY := 1

	if !gs.IsFinished {
		t.drawTypingScreen(renderer, gs, startY)
	} else {
		t.drawResultScreen(renderer, gs, startY)
	}

	renderer.Show()
}

func (t *SimpleTheme) drawTypingScreen(renderer domain.Renderer, gs *domain.GameState, startY int) {
	w, h := renderer.Size()
	maxLines := h - startY - 3 // leave room for the hint line below
	if maxLines < 1 {
		maxLines = 1
	}
	tr := &ui.TypingRenderer{}
	rows := tr.Draw(renderer, gs, ui.TypingRendererOptions{
		StartY:      startY,
		Width:       w - 2, // 1 padding on each side
		PrefixWidth: 0,
		CenterText:  false,
		MaxLines:    maxLines,
	})
	renderer.DrawText(1, startY+rows+1, tcell.StyleDefault.Foreground(tcell.ColorWhite), "(Esc for menu)")
}

func (t *SimpleTheme) drawResultScreen(renderer domain.Renderer, gs *domain.GameState, startY int) {
	renderer.HideCursor()
	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	resultText1 := ui.ResultText(gs)
	resultText2 := "Press Enter to continue or ESC to exit."
	renderer.DrawText(1, startY, defStyle, resultText1)
	renderer.DrawText(1, startY+2, defStyle, resultText2)
}

func (t *SimpleTheme) OnTick(gs *domain.GameState) { /* Do nothing */ }
