package ui

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/chart"
)

// TestDrawLineChartAdapts verifies the adapter places chart cells at the
// given origin and styles line vs axis cells differently.
func TestDrawLineChartAdapts(t *testing.T) {
	SetChartOptions(chart.Options{Style: chart.StyleBraille, Interp: chart.InterpLinear, Thickness: 1})
	s := NewMockScreen(30, 10)
	r := NewRenderer(s)
	axis := tcell.StyleDefault.Foreground(tcell.ColorGray)
	line := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	DrawLineChart(r, 2, 1, 20, 5, []float64{0, 50, 100}, axis, line)

	// the tick column sits at x = 2 + 4
	rn, st := s.Cell(6, 1)
	if rn != '┤' || st != axis {
		t.Fatalf("tick cell = %q with wrong style", rn)
	}
	// at least one braille line cell exists right of the axis, line-styled
	found := false
	for y := 1; y < 6; y++ {
		for x := 7; x < 22; x++ {
			if rn, st := s.Cell(x, y); rn >= 0x2800 && rn <= 0x28FF {
				if st != line {
					t.Fatalf("line cell (%d,%d) has axis style", x, y)
				}
				found = true
			}
		}
	}
	if !found {
		t.Fatal("no line cells drawn")
	}
}

func TestSparklineDelegates(t *testing.T) {
	SetASCII(false)
	if got := Sparkline([]float64{0, 100}); got != "▁█" {
		t.Fatalf("got %q", got)
	}
	SetASCII(true)
	defer SetASCII(false)
	if got := Sparkline([]float64{0, 100}); got != ".%" {
		t.Fatalf("ascii got %q", got)
	}
}
