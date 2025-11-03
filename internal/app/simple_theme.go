package app

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
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

func (t *SimpleTheme) UpdateScreen(s tcell.Screen, gs *GameState) {
	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	correctStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	incorrectStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	s.Clear()

	if !gs.isFinished {
		// 타이핑 중 화면
		targetRunes := []rune(gs.targetSentence)
		inputRunes := []rune(gs.userInput)

		for i, r := range targetRunes {
			style := defStyle
			if i < len(inputRunes) {
				if inputRunes[i] == r {
					style = correctStyle
				} else {
					style = incorrectStyle
				}
			}
			s.SetContent(i+1, 1, r, nil, style)
		}
		drawText(s, 1, 3, defStyle, "(ESC to exit)")

		// 커서 위치 설정
		cursorX := 1 + runewidth.StringWidth(gs.userInput)
		s.ShowCursor(cursorX, 1)

	} else {
		// 결과 화면
		s.HideCursor()
		resultText1 := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
		resultText2 := "Press Enter to continue or ESC to exit."
		drawText(s, 1, 1, defStyle, resultText1)
		drawText(s, 1, 3, defStyle, resultText2)
	}

	s.Show()
}

func (t *SimpleTheme) OnTick(gs *GameState) { /* Do nothing */ }
