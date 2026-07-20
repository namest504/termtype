package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/domain"
)

func totalRunes(lines []string) int {
	n := 0
	for _, l := range lines {
		n += len([]rune(l))
	}
	return n
}

// BUG 4 regression: when a new line's first word width exactly equals the width, no blank " " line should appear.
func TestWrapText_NoSpuriousBlankLine(t *testing.T) {
	got := WrapText("abc", 3)
	if len(got) != 1 || got[0] != "abc" {
		t.Fatalf("WrapText(\"abc\", 3) = %q, want [\"abc\"]", got)
	}
}

// A word longer than the width must be placed on its own line with no blank line.
func TestWrapText_LongWordOnOwnLine(t *testing.T) {
	got := WrapText("abcdefgh", 3)
	if len(got) != 1 || got[0] != "abcdefgh" {
		t.Fatalf("WrapText long word = %q, want [\"abcdefgh\"]", got)
	}
}

// Invariant that TypingRenderer relies on: total runes across wrapped lines == original rune count.
// (For single-space-separated sentences.) If broken, character coloring/cursor offsets get misaligned.
func TestWrapText_OffsetAlignmentInvariant(t *testing.T) {
	sentences := []string{
		"the quick brown fox jumps over the lazy dog",
		"hello world foo bar baz",
		"a b c d e f g h i j k l m n o p",
		"aaa bbb ccc ddd eee fff ggg",
	}
	for _, s := range sentences {
		want := len([]rune(s))
		for w := 3; w <= 60; w++ {
			lines := WrapText(s, w)
			if got := totalRunes(lines); got != want {
				t.Errorf("WrapText(%q, %d): total runes %d, want %d (lines=%q)", s, w, got, want, lines)
				break
			}
		}
	}
}

// Mock screen that records the cursor position.
type cursorMock struct {
	tcell.Screen
	w, h       int
	curX, curY int
}

func (m *cursorMock) SetContent(x, y int, mainc rune, combc []rune, style tcell.Style) {}
func (m *cursorMock) Size() (int, int)                                                 { return m.w, m.h }
func (m *cursorMock) ShowCursor(x, y int)                                              { m.curX, m.curY = x, y }
func (m *cursorMock) HideCursor()                                                      {}

// BUG 2 regression: when input reaches the wrap boundary, the cursor should be at the start of the next line, not the end of the previous line.
func TestTypingRenderer_CursorAtLineBreak(t *testing.T) {
	mock := &cursorMock{w: 40, h: 10}
	renderer := NewRenderer(mock)

	// availableWidth = Width-3 = 9 → ["aaaa bbbb ", "cccc"] (line0 length 10)
	gs := &domain.GameState{
		TargetSentence: "aaaa bbbb cccc",
		UserInput:      "aaaa bbbb ", // input line0 exactly including the trailing space
	}
	tr := &TypingRenderer{}
	tr.Draw(renderer, gs, TypingRendererOptions{StartY: 0, Width: 12, PrefixWidth: 0, CenterText: false})

	if mock.curY != 1 {
		t.Errorf("cursorY = %d, want 1 (start of next line)", mock.curY)
	}
	if mock.curX != 1 {
		t.Errorf("cursorX = %d, want 1 (line start column)", mock.curX)
	}
}
