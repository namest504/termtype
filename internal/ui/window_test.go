package ui

import (
	"strings"
	"testing"

	"github.com/namest504/termtype/internal/domain"
)

func TestWindowStart(t *testing.T) {
	cases := []struct {
		total, cursor, visible, want int
	}{
		{1, 0, 3, 0},   // shorter than the window
		{3, 2, 3, 0},   // exactly the window
		{10, 0, 3, 0},  // start of the text
		{10, 5, 3, 4},  // cursor held on the middle row
		{10, 9, 3, 7},  // end pins the window
		{10, 20, 3, 7}, // cursor past the end stays clamped
		{20, 10, 7, 7}, // wider windows center too
	}
	for _, c := range cases {
		if got := WindowStart(c.total, c.cursor, c.visible); got != c.want {
			t.Errorf("WindowStart(%d, %d, %d) = %d, want %d", c.total, c.cursor, c.visible, got, c.want)
		}
	}
}

func TestLineOfRune(t *testing.T) {
	wrapped := WrapText("one two three four five six seven eight nine ten", 10)
	if got := LineOfRune(wrapped, 0); got != 0 {
		t.Errorf("rune 0 on line %d, want 0", got)
	}
	total := 0
	for _, l := range wrapped {
		total += len([]rune(l))
	}
	if got := LineOfRune(wrapped, total+3); got != len(wrapped)-1 {
		t.Errorf("rune past the end on line %d, want %d", got, len(wrapped)-1)
	}
	if got := LineOfRune(nil, 5); got != 0 {
		t.Errorf("empty wrap gave line %d, want 0", got)
	}
}

func TestDrawWindowsLongTargets(t *testing.T) {
	mock := NewMockScreen(30, 12)
	renderer := NewRenderer(mock)
	gs := &domain.GameState{
		TargetSentence: strings.TrimSpace(strings.Repeat("word ", 60)),
		UserInput:      "",
	}

	tr := &TypingRenderer{}
	rows := tr.Draw(renderer, gs, TypingRendererOptions{
		StartY: 0, Width: 30, MaxLines: 3,
	})
	if rows != 3 {
		t.Fatalf("Draw returned %d rows, want the 3-line window", rows)
	}
	for y := 3; y < 12; y++ {
		for x, ch := range mock.cells[y] {
			if ch != ' ' && ch != 0 {
				t.Fatalf("content %q at (%d,%d) below the window", ch, x, y)
			}
		}
	}
}

func TestDrawReturnsAllRowsWithoutCap(t *testing.T) {
	mock := NewMockScreen(30, 40)
	renderer := NewRenderer(mock)
	target := strings.TrimSpace(strings.Repeat("word ", 20))
	gs := &domain.GameState{TargetSentence: target}

	tr := &TypingRenderer{}
	rows := tr.Draw(renderer, gs, TypingRendererOptions{StartY: 0, Width: 30})
	if want := len(WrapText(target, 27)); rows != want {
		t.Errorf("Draw returned %d rows, want every wrapped line (%d)", rows, want)
	}
}
