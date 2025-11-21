package app

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

func init() {
	Themes["simple"] = &SimpleTheme{}
}

// --- Simple Theme --- //

type SimpleTheme struct{}

func (t *SimpleTheme) ResetState(gs *GameState) {
	gs.resetCommon()
	gs.targetSentence = gs.sentences[rand.Intn(len(gs.sentences))]
}

func (t *SimpleTheme) UpdateScreen(renderer *Renderer, gs *GameState) {
	renderer.Clear()

	// 텍스트를 그릴 시작 Y 좌표
	startY := 1

	if !gs.isFinished {
		t.drawTypingScreen(renderer, gs, startY)
	} else {
		t.drawResultScreen(renderer, gs, startY)
	}

	renderer.Show()
}

func (t *SimpleTheme) drawTypingScreen(renderer *Renderer, gs *GameState, startY int) {
	w, _ := renderer.Size()
	tr := &TypingRenderer{}
	tr.Draw(renderer, gs, TypingRendererOptions{
		StartY:      startY,
		Width:       w - 2, // 좌우 패딩 1씩
		PrefixWidth: 0,
		CenterText:  false,
	})
	renderer.DrawText(1, startY+len(wrapText(gs.targetSentence, w-2))+1, tcell.StyleDefault.Foreground(tcell.ColorWhite), "(ESC to exit)")
}

func (t *SimpleTheme) drawResultScreen(renderer *Renderer, gs *GameState, startY int) {
	renderer.HideCursor()
	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	resultText1 := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
	resultText2 := "Press Enter to continue or ESC to exit."
	renderer.DrawText(1, startY, defStyle, resultText1)
	renderer.DrawText(1, startY+2, defStyle, resultText2)
}

func (t *SimpleTheme) OnTick(gs *GameState) { /* Do nothing */ }
