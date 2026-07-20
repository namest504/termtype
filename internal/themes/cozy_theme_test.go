package themes

import (
	"testing"

	"github.com/namest504/termtype/internal/ui"
)

func TestCozyRegistered(t *testing.T) {
	if _, ok := Themes["cozy"]; !ok {
		t.Fatal("cozy theme is not registered")
	}
}

func TestCozyWindowStart(t *testing.T) {
	cases := []struct {
		total, cursor, want int
	}{
		{1, 0, 0},  // shorter than the window
		{3, 2, 0},  // exactly the window
		{10, 0, 0}, // start of the text
		{10, 1, 0},
		{10, 5, 4},  // cursor centered on the middle row
		{10, 9, 7},  // end of the text: window pins to the last lines
		{10, 20, 7}, // cursor line past the end stays clamped
	}
	for _, c := range cases {
		if got := cozyWindowStart(c.total, c.cursor); got != c.want {
			t.Errorf("cozyWindowStart(%d, %d) = %d, want %d", c.total, c.cursor, got, c.want)
		}
	}
}

func TestCursorLineOf(t *testing.T) {
	wrapped := ui.WrapText("one two three four five six seven eight", 10)
	if got := cursorLineOf(wrapped, 0); got != 0 {
		t.Errorf("cursor at 0 on line %d, want 0", got)
	}
	total := 0
	for _, l := range wrapped {
		total += len([]rune(l))
	}
	if got := cursorLineOf(wrapped, total+5); got != len(wrapped)-1 {
		t.Errorf("cursor past the end on line %d, want last line %d", got, len(wrapped)-1)
	}
}

func TestCozyColumnWidth(t *testing.T) {
	cases := []struct{ w, want int }{
		{120, 60}, // capped at the reading measure
		{40, 34},
		{20, 14},
		{5, 1}, // never below one column
	}
	for _, c := range cases {
		if got := cozyColumnWidth(c.w); got != c.want {
			t.Errorf("cozyColumnWidth(%d) = %d, want %d", c.w, got, c.want)
		}
	}
}
