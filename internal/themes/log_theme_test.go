package themes

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/namest504/termtype/internal/domain"
)

var logLineTsRE = regexp.MustCompile(`^\[([^\]]+)\]`)

// TestLogTimestampsSpread verifies that the decorative background log lines
// are not all stamped with the same timestamp: they should read as a stream
// of past events leading up to now, monotonically increasing top to bottom.
func TestLogTimestampsSpread(t *testing.T) {
	theme := &LogTheme{}
	w, h := 100, 30

	gs := &domain.GameState{Sentences: []string{"hello world", "the quick fox", "another pool sentence"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hello world"

	r := newGridRenderer(w, h)
	theme.UpdateScreen(r, gs)

	var timestamps []time.Time
	for y := 0; y < h; y++ {
		line := strings.TrimSpace(string(r.grid[y]))
		m := logLineTsRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		ts, err := time.Parse("2006-01-02T15:04:05Z", m[1])
		if err != nil {
			continue
		}
		timestamps = append(timestamps, ts)
	}

	if len(timestamps) < 2 {
		t.Fatalf("expected multiple background log lines with timestamps, got %d", len(timestamps))
	}

	distinct := map[int64]bool{}
	for _, ts := range timestamps {
		distinct[ts.Unix()] = true
	}
	if len(distinct) <= 1 {
		t.Errorf("expected background log timestamps to differ, got all equal: %v", timestamps)
	}

	for i := 1; i < len(timestamps); i++ {
		if timestamps[i].Before(timestamps[i-1]) {
			t.Errorf("expected timestamps to be monotonically increasing top to bottom, row %d (%v) is before row %d (%v)",
				i, timestamps[i], i-1, timestamps[i-1])
		}
	}
}

// TestLogResultLine verifies the finished screen renders typing-daemon style
// INFO log lines instead of the old "Round finished" DEBUG summary.
func TestLogResultLine(t *testing.T) {
	theme := &LogTheme{}
	w, h := 100, 30

	gs := &domain.GameState{Sentences: []string{"hello world"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hello world"
	gs.UserInput = "hello world"
	gs.IsFinished = true
	gs.WPM = 61.2
	gs.Accuracy = 97.5
	gs.FinalDurS = 12

	r := newGridRenderer(w, h)
	theme.UpdateScreen(r, gs)

	var screen strings.Builder
	for y := 0; y < h; y++ {
		screen.WriteString(strings.TrimRight(string(r.grid[y]), " "))
		screen.WriteString("\n")
	}
	text := screen.String()

	if !strings.Contains(text, "session complete") {
		t.Error("expected the result screen to contain 'session complete'")
	}
	if !strings.Contains(text, "wpm") {
		t.Error("expected the result screen to contain 'wpm'")
	}
	if strings.Contains(text, "Round finished") {
		t.Error("expected the old 'Round finished' DEBUG line to be gone")
	}
	if !strings.Contains(text, "waiting for input") {
		t.Error("expected the result screen to contain 'waiting for input' key guidance")
	}
}
