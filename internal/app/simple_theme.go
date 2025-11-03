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

func (t *SimpleTheme) UpdateScreen(renderer *Renderer, gs *GameState) {
	defStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	correctStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	incorrectStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	renderer.Clear()
	w, _ := renderer.Size()

	// 텍스트를 그릴 시작 Y 좌표
	startY := 1

	if !gs.isFinished {
		// 타이핑 중 화면
		wrappedTarget := wrapText(gs.targetSentence, w-2) // 좌우 패딩 1씩
		inputRunes := []rune(gs.userInput)
		inputOffset := 0 // 현재 입력된 글자가 몇 번째 줄에 있는지 계산하기 위한 오프셋

		for lineIdx, line := range wrappedTarget {
			lineRunes := []rune(line)
			for charIdx, r := range lineRunes {
				currentInputIdx := inputOffset + charIdx
				style := defStyle

				if currentInputIdx < len(inputRunes) {
					if inputRunes[currentInputIdx] == r {
						style = correctStyle
					} else {
						style = incorrectStyle
					}
				}
				renderer.SetContent(1+charIdx, startY+lineIdx, r, style)
			}
			inputOffset += len(lineRunes) // 다음 줄의 입력 오프셋 계산
		}

		renderer.DrawText(1, startY+len(wrappedTarget)+1, defStyle, "(ESC to exit)")

		// 커서 위치 설정
		cursorLineIdx := 0
		cursorCharIdx := 0
		currentOffset := 0
		for i, line := range wrappedTarget {
			lineLen := len([]rune(line))
			if len(inputRunes) >= currentOffset && len(inputRunes) <= currentOffset+lineLen {
				cursorLineIdx = i
				cursorCharIdx = runewidth.StringWidth(string(inputRunes[currentOffset:len(inputRunes)]))
				break
			}
			currentOffset += lineLen
		}
		renderer.ShowCursor(1+cursorCharIdx, startY+cursorLineIdx)

	} else {
		// 결과 화면
		renderer.HideCursor()
		resultText1 := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
		resultText2 := "Press Enter to continue or ESC to exit."
		renderer.DrawText(1, startY, defStyle, resultText1)
		renderer.DrawText(1, startY+2, defStyle, resultText2)
	}

	renderer.Show()
}

func (t *SimpleTheme) OnTick(gs *GameState) { /* Do nothing */ }
