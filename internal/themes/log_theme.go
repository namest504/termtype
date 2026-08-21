package themes

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

var logLevels = []string{"INFO", "WARN", "DEBUG", "ERROR"}
var sources = []string{"auth-service", "api-gateway", "db-connector", "cache-worker", "metrics-agent"}

func formatAsLogLine(ts time.Time, sentence string) (string, string, string) {
	tsStr := ts.Format("2006-01-02T15:04:05Z")
	level := logLevels[rand.Intn(len(logLevels))]
	source := sources[rand.Intn(len(sources))]
	prefix := fmt.Sprintf("[%s] [%s] [%s] ", tsStr, level, source)
	return prefix + sentence, prefix, sentence
}

func init() {
	Themes["log"] = &LogTheme{}
}

// LogTheme mimics a streaming application log: the target text appears as
// the line currently being typed, framed by decorative background log lines
// above and below it.
type LogTheme struct{}

// LogThemeState is LogTheme's per-round state: the formatted target log
// line and the decorative background lines around it.
type LogThemeState struct {
	targetLogLine  string
	logPrefix      string
	backgroundLogs []string
}

func (t *LogTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()

	// Initialize the custom state for this theme.
	logState := &LogThemeState{
		backgroundLogs: make([]string, 0), // empty to start
	}
	gs.CustomState = logState

	selectedSentence := gs.RandomSentence()
	fullLog, prefix, sentence := formatAsLogLine(time.Now(), selectedSentence)

	logState.targetLogLine = fullLog
	logState.logPrefix = prefix
	gs.TargetSentence = sentence
}

func (t *LogTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	logState, ok := gs.CustomState.(*LogThemeState)
	if !ok {
		return // state is not ready
	}
	renderer.Clear()
	w, h := renderer.Size()

	t.drawBackgroundLogs(renderer, gs, logState, h)

	targetY := t.calculateTargetY(h)
	if !gs.IsFinished {
		t.drawTypingLine(renderer, gs, logState, w, targetY)
	} else {
		t.drawResultLine(renderer, gs, logState, targetY)
	}

	renderer.Show()
}

func (t *LogTheme) calculateTargetY(h int) int {
	numLogs := h - 4
	if numLogs < 0 {
		numLogs = 0
	}
	return numLogs + 1
}

func (t *LogTheme) drawBackgroundLogs(renderer domain.Renderer, gs *domain.GameState, logState *LogThemeState, h int) {
	// Dynamically adjust the number of log lines to the terminal height.
	numLogs := h - 4 // reserve space for top margin, target line, result line, etc.
	if numLogs < 0 {
		numLogs = 0
	}
	if needed := numLogs - len(logState.backgroundLogs); needed > 0 {
		now := time.Now()
		newLines := make([]string, needed)
		// Stagger each newly generated line into the past, oldest at the top
		// and closest to now at the bottom, so the log reads as a stream of
		// events rather than a wall of identical timestamps. Offsets
		// accumulate from the bottom up so timestamps stay monotonically
		// increasing top to bottom.
		var offset time.Duration
		for i := needed - 1; i >= 0; i-- {
			offset += time.Duration(3+rand.Intn(20)) * time.Second
			newLog, _, _ := formatAsLogLine(now.Add(-offset), gs.RandomPoolSentence())
			newLines[i] = newLog
		}
		logState.backgroundLogs = append(newLines, logState.backgroundLogs...)
	}
	if len(logState.backgroundLogs) > numLogs {
		logState.backgroundLogs = logState.backgroundLogs[len(logState.backgroundLogs)-numLogs:]
	}

	// Draw the background logs.
	logY := 0
	for _, logLine := range logState.backgroundLogs {
		// Prevent drawing past the available area.
		if logY >= numLogs {
			break
		}
		renderer.DrawText(1, logY, tcell.StyleDefault.Foreground(tcell.ColorDimGray), logLine)
		logY++
	}
}

func (t *LogTheme) drawTypingLine(renderer domain.Renderer, gs *domain.GameState, logState *LogThemeState, w, targetY int) {
	prefixStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	if strings.Contains(logState.logPrefix, "[ERROR]") {
		prefixStyle = tcell.StyleDefault.Foreground(tcell.ColorRed)
	} else if strings.Contains(logState.logPrefix, "[WARN]") {
		prefixStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	renderer.DrawText(1, targetY, prefixStyle, logState.logPrefix)

	prefixWidth := runewidth.StringWidth(logState.logPrefix)

	// The typing line sits near the bottom, so long targets (the words
	// stream) scroll within the rows that remain below it.
	_, h := renderer.Size()
	maxLines := h - targetY
	if maxLines < 1 {
		maxLines = 1
	}
	tr := &ui.TypingRenderer{}
	tr.Draw(renderer, gs, ui.TypingRendererOptions{
		StartY:      targetY,
		Width:       w - 1, // full width minus the left margin of 1 (PrefixWidth is handled internally)
		PrefixWidth: prefixWidth,
		CenterText:  false,
		MaxLines:    maxLines,
	})
}

func (t *LogTheme) drawResultLine(renderer domain.Renderer, gs *domain.GameState, logState *LogThemeState, targetY int) {
	renderer.HideCursor()
	renderer.DrawText(1, targetY, tcell.StyleDefault.Foreground(tcell.ColorDimGray), logState.targetLogLine)

	now := time.Now().Format("2006-01-02T15:04:05Z")
	line1 := fmt.Sprintf("[%s] [INFO] [typing-daemon] session complete — %.0f wpm, %.1f%% acc, %.0fs",
		now, gs.WPM, gs.Accuracy, gs.FinalDurS)
	line2 := fmt.Sprintf("[%s] [INFO] [typing-daemon] process exited (0)", now)
	renderer.DrawText(1, targetY+1, getStyleForLogLevel("INFO"), line1)
	renderer.DrawText(1, targetY+2, getStyleForLogLevel("INFO"), line2)
}

// OnTick gives the LogTheme a real-time scrolling effect.
func (t *LogTheme) OnTick(gs *domain.GameState) {
	logState, ok := gs.CustomState.(*LogThemeState)
	if !ok {
		return
	}

	// If there are no background logs (terminal height <= 4), there is nothing to scroll.
	if len(logState.backgroundLogs) == 0 {
		return
	}

	// Append a new log (stamped with the current time) and remove the oldest one.
	newLog, _, _ := formatAsLogLine(time.Now(), gs.RandomPoolSentence())
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
