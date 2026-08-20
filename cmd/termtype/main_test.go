package main

import (
	"testing"

	"github.com/namest504/termtype/internal/chart"
)

func TestChartOptionsFor(t *testing.T) {
	cases := []struct {
		code  string
		style chart.Style
		thick int
	}{
		{"braille1", chart.StyleBraille, 1},
		{"braille2", chart.StyleBraille, 2},
		{"braille3", chart.StyleBraille, 3},
		{"box", chart.StyleBox, 1},
		{"unknown", chart.StyleBraille, 2}, // 알 수 없는 값은 기본값
	}
	for _, c := range cases {
		o := chartOptionsFor(c.code)
		if o.Style != c.style || o.Thickness != c.thick || o.Interp != chart.InterpSmooth {
			t.Fatalf("%s → %+v", c.code, o)
		}
	}
}
