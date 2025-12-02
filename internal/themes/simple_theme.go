package themes

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

func init() {
	Themes["simple"] = &SimpleTheme{}
}

// --- Simple Theme --- //

type SimpleTheme struct{}

func (t *SimpleTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = domain.Sentences[rand.Intn(len(domain.Sentences))]
}

func (t *SimpleTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	renderer.Clear()

	// 텍스트를 그릴 시작 Y 좌표
	startY := 1

	if !gs.IsFinished {
		t.drawTypingScreen(renderer, gs, startY)
	} else {
		t.drawResultScreen(renderer, gs, startY)
	}

	renderer.Show()
}

func (t *SimpleTheme) drawTypingScreen(renderer domain.Renderer, gs *domain.GameState, startY int) {
	w, _ := renderer.Size()
	tr := &ui.TypingRenderer{}
	tr.Draw(renderer, gs, ui.TypingRendererOptions{
		StartY:      startY,
		Width:       w - 2, // 좌우 패딩 1씩
		PrefixWidth: 0,
		CenterText:  false,
	})
	renderer.DrawText(1, startY+len(ui.WrapText(gs.TargetSentence, w-2))+1, tcell.StyleDefault.Foreground(tcell.ColorWhite), "(ESC to exit)")
}

func (t *SimpleTheme) drawResultScreen(renderer domain.Renderer, gs *domain.GameState, startY int) {
	renderer.HideCursor()
	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	resultText1 := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.Wpm, gs.Accuracy)
	resultText2 := "Press Enter to continue or ESC to exit."
	renderer.DrawText(1, startY, defStyle, resultText1)
	renderer.DrawText(1, startY+2, defStyle, resultText2)
}

func (t *SimpleTheme) OnTick(gs *domain.GameState) { /* Do nothing */ }
