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

func (t *DiffTheme) UpdateScreen(s tcell.Screen, gs *GameState) {
}

func (t *DiffTheme) OnTick(gs *GameState) {}