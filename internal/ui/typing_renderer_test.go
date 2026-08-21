package ui

import (
	"github.com/namest504/termtype/internal/domain"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// mockCell holds the rune and style last written to one screen position.
type mockCell struct {
	r     rune
	style tcell.Style
}

// MockScreen is a mock implementation of tcell.Screen for testing
type MockScreen struct {
	tcell.Screen
	cells map[int]map[int]mockCell
	w, h  int
}

func NewMockScreen(w, h int) *MockScreen {
	return &MockScreen{
		cells: make(map[int]map[int]mockCell),
		w:     w,
		h:     h,
	}
}

func (m *MockScreen) SetContent(x, y int, mainc rune, combc []rune, style tcell.Style) {
	if m.cells[y] == nil {
		m.cells[y] = make(map[int]mockCell)
	}
	m.cells[y][x] = mockCell{r: mainc, style: style}
}

// Cell returns the rune and style last written at (x, y). A position never
// written returns the zero rune and tcell.StyleDefault.
func (m *MockScreen) Cell(x, y int) (rune, tcell.Style) {
	if row, ok := m.cells[y]; ok {
		if c, ok := row[x]; ok {
			return c.r, c.style
		}
	}
	return 0, tcell.StyleDefault
}

func (m *MockScreen) Size() (int, int) {
	return m.w, m.h
}

func (m *MockScreen) ShowCursor(x, y int) {
	// Do nothing
}

func TestTypingRenderer_Draw_Padding(t *testing.T) {
	width := 50
	height := 10
	mockScreen := NewMockScreen(width, height)
	renderer := NewRenderer(mockScreen)

	tr := &TypingRenderer{}
	gs := &domain.GameState{
		TargetSentence: "This is a very long sentence that should be wrapped properly and have padding at the end.",
		UserInput:      "",
	}

	opts := TypingRendererOptions{
		StartY:      0,
		Width:       width,
		PrefixWidth: 0,
		CenterText:  false, // This should trigger the padding logic
	}

	tr.Draw(renderer, gs, opts)

	// Check if any content is written to the last 3 columns (width-1, width-2, width-3)
	// Since we subtract 3, the max X index used should be width - 3 - 1 (0-indexed) = width - 4
	// So columns width-1, width-2, width-3 should be empty.
	// Wait, if availableWidth is width-3, wrapText uses that width.
	// So text can go up to width-3.
	// 0-indexed: 0 to width-4.
	// So width-3, width-2, width-1 should be empty?
	// Let's check the code. wrapText logic:
	// if runewidth.StringWidth(currentLine+" "+word) <= width
	// So max width is 'width'.
	// If we pass width-3, max width is width-3.
	// So characters can be at x=0 to x=width-4.
	// x=width-3 should be empty.

	for y := 0; y < height; y++ {
		if row, ok := mockScreen.cells[y]; ok {
			for x := width - 3; x < width; x++ {
				if c, exists := row[x]; exists && c.r != ' ' && c.r != 0 {
					t.Errorf("Found character '%c' at (%d, %d), expected padding", c.r, x, y)
				}
			}
		}
	}
}
