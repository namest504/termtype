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

func (t *HexTheme) UpdateScreen(renderer *Renderer, gs *GameState) {
	state, ok := gs.CustomState.(*HexThemeState)
	if !ok {
		return
	}
	renderer.Clear()
	_, h := renderer.Size()

	// 화면 크기가 변경되었거나 처음 그릴 때 StartLine 설정
	if state.StartLine == -1 {
		state.StartLine = h / 2
	}

	t.drawHexDump(renderer, h)
	t.drawInputOverlay(renderer, gs, state)

	if gs.isFinished {
		t.drawResult(renderer, gs, h)
	} else {
		t.drawCursor(renderer, gs, state)
	}

	renderer.Show()
}

func (t *HexTheme) drawHexDump(renderer *Renderer, h int) {
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
		renderer.DrawText(0, y, addrStyle, offset)
		renderer.DrawText(10, y, hexStyle, hexStr)
		renderer.DrawText(62, y, asciiStyle, asciiStr)
	}
}

func (t *HexTheme) drawInputOverlay(renderer *Renderer, gs *GameState, state *HexThemeState) {
	hexStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	asciiStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	correctStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	incorrectStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	// 실제 타이핑할 문장을 중앙에 덮어쓰기
	targetBytes := []byte(gs.targetSentence)
	for i, b := range targetBytes {
		lineIdx := state.StartLine + (i / 16)
		charIdx := i % 16

		hexStr := fmt.Sprintf("%02x", b)
		asciiChar := "."
		if b >= 32 && b <= 126 {
			asciiChar = string(b)
		}

		renderer.SetContent(10+charIdx*3, lineIdx, []rune(hexStr)[0], hexStyle)
		renderer.SetContent(10+charIdx*3+1, lineIdx, []rune(hexStr)[1], hexStyle)
		renderer.SetContent(62+charIdx, lineIdx, []rune(asciiChar)[0], asciiStyle)
	}

	// 사용자 입력 피드백
	inputRunes := []rune(gs.userInput)
	for i, r := range inputRunes {
		lineIdx := state.StartLine + (i / 16)
		charIdx := i % 16
		style := correctStyle
		if i < len(targetBytes) && r != rune(targetBytes[i]) { // 타겟 바이트와 룬 비교
			style = incorrectStyle
		}
		if i < len(targetBytes) {
			renderer.SetContent(62+charIdx, lineIdx, rune(targetBytes[i]), style)
		}
	}
}

func (t *HexTheme) drawCursor(renderer *Renderer, gs *GameState, state *HexThemeState) {
	inputRunes := []rune(gs.userInput)
	cursorLine := state.StartLine + (len(inputRunes) / 16)
	cursorCol := len(inputRunes) % 16
	renderer.ShowCursor(62+cursorCol, cursorLine)
}

func (t *HexTheme) drawResult(renderer *Renderer, gs *GameState, h int) {
	renderer.HideCursor()
	resultText := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
	renderer.DrawText(0, h-1, tcell.StyleDefault, resultText)
}

func (t *HexTheme) OnTick(gs *GameState) {}
