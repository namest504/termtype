package app

import (
	"math/rand"

	"github.com/gdamore/tcell/v2"
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

func (t *DiffTheme) UpdateScreen(s tcell.Screen, gs *GameState) {
	state, ok := gs.CustomState.(*DiffThemeState)
	if !ok {
		return
	}
	s.Clear()

	// 컨텍스트 라인 그리기
	drawText(s, 0, 0, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "diff --git a/main.go b/main.go")
	drawText(s, 0, 1, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "--- a/main.go")
	drawText(s, 0, 2, tcell.StyleDefault.Foreground(tcell.ColorDimGray), "+++ b/main.go")
	drawText(s, 0, 3, tcell.StyleDefault.Foreground(tcell.ColorBlue), "@@ -1,5 +1,5 @@")

	y := 4
	for i, line := range state.ContextLines {
		if i == 2 { // 문장이 들어갈 위치
			plusStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
			drawText(s, 0, y, plusStyle, "+ "+gs.targetSentence)
		} else {
			drawText(s, 0, y, tcell.StyleDefault, " "+line)
		}
		y++
	}

	s.Show()
}

func (t *DiffTheme) OnTick(gs *GameState) {}