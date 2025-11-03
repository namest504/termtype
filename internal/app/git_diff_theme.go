package app

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func init() {
	Themes["diff"] = &DiffTheme{}
}

// DiffTheme는 git diff UI를 흉내 냅니다.
type DiffTheme struct{}

type DiffThemeState struct {
	ContextLines []string
}

var fakeCode = []string{
	" func main() {",
	" ",
	" ",
	" ",
	" }",
}

func (t *DiffTheme) ResetState(gs *GameState) {
	gs.resetCommon()
	gs.targetSentence = gs.sentences[rand.Intn(len(gs.sentences))]

	state := &DiffThemeState{}
	state.ContextLines = make([]string, 5)
	copy(state.ContextLines, fakeCode)
	gs.CustomState = state
}

func (t *DiffTheme) UpdateScreen(renderer *Renderer, gs *GameState) {
	state, ok := gs.CustomState.(*DiffThemeState)
	if !ok {
		return
	}
	renderer.Clear()

	// 컨텍스트 라인 그리기
	renderer.DrawText(0, 0, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "diff --git a/main.go b/main.go")
	renderer.DrawText(0, 1, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "--- a/main.go")
	renderer.DrawText(0, 2, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "+++ b/main.go")
	renderer.DrawText(0, 3, tcell.StyleDefault.Foreground(tcell.ColorBlue), "@@ -1,5 +1,5 @@")

	y := 4
	for i, line := range state.ContextLines {
		if i == 2 { // 문장이 들어갈 위치
			plusStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
			renderer.DrawText(0, y, plusStyle, "+ "+gs.targetSentence)

			// 사용자 입력 피드백
			for i, r := range []rune(gs.userInput) {
				style := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorDarkGreen)
				if i < len([]rune(gs.targetSentence)) && r != []rune(gs.targetSentence)[i] {
					style = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorDarkRed)
				}
				renderer.SetContent(i+2, y, []rune(gs.targetSentence)[i], style)
			}
		} else {
			renderer.DrawText(0, y, tcell.StyleDefault, " "+line)
		}
		y++
	}

	if gs.isFinished {
		renderer.HideCursor()
		resultText := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
		renderer.DrawText(0, y+2, tcell.StyleDefault, resultText)
	} else {
		cursorX := 2 + runewidth.StringWidth(gs.userInput)
		renderer.ShowCursor(cursorX, 6)
	}

	renderer.Show()
}

func (t *DiffTheme) OnTick(gs *GameState) {}