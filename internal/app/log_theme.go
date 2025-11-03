package app

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func init() {
	Themes["log"] = &LogTheme{}
}

// --- Log Theme --- //

type LogTheme struct{}

type LogThemeState struct {
	targetLogLine  string
	logPrefix      string
	backgroundLogs []string
}

func (t *LogTheme) ResetState(gs *GameState) {
	gs.resetCommon()

	// 테마에 맞는 커스텀 상태 초기화
	logState := &LogThemeState{
		backgroundLogs: make([]string, 0), // 처음에는 비워둠
	}
	gs.CustomState = logState

	selectedSentence := gs.sentences[rand.Intn(len(gs.sentences))]
	fullLog, prefix, sentence := formatAsLogLine(selectedSentence)

	logState.targetLogLine = fullLog
	logState.logPrefix = prefix
	gs.targetSentence = sentence
}

func (t *LogTheme) UpdateScreen(s tcell.Screen, gs *GameState) {
	logState, ok := gs.CustomState.(*LogThemeState)
	if !ok {
		return // 상태가 준비되지 않음
	}
	s.Clear()
	_, h := s.Size()

	// 터미널 높이에 맞춰 동적으로 로그 줄 수 조절
	numLogs := h - 4 // 상단 여백, 타겟 라인, 결과 라인 등을 위한 공간 확보
	if numLogs < 0 {
		numLogs = 0
	}
	for len(logState.backgroundLogs) < numLogs {
		newLog, _, _ := formatAsLogLine(gs.sentences[rand.Intn(len(gs.sentences))])
		logState.backgroundLogs = append([]string{newLog}, logState.backgroundLogs...)
	}
	if len(logState.backgroundLogs) > numLogs {
		logState.backgroundLogs = logState.backgroundLogs[len(logState.backgroundLogs)-numLogs:]
	}

	// 배경 로그 그리기
	logY := 0
	for _, logLine := range logState.backgroundLogs {
		drawText(s, 1, logY, tcell.StyleDefault.Foreground(tcell.ColorDimGray), logLine)
		logY++
	}

	targetY := logY + 1
	if !gs.isFinished {
		prefixStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
		if strings.Contains(logState.logPrefix, "[ERROR]") {
			prefixStyle = tcell.StyleDefault.Foreground(tcell.ColorRed)
		} else if strings.Contains(logState.logPrefix, "[WARN]") {
			prefixStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
		}
		drawText(s, 1, targetY, prefixStyle, logState.logPrefix)

		prefixWidth := runewidth.StringWidth(logState.logPrefix)
		targetRunes := []rune(gs.targetSentence)
		inputRunes := []rune(gs.userInput)

		for i, r := range targetRunes {
			style := tcell.StyleDefault.Foreground(tcell.ColorGray)
			if i < len(inputRunes) {
				if inputRunes[i] == r {
					style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
				} else {
					style = tcell.StyleDefault.Foreground(tcell.ColorRed)
				}
			}
			s.SetContent(1+prefixWidth+i, targetY, r, nil, style)
		}

		cursorX := 1 + prefixWidth + runewidth.StringWidth(gs.userInput)
		s.ShowCursor(cursorX, targetY)

	} else {
		s.HideCursor()
		drawText(s, 1, targetY, tcell.StyleDefault.Foreground(tcell.ColorDimGray), logState.targetLogLine)

		resultLog := fmt.Sprintf("[%s] [DEBUG] [metrics-agent] Round finished. WPM: %.2f, Accuracy: %.2f%%", time.Now().Format("2006-01-02T15:04:05Z"), gs.wpm, gs.accuracy)
		drawText(s, 1, targetY+1, getStyleForLogLevel("DEBUG"), resultLog)

		guideText := "Press Enter to continue or ESC to exit."
		drawText(s, 1, targetY+3, tcell.StyleDefault, guideText)
	}

	s.Show()
}

// OnTick은 LogTheme에 실시간 스크롤 효과를 줍니다.
func (t *LogTheme) OnTick(gs *GameState) {
	logState, ok := gs.CustomState.(*LogThemeState)
	if !ok {
		return
	}

	// 새 로그를 추가하고 가장 오래된 로그를 제거
	newLog, _, _ := formatAsLogLine(gs.sentences[rand.Intn(len(gs.sentences))])
	logState.backgroundLogs = append(logState.backgroundLogs[1:], newLog)
}
