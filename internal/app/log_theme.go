package app

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var logLevels = []string{"INFO", "WARN", "DEBUG", "ERROR"}
var sources = []string{"auth-service", "api-gateway", "db-connector", "cache-worker", "metrics-agent"}

func formatAsLogLine(sentence string) (string, string, string) {
	ts := time.Now().Format("2006-01-02T15:04:05Z")
	level := logLevels[rand.Intn(len(logLevels))]
	source := sources[rand.Intn(len(sources))]
	prefix := fmt.Sprintf("[%s] [%s] [%s] ", ts, level, source)
	return prefix + sentence, prefix, sentence
}

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

func (t *LogTheme) UpdateScreen(renderer *Renderer, gs *GameState) {
	logState, ok := gs.CustomState.(*LogThemeState)
	if !ok {
		return // 상태가 준비되지 않음
	}
	renderer.Clear()
	w, h := renderer.Size()

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
		renderer.DrawText(1, logY, tcell.StyleDefault.Foreground(tcell.ColorDimGray), logLine)
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
		renderer.DrawText(1, targetY, prefixStyle, logState.logPrefix)

		prefixWidth := runewidth.StringWidth(logState.logPrefix)
		availableWidth := w - 1 - prefixWidth

		wrappedTarget := wrapText(gs.targetSentence, availableWidth)
		inputRunes := []rune(gs.userInput)
		inputOffset := 0

		for lineIdx, line := range wrappedTarget {
			lineRunes := []rune(line)
			for charIdx, r := range lineRunes {
				currentInputIdx := inputOffset + charIdx
				style := tcell.StyleDefault.Foreground(tcell.ColorGray)

				if currentInputIdx < len(inputRunes) {
					if inputRunes[currentInputIdx] == r {
						style = tcell.StyleDefault.Foreground(tcell.ColorGreen)
					} else {
						style = tcell.StyleDefault.Foreground(tcell.ColorRed)
					}
				}
				renderer.SetContent(1+prefixWidth+charIdx, targetY+lineIdx, r, style)
			}
			inputOffset += len(lineRunes)
		}

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
		renderer.ShowCursor(1+prefixWidth+cursorCharIdx, targetY+cursorLineIdx)

	} else {
		renderer.HideCursor()
		renderer.DrawText(1, targetY, tcell.StyleDefault.Foreground(tcell.ColorDimGray), logState.targetLogLine)

		resultLog := fmt.Sprintf("[%s] [DEBUG] [metrics-agent] Round finished. WPM: %.2f, Accuracy: %.2f%%", time.Now().Format("2006-01-02T15:04:05Z"), gs.wpm, gs.accuracy)
		renderer.DrawText(1, targetY+1, getStyleForLogLevel("DEBUG"), resultLog)

		guideText := "Press Enter to continue or ESC to exit."
		renderer.DrawText(1, targetY+3, tcell.StyleDefault, guideText)
	}

	renderer.Show()
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

func getStyleForLogLevel(level string) tcell.Style {
	switch level {
	case "ERROR":
		return tcell.StyleDefault.Foreground(tcell.ColorRed)
	case "WARN":
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	case "DEBUG":
		return tcell.StyleDefault.Foreground(tcell.ColorBlue)
	default:
		return tcell.StyleDefault.Foreground(tcell.ColorWhite)
	}
}
