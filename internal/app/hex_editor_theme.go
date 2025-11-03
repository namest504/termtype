package app

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

func init() {
	Themes["hex"] = &HexTheme{}
}

// HexTheme는 헥스 에디터 UI를 흉내 냅니다.
type HexTheme struct{}

type HexThemeState struct {
	StartLine int
}

func (t *HexTheme) ResetState(gs *GameState) {
	gs.resetCommon()
	gs.targetSentence = gs.sentences[rand.Intn(len(gs.sentences))]
	gs.CustomState = &HexThemeState{StartLine: -1} // StartLine을 -1로 초기화하여 첫 UpdateScreen에서 설정하도록 함
}

func (t *HexTheme) UpdateScreen(s tcell.Screen, gs *GameState) {
}

func (t *HexTheme) OnTick(gs *GameState) {}