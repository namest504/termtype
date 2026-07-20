package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
)

// colOf returns the x column where rune target first appears on row y, or -1.
func colOf(ss tcell.SimulationScreen, y int, target rune) int {
	cells, w, _ := ss.GetContents()
	for x := 0; x < w; x++ {
		c := cells[y*w+x]
		if len(c.Runes) > 0 && c.Runes[0] == target {
			return x
		}
	}
	return -1
}

// Hangul syllables are two cells wide, so the TypingRenderer must advance by
// display width — each successive syllable lands two columns further right, not
// one. This guards against the rune-index drawing bug.
func TestTypingRenderer_HangulPlacedByDisplayWidth(t *testing.T) {
	ss := tcell.NewSimulationScreen("")
	if err := ss.Init(); err != nil {
		t.Fatal(err)
	}
	ss.SetSize(40, 6)

	gs := &domain.GameState{TargetSentence: "가나다"}
	tr := &TypingRenderer{}
	tr.Draw(NewRenderer(ss), gs, TypingRendererOptions{StartY: 0, Width: 40, PrefixWidth: 0, CenterText: false})
	ss.Show()

	// startX = 1 + PrefixWidth(0) = 1, then +2 per wide rune.
	for syllable, want := range map[rune]int{'가': 1, '나': 3, '다': 5} {
		if got := colOf(ss, 0, syllable); got != want {
			t.Errorf("%q drawn at column %d, want %d", syllable, got, want)
		}
	}
}

// WrapText must wrap Korean by display width, not rune count: a line of N
// width-2 syllables fills 2N cells.
func TestWrapText_KoreanByDisplayWidth(t *testing.T) {
	// Six syllables (display width 12) into a width-8 box → wrap after the
	// widest run of syllables that fits.
	lines := WrapText("가 나 다 라 마 바", 8)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %d line(s): %q", len(lines), lines)
	}
	for _, line := range lines {
		// WrapText keeps one trailing space at the wrap point by design; the
		// content itself must still fit the width by display measure.
		content := strings.TrimRight(line, " ")
		if runewidth.StringWidth(content) > 8 {
			t.Errorf("line content %q exceeds width 8 (display width %d)", content, runewidth.StringWidth(content))
		}
	}
}

// The bundled Korean pool should be present and well-formed.
func TestKoreanSentences_NonEmpty(t *testing.T) {
	if len(domain.KoreanSentences) == 0 {
		t.Fatal("KoreanSentences is empty")
	}
	for _, s := range domain.KoreanSentences {
		if s == "" {
			t.Error("KoreanSentences contains an empty string")
		}
	}
}
