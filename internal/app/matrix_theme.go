package app

import (
	"fmt"
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func init() {
	Themes["matrix"] = &MatrixTheme{}
}

// MatrixTheme은 떨어지는 글자 효과를 구현합니다.
type MatrixTheme struct{}

// Matrix의 각 "빗방울"의 상태
type Raindrop struct {
	X, Y   int
	Speed  int
	Chars  []rune
	Length int
}

// Matrix 테마의 전체 상태
type MatrixThemeState struct {
	drops []*Raindrop
	width, height int
}

var matrixChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:'\",./<>?")

func (t *MatrixTheme) ResetState(gs *GameState) {
	gs.resetCommon()
	gs.targetSentence = gs.sentences[rand.Intn(len(gs.sentences))]

	// MatrixThemeState를 초기화하지 않았으면 새로 만듭니다.
	if _, ok := gs.CustomState.(*MatrixThemeState); !ok {
		gs.CustomState = &MatrixThemeState{}
	}
}

func (t *MatrixTheme) UpdateScreen(renderer *Renderer, gs *GameState) {
	matrixState, ok := gs.CustomState.(*MatrixThemeState)
	if !ok {
		return // 상태가 아직 준비되지 않음
	}

	renderer.Clear()
	w, h := renderer.Size()
	if matrixState.width != w || matrixState.height != h {
		matrixState.width = w
		matrixState.height = h
		matrixState.drops = make([]*Raindrop, w)
		for i := 0; i < w; i++ {
			matrixState.drops[i] = &Raindrop{X: i, Y: rand.Intn(h), Speed: rand.Intn(4) + 1, Length: rand.Intn(10) + 5}
		}
	}

	// 배경을 검은색으로 채웁니다.
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			renderer.SetContent(x, y, ' ', tcell.StyleDefault.Background(tcell.ColorBlack))
		}
	}

	// 빗방울을 그립니다.
	for _, drop := range matrixState.drops {
		for i := 0; i < drop.Length; i++ {
			y := drop.Y - i
			if y >= 0 && y < h {
				style := tcell.StyleDefault.Foreground(tcell.ColorGreen)
				if i == 0 {
					style = tcell.StyleDefault.Foreground(tcell.ColorWhite)
				} else if i > drop.Length-3 {
					style = tcell.StyleDefault.Foreground(tcell.ColorDarkGreen)
				}
							if len(drop.Chars) > 0 {
								renderer.SetContent(drop.X, y, drop.Chars[i%len(drop.Chars)], style)
							}		}
		}
	}

	// 타이핑할 문장을 중앙에 그립니다.
	// targetY := h / 2
	// targetX := (w - len(gs.targetSentence)) / 2

	// targetRunes := []rune(gs.targetSentence)
	// inputRunes := []rune(gs.userInput)

	// for i, r := range targetRunes {
	// 	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	// 	if i < len(inputRunes) {
	// 		if inputRunes[i] == r {
	// 			style = tcell.StyleDefault.Foreground(tcell.ColorLawnGreen).Background(tcell.ColorBlack)
	// 		} else {
	// 			style = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack)
	// 		}
	// 	}
	// 	renderer.SetContent(targetX+i, targetY, r, style)
	// }

	// if gs.isFinished {
	// 	renderer.HideCursor()
	// 	resultText := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
	// 	renderer.DrawText((w-len(resultText))/2, targetY+2, tcell.StyleDefault.Background(tcell.ColorBlack), resultText)
	// } else {
	// 	cursorX := targetX + runewidth.StringWidth(gs.userInput)
	// 	renderer.ShowCursor(cursorX, targetY)
	// }

	// renderer.Show()

	// 텍스트를 그릴 시작 Y 좌표
	startY := h/2 - 2

	// 줄바꿈 처리된 대상 문장
	wrappedTarget := wrapText(gs.targetSentence, w-4) // 좌우 패딩 2씩
	inputRunes := []rune(gs.userInput)
	inputOffset := 0 // 현재 입력된 글자가 몇 번째 줄에 있는지 계산하기 위한 오프셋

	for lineIdx, line := range wrappedTarget {
		lineRunes := []rune(line)
		lineX := (w - runewidth.StringWidth(line)) / 2 // 각 줄을 중앙 정렬
		for charIdx, r := range lineRunes {
			currentInputIdx := inputOffset + charIdx
			style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

			if currentInputIdx < len(inputRunes) {
				if inputRunes[currentInputIdx] == r {
					style = tcell.StyleDefault.Foreground(tcell.ColorLawnGreen).Background(tcell.ColorBlack)
				} else {
					style = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack)
				}
			}
			renderer.SetContent(lineX+charIdx, startY+lineIdx, r, style)
		}
		inputOffset += len(lineRunes)
	}

	if gs.isFinished {
		renderer.HideCursor()
		resultText := fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.wpm, gs.accuracy)
		renderer.DrawText((w-len(resultText))/2, startY+len(wrappedTarget)+1, tcell.StyleDefault.Background(tcell.ColorBlack), resultText)
	} else {
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
		currentLine := wrappedTarget[cursorLineIdx]
		cursorX := (w - runewidth.StringWidth(currentLine)) / 2 + cursorCharIdx
		renderer.ShowCursor(cursorX, startY+cursorLineIdx)
	}

	renderer.Show()
}

func (t *MatrixTheme) OnTick(gs *GameState) {
	matrixState, ok := gs.CustomState.(*MatrixThemeState)
	if !ok || matrixState.drops == nil {
		return
	}

	for _, drop := range matrixState.drops {
		drop.Y += drop.Speed
		if drop.Y-drop.Length > matrixState.height {
			drop.Y = 0
		}
		// 빗방울의 문자 내용을 계속 바꿔줍니다.
		drop.Chars = append(drop.Chars, matrixChars[rand.Intn(len(matrixChars))])
		if len(drop.Chars) > 50 {
			drop.Chars = drop.Chars[1:]
		}
	}
}
