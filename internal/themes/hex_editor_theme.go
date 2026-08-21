package themes

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["hex"] = &HexTheme{}
}

// HexTheme mimics a hex editor UI. The target is laid out in 16-byte rows
// over a stable background dump; a few background bytes flip each tick, like
// memory being written. Everything is byte-addressed — a hex dump of UTF-8 —
// so wide runes (e.g. Hangul) occupy several cells and the cursor tracks the
// byte offset.
type HexTheme struct{}

// HexThemeState is HexTheme's per-round state: the stable background dump
// and the current scroll offset into it.
type HexThemeState struct {
	StartLine int
	rows      [][]byte // stable background dump, one 16-byte row per line
}

func (t *HexTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
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
	t.ensureDump(state, h)

	t.drawHexDump(renderer, state, h)
	winStart := t.drawTarget(renderer, gs, state, h)

	if gs.IsFinished {
		t.drawResult(renderer, gs, state, h)
	} else {
		t.drawCursor(renderer, gs, state, winStart)
	}

	renderer.Show()
}

// ensureDump keeps a stable random dump at least h rows tall, so the
// background doesn't reroll (and flicker) on every keystroke.
func (t *HexTheme) ensureDump(state *HexThemeState, h int) {
	for len(state.rows) < h {
		row := make([]byte, 16)
		for i := range row {
			row[i] = byte(rand.Intn(256))
		}
		state.rows = append(state.rows, row)
	}
}

func (t *HexTheme) drawHexDump(renderer domain.Renderer, state *HexThemeState, h int) {
	addrStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	hexStyle := tcell.StyleDefault
	asciiStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)

	for y := 0; y < h && y < len(state.rows); y++ {
		offset := fmt.Sprintf("%08x", y*16)
		hexStr, asciiStr := "", ""
		for _, b := range state.rows[y] {
			hexStr += fmt.Sprintf("%02x ", b)
			if b >= 32 && b <= 126 {
				asciiStr += string(rune(b))
			} else {
				asciiStr += "."
			}
		}
		renderer.DrawText(0, y, addrStyle, offset)
		renderer.DrawText(10, y, hexStyle, hexStr)
		renderer.DrawText(62, y, asciiStyle, asciiStr)
	}
}

// hexWindow returns the first visible 16-byte row of the target and how many
// rows fit, keeping the row being typed on screen.
func hexWindow(targetLen, inputLen, availRows int) (winStart, visible int) {
	totalRows := (targetLen + 15) / 16
	if totalRows < 1 {
		totalRows = 1
	}
	visible = availRows
	if visible < 1 {
		visible = 1
	}
	if visible > totalRows {
		visible = totalRows
	}
	cursorRow := inputLen / 16
	if cursorRow > totalRows-1 {
		cursorRow = totalRows - 1
	}
	return ui.WindowStart(totalRows, cursorRow, visible), visible
}

// drawTarget lays the target sentence over the dump in 16-byte rows,
// coloring the ascii pane with the typed input. It returns the window start
// row so the cursor can be placed.
func (t *HexTheme) drawTarget(renderer domain.Renderer, gs *domain.GameState, state *HexThemeState, h int) int {
	hexStyle := tcell.StyleDefault
	asciiStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	correctStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	incorrectStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	targetBytes := []byte(gs.TargetSentence)
	inputBytes := []byte(gs.UserInput)
	winStart, visible := hexWindow(len(targetBytes), len(inputBytes), h-state.StartLine-1)

	for i, b := range targetBytes {
		row := i/16 - winStart
		if row < 0 || row >= visible {
			continue
		}
		lineIdx := state.StartLine + row
		charIdx := i % 16

		hexStr := fmt.Sprintf("%02x", b)
		asciiChar := "."
		if b >= 32 && b <= 126 {
			asciiChar = string(rune(b))
		}

		renderer.SetContent(10+charIdx*3, lineIdx, []rune(hexStr)[0], hexStyle)
		renderer.SetContent(10+charIdx*3+1, lineIdx, []rune(hexStr)[1], hexStyle)
		renderer.SetContent(62+charIdx, lineIdx, []rune(asciiChar)[0], asciiStyle)
	}

	// User input feedback in the ascii pane, byte for byte.
	for i, b := range inputBytes {
		if i >= len(targetBytes) {
			break
		}
		row := i/16 - winStart
		if row < 0 || row >= visible {
			continue
		}
		style := correctStyle
		if b != targetBytes[i] {
			style = incorrectStyle
		}
		renderer.SetContent(62+i%16, state.StartLine+row, rune(targetBytes[i]), style)
	}
	return winStart
}

func (t *HexTheme) drawCursor(renderer domain.Renderer, gs *domain.GameState, state *HexThemeState, winStart int) {
	byteIdx := len([]byte(gs.UserInput))
	row := byteIdx/16 - winStart
	if row < 0 {
		row = 0
	}
	renderer.ShowCursor(62+byteIdx%16, state.StartLine+row)
}

// drawResult writes the round's stats INTO the dump as real encoded rows:
// the stat string's UTF-8 bytes appear in the hex columns and the string in
// the ascii gutter, highlighted, right below the target rows.
func (t *HexTheme) drawResult(renderer domain.Renderer, gs *domain.GameState, state *HexThemeState, h int) {
	renderer.HideCursor()
	statStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	addrStyle := tcell.StyleDefault.Foreground(tcell.ColorBlue)

	stats := []byte(ui.ResultText(gs))
	_, visible := hexWindow(len([]byte(gs.TargetSentence)), len([]byte(gs.UserInput)), h-state.StartLine-1)
	startRow := state.StartLine + visible + 1 // one blank dump row of breathing room

	statRows := (len(stats) + 15) / 16
	if statRows < 1 {
		statRows = 1
	}
	if startRow+statRows > h {
		startRow = h - statRows
	}
	if startRow < 0 {
		startRow = 0
	}

	for r := 0; r*16 < len(stats); r++ {
		y := startRow + r
		if y >= h {
			break
		}
		chunk := stats[r*16 : min(len(stats), r*16+16)]
		// clear the whole dump row first so no background bytes bleed through
		for x := 0; x < 62+16; x++ {
			renderer.SetContent(x, y, ' ', tcell.StyleDefault)
		}
		renderer.DrawText(0, y, addrStyle, fmt.Sprintf("%08x", (startRow+r)*16))
		hexStr, asciiStr := "", ""
		for _, b := range chunk {
			hexStr += fmt.Sprintf("%02x ", b)
			if b >= 32 && b <= 126 {
				asciiStr += string(rune(b))
			} else {
				asciiStr += "."
			}
		}
		renderer.DrawText(10, y, statStyle, hexStr)
		renderer.DrawText(62, y, statStyle, asciiStr)
	}
}

// OnTick flips a few background bytes, like memory being written.
func (t *HexTheme) OnTick(gs *domain.GameState) {
	state, ok := gs.CustomState.(*HexThemeState)
	if !ok || len(state.rows) == 0 {
		return
	}
	for i := 0; i < 8; i++ {
		row := state.rows[rand.Intn(len(state.rows))]
		row[rand.Intn(len(row))] = byte(rand.Intn(256))
	}
}
