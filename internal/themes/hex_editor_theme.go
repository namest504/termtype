package themes

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

func init() {
	Themes["hex"] = &HexTheme{}
}

// HexTheme mimics a hex editor UI.
type HexTheme struct{}

type HexThemeState struct {
	StartLine int
}

func (t *HexTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = domain.Sentences[rand.Intn(len(domain.Sentences))]
	gs.CustomState = &HexThemeState{StartLine: -1} // Init StartLine to -1 so the first UpdateScreen sets it
}

func (t *HexTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	state, ok := gs.CustomState.(*HexThemeState)
	if !ok {
		return
	}
	renderer.Clear()
	_, h := renderer.Size()

	// Center the input area on screen (recomputed on every resize)
	state.StartLine = h / 2

	t.drawHexDump(renderer, h)
	t.drawInputOverlay(renderer, gs, state)

	if gs.IsFinished {
		t.drawResult(renderer, gs, h)
	} else {
		t.drawCursor(renderer, gs, state)
	}

	renderer.Show()
}

func (t *HexTheme) drawHexDump(renderer domain.Renderer, h int) {
	addrStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	hexStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	asciiStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	// Draw random hex data across the whole screen
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

func (t *HexTheme) drawInputOverlay(renderer domain.Renderer, gs *domain.GameState, state *HexThemeState) {
	hexStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	asciiStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	correctStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	incorrectStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	// Display the target sentence in 16-byte rows
	targetBytes := []byte(gs.TargetSentence)
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

	// User input feedback
	inputBytes := []byte(gs.UserInput)
	for i, r := range inputBytes {
		lineIdx := state.StartLine + (i / 16)
		charIdx := i % 16
		style := correctStyle
		if i < len(targetBytes) && r != targetBytes[i] { // compare target byte with rune
			style = incorrectStyle
		}
		if i < len(targetBytes) {
			renderer.SetContent(62+charIdx, lineIdx, rune(targetBytes[i]), style)
		}
	}
}

func (t *HexTheme) drawCursor(renderer domain.Renderer, gs *domain.GameState, state *HexThemeState) {
	inputRunes := []rune(gs.UserInput)
	cursorLine := state.StartLine + (len(inputRunes) / 16)
	cursorCol := len(inputRunes) % 16
	renderer.ShowCursor(62+cursorCol, cursorLine)
}

func (t *HexTheme) drawResult(renderer domain.Renderer, gs *domain.GameState, h int) {
	renderer.HideCursor()
	resultText := ui.ResultText(gs)
	renderer.DrawText(0, h-1, tcell.StyleDefault, resultText)
}

func (t *HexTheme) OnTick(gs *domain.GameState) {}
