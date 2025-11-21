package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// TypingRendererOptions 타이핑 영역 렌더링 옵션
type TypingRendererOptions struct {
	// StartY 그리기 시작 Y 좌표
	StartY int
	// Width 텍스트 줄바꿈 가용 너비
	Width int
	// PrefixWidth 접두어(로그 타임스탬프 등) 너비
	PrefixWidth int
	// CenterText 텍스트 중앙 정렬 여부
	CenterText bool
}

// TypingRenderer 타이핑 영역 및 커서 그리기 공통 로직
type TypingRenderer struct{}

// Draw 대상 문장, 사용자 입력, 커서 렌더링
func (tr *TypingRenderer) Draw(renderer *Renderer, gs *GameState, opts TypingRendererOptions) {
	// 텍스트 가용 너비 계산
	availableWidth := opts.Width
	if !opts.CenterText {
		availableWidth -= opts.PrefixWidth
	}

	// 텍스트 줄바꿈
	wrappedTarget := wrapText(gs.targetSentence, availableWidth)
	inputRunes := []rune(gs.userInput)
	inputOffset := 0

	// 텍스트 그리기
	for lineIdx, line := range wrappedTarget {
		lineRunes := []rune(line)

		// 현재 줄 시작 X 좌표 계산
		startX := 1 + opts.PrefixWidth
		if opts.CenterText {
			startX = (opts.Width - runewidth.StringWidth(line)) / 2
		}

		currentX := startX
		for charIdx, r := range lineRunes {
			currentInputIdx := inputOffset + charIdx

			// 스타일 결정
			style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
			if opts.CenterText {
				style = style.Background(tcell.ColorBlack)
			} else {
				style = tcell.StyleDefault.Foreground(tcell.ColorGray)
			}

			if currentInputIdx < len(inputRunes) {
				if inputRunes[currentInputIdx] == r {
					if opts.CenterText {
						style = tcell.StyleDefault.Foreground(tcell.ColorLawnGreen).Background(tcell.ColorBlack)
					} else {
						style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
					}
				} else {
					if opts.CenterText {
						style = tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlack)
					} else {
						style = tcell.StyleDefault.Foreground(tcell.ColorRed)
					}
				}
			}

			width := renderer.DrawRune(currentX, opts.StartY+lineIdx, r, style)
			currentX += width
		}
		inputOffset += len(lineRunes)
	}

	// 커서 그리기
	tr.drawCursor(renderer, wrappedTarget, inputRunes, opts)
}

func (tr *TypingRenderer) drawCursor(renderer *Renderer, wrappedTarget []string, inputRunes []rune, opts TypingRendererOptions) {
	cursorLineIdx := 0
	cursorX := 1 + opts.PrefixWidth
	if opts.CenterText {
		cursorX = 0
	}

	currentOffset := 0
	foundCursor := false

	for i, line := range wrappedTarget {
		lineRunes := []rune(line)
		lineLen := len(lineRunes)

		if len(inputRunes) >= currentOffset && len(inputRunes) <= currentOffset+lineLen {
			cursorLineIdx = i
			cursorRelIdx := len(inputRunes) - currentOffset

			startX := 1 + opts.PrefixWidth
			if opts.CenterText {
				startX = (opts.Width - runewidth.StringWidth(line)) / 2
			}

			cursorX = startX + runewidth.StringWidth(string(lineRunes[:cursorRelIdx]))
			foundCursor = true
			break
		}
		currentOffset += lineLen
	}

	if !foundCursor && len(wrappedTarget) > 0 {
		cursorLineIdx = len(wrappedTarget) - 1
		lastLine := wrappedTarget[len(wrappedTarget)-1]

		startX := 1 + opts.PrefixWidth
		if opts.CenterText {
			startX = (opts.Width - runewidth.StringWidth(lastLine)) / 2
		}

		cursorX = startX + runewidth.StringWidth(lastLine)
	}

	renderer.ShowCursor(cursorX, opts.StartY+cursorLineIdx)
}
