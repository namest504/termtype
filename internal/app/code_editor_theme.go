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

func (t *CodeTheme) UpdateScreen(s tcell.Screen, gs *GameState) {
}

func (t *CodeTheme) OnTick(gs *GameState) {}