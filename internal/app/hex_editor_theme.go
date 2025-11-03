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
	state, ok := gs.CustomState.(*HexThemeState)
	if !ok {
		return
	}
	s.Clear()
	_, h := s.Size()

	// 화면 크기가 변경되었거나 처음 그릴 때 StartLine 설정
	if state.StartLine == -1 {
		state.StartLine = h / 2
	}

	addrStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	hexStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	asciiStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	// 화면 전체에 임의의 헥스 데이터 그리기
	for y := 0; y < h; y++ {
		offset := fmt.Sprintf("%08x", y*16)
		hexStr, asciiStr := "", ""
		for i := 0; i < 16; i++ {
			randByte := byte(rand.Intn(256))
			hexStr += fmt.Sprintf("%02x ", randByte)
			if randByte >= 32 && randByte <= 126 {
				asciiStr += string(randByte)
			} else {
				asciiStr += "."
			}
		}
		drawText(s, 0, y, addrStyle, offset)
		drawText(s, 10, y, hexStyle, hexStr)
		drawText(s, 62, y, asciiStyle, asciiStr)
	}

	s.Show()
}

func (t *HexTheme) OnTick(gs *GameState) {}
