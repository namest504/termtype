package themes

import (
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

// gridRenderer is a screen-capturing fake implementing domain.Renderer: it
// keeps the full cell grid so tests can read back what a theme drew. Unlike
// mockRenderer (responsive_test.go), which only reports size, this one
// actually records content. Every string this theme writes is single-byte
// ASCII (hex digits, ".", printable bytes), so a naive one-rune-per-cell
// write is faithful to what the real tcell renderer would do here.
type gridRenderer struct {
	w, h int
	grid [][]rune
}

func newGridRenderer(w, h int) *gridRenderer {
	g := &gridRenderer{w: w, h: h}
	g.grid = make([][]rune, h)
	for y := range g.grid {
		g.grid[y] = make([]rune, w)
	}
	g.Clear()
	return g
}

func (g *gridRenderer) Clear() {
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			g.grid[y][x] = ' '
		}
	}
}
func (g *gridRenderer) DrawText(x, y int, style tcell.Style, text string) {
	for i, r := range []rune(text) {
		g.SetContent(x+i, y, r, style)
	}
}
func (g *gridRenderer) DrawRune(x, y int, r rune, style tcell.Style) int {
	g.SetContent(x, y, r, style)
	return 1
}
func (g *gridRenderer) Show()            {}
func (g *gridRenderer) Size() (int, int) { return g.w, g.h }
func (g *gridRenderer) SetContent(x, y int, r rune, style tcell.Style) {
	if x >= 0 && x < g.w && y >= 0 && y < g.h {
		g.grid[y][x] = r
	}
}
func (g *gridRenderer) HideCursor()         {}
func (g *gridRenderer) ShowCursor(x, y int) {}

// TestHexResultEncodesStats verifies the result screen encodes the round's
// stats as a real dump row right below the target, instead of overlapping an
// existing background row: (1) the stat text appears in the ascii gutter,
// (2) the hex columns on that row parse back to the stat string's UTF-8
// bytes, and (3) nothing bleeds through past the stat content on that row —
// i.e. the row was cleared, not drawn over.
func TestHexResultEncodesStats(t *testing.T) {
	theme := &HexTheme{}
	w, h := 100, 30

	gs := &domain.GameState{Sentences: []string{"hi"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hi" // short + deterministic: exactly one target dump row
	gs.IsFinished = true
	gs.WPM = 61.2
	gs.Accuracy = 97.5
	gs.FinalDurS = 12

	r := newGridRenderer(w, h)
	theme.UpdateScreen(r, gs)

	stats := []byte(ui.ResultText(gs))
	targetRows := (len([]byte(gs.TargetSentence)) + 15) / 16
	startRow := h/2 + targetRows + 1 // StartLine (h/2) + target rows + one blank row
	if startRow >= h {
		t.Fatalf("test setup: expected result row %d falls off screen height %d", startRow, h)
	}
	chunk := stats[:min(len(stats), 16)]

	// (1) stat text appears in the ascii gutter (x=62..77)
	asciiLine := string(r.grid[startRow][62:78])
	if !strings.Contains(asciiLine, "wpm") {
		t.Fatalf("ascii gutter row %d = %q, want it to contain \"wpm\"", startRow, asciiLine)
	}

	// (2) hex columns (x=10..) parse back to the stat string's UTF-8 bytes
	for i, want := range chunk {
		hi, lo := r.grid[startRow][10+i*3], r.grid[startRow][10+i*3+1]
		got, err := strconv.ParseUint(string(hi)+string(lo), 16, 8)
		if err != nil {
			t.Fatalf("row %d byte %d: hex cell %q%q does not parse: %v", startRow, i, hi, lo, err)
		}
		if byte(got) != want {
			t.Errorf("row %d byte %d = %#02x, want %#02x (stats %q)", startRow, i, byte(got), want, string(chunk))
		}
	}

	// ascii gutter should be the matching byte-for-byte translation too
	// (printable ASCII as itself, everything else -- including multibyte
	// UTF-8 continuation bytes -- as '.').
	wantAscii := make([]rune, len(chunk))
	for i, b := range chunk {
		if b >= 32 && b <= 126 {
			wantAscii[i] = rune(b)
		} else {
			wantAscii[i] = '.'
		}
	}
	if got := string(r.grid[startRow][62 : 62+len(chunk)]); got != string(wantAscii) {
		t.Errorf("row %d ascii gutter = %q, want %q", startRow, got, string(wantAscii))
	}

	// (3) no leftover background bytes after the stat content on this row
	for x := 62 + len(chunk); x < 78; x++ {
		if r.grid[startRow][x] != ' ' {
			t.Errorf("row %d ascii cell x=%d = %q, want cleared space (background bleed-through)", startRow, x, r.grid[startRow][x])
		}
	}
	for x := 10 + len(chunk)*3; x < 10+16*3; x++ {
		if r.grid[startRow][x] != ' ' {
			t.Errorf("row %d hex cell x=%d = %q, want cleared space (background bleed-through)", startRow, x, r.grid[startRow][x])
		}
	}
}

func TestHexWindow(t *testing.T) {
	cases := []struct {
		name               string
		targetLen, itLen   int
		avail              int
		wantStart, wantVis int
	}{
		{"short target fits", 40, 0, 10, 0, 3},        // 3 rows, all visible
		{"start of long target", 400, 0, 5, 0, 5},     // 25 rows, window at top
		{"middle of long target", 400, 200, 5, 10, 5}, // cursor row 12 centered
		{"end of long target", 400, 399, 5, 20, 5},    // pinned to the last rows
		{"input past target end", 400, 500, 5, 20, 5}, // clamped
		{"tiny terminal", 400, 0, 0, 0, 1},            // at least one row
		{"empty target", 0, 0, 5, 0, 1},               // never zero rows
	}
	for _, c := range cases {
		start, vis := hexWindow(c.targetLen, c.itLen, c.avail)
		if start != c.wantStart || vis != c.wantVis {
			t.Errorf("%s: hexWindow(%d, %d, %d) = (%d, %d), want (%d, %d)",
				c.name, c.targetLen, c.itLen, c.avail, start, vis, c.wantStart, c.wantVis)
		}
	}
}
