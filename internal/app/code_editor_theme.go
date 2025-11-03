package app

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func init() {
	Themes["code"] = &CodeTheme{}
}

// CodeTheme는 코드 에디터 UI를 흉내 냅니다.
type CodeTheme struct{}

type CodeThemeState struct {
	FormattedLine string
	Language      string
	Keywords      []string
}

var codeFormats = map[string]struct {
	Format   string
	Keywords []string
}{
	"Go":         {"fmt.Println(\"%s\")", []string{"fmt.Println"}},
	"JavaScript": {"console.log(\"%s\");", []string{"console.log"}},
	"Python":     {"print(\"%s\")", []string{"print"}},
}

func (t *CodeTheme) ResetState(gs *GameState) {
	gs.resetCommon()

	// 랜덤 언어와 포맷 선택
	langKeys := make([]string, 0, len(codeFormats))
	for k := range codeFormats {
		langKeys = append(langKeys, k)
	}
	lang := langKeys[rand.Intn(len(langKeys))]
	formatInfo := codeFormats[lang]

	gs.targetSentence = gs.sentences[rand.Intn(len(gs.sentences))]
	gs.CustomState = &CodeThemeState{
		FormattedLine: fmt.Sprintf(formatInfo.Format, gs.targetSentence),
		Language:      lang,
		Keywords:      formatInfo.Keywords,
	}
}

func (t *CodeTheme) UpdateScreen(r *Renderer, gs *GameState) {
	state, ok := gs.CustomState.(*CodeThemeState)
	if !ok {
		return
	}

	r.Clear()
	w, h := r.Size()

	// 라인 번호 그리기
	lineNumStyle := tcell.StyleDefault.Foreground(tcell.ColorDimGray)
	r.DrawText(1, 1, lineNumStyle, "1")

	// 코드 라인 그리기 (구문 강조 포함)
	line := state.FormattedLine
	quoteStyle := tcell.StyleDefault.Foreground(tcell.ColorOrange)
	keywordStyle := tcell.StyleDefault.Foreground(tcell.ColorCornflowerBlue)
	defaultStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite)

	// 특수 문자를 사용하여 하이라이팅 영역 표시
	highlightedLine := line
	for _, kw := range state.Keywords {
		highlightedLine = strings.ReplaceAll(highlightedLine, kw, "\x1b"+kw+"\x1b")
	}
	highlightedLine = strings.ReplaceAll(highlightedLine, `"`+gs.targetSentence+`"`, "\x1c`\"`"+gs.targetSentence+`\"`+"\x1c")

	x := 4
	currentStyle := defaultStyle
	for _, runeVal := range []rune(highlightedLine) {
		switch runeVal {
		case '\x1b': // 키워드 스타일 토글
			if currentStyle == defaultStyle {
				currentStyle = keywordStyle
			} else {
				currentStyle = defaultStyle
			}
			continue
		case '\x1c': // 문자열 스타일 토글
			if currentStyle == defaultStyle {
				currentStyle = quoteStyle
			} else {
				currentStyle = defaultStyle
			}
			continue
		}
		r.SetContent(x, 1, runeVal, currentStyle)
		x++
	}

	quoteIndex := strings.Index(state.FormattedLine, `"`+gs.targetSentence+`"`)
	if quoteIndex != -1 {
				startX := 4 + quoteIndex
				for i, r := range []rune(gs.userInput) {
					style := tcell.StyleDefault.Foreground(tcell.ColorGreen)
					if i < len([]rune(gs.targetSentence)) && r != []rune(gs.targetSentence)[i] {
						style = tcell.StyleDefault.Foreground(tcell.ColorRed)
					}
					r.SetContent(startX+1+i, 1, []rune(gs.targetSentence)[i], style)
				}
			}
		
			// 상태 표시줄
			statusBarStyle := tcell.StyleDefault.Reverse(true)
			statusText := fmt.Sprintf(" NORMAL | %s | %d/%d ", state.Language, len(gs.userInput), len(gs.targetSentence))
			for i := 0; i < w; i++ {
				r.SetContent(i, h-1, ' ', statusBarStyle)
			}
			r.DrawText(0, h-1, statusBarStyle, statusText)
		
			if gs.isFinished {
				r.HideCursor()
				resultText := fmt.Sprintf("WPM: %.2f | ACC: %.2f%%", gs.wpm, gs.accuracy)
				r.DrawText(len(statusText), h-1, statusBarStyle, " | "+resultText)
			} else {
				if quoteIndex != -1 {
					startX := 4 + quoteIndex
					cursorX := startX + 1 + runewidth.StringWidth(gs.userInput)
					r.ShowCursor(cursorX, 1)
				}	}

	r.Show()
}

func (t *CodeTheme) OnTick(gs *GameState) {}