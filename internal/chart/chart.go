// Package chart renders float series as terminal-cell charts. It is pure
// computation: no screen, no styles — callers color cells by Kind.
package chart

// Style selects the drawing technique for line charts.
type Style int

const (
	StyleBraille Style = iota // 2×4 pixels per cell (Unicode braille)
	StyleBox                  // cell-resolution solid line (╭─╮ box drawing)
	StyleASCII                // plain-ASCII fallback (marker per column)
)

// Interp selects how samples are interpolated onto pixel columns.
type Interp int

const (
	InterpLinear Interp = iota
	InterpSmooth        // monotone cubic (Fritsch–Carlson); no overshoot
)

// Kind tags what a cell is part of, so callers can style layers separately.
type Kind int

const (
	KindNone Kind = iota
	KindLine
	KindAxis
	KindLabel
)

// Cell is one terminal cell of a rendered chart.
type Cell struct {
	Rune rune
	Kind Kind
}

// Options configures Render. Thickness (1..3) only affects StyleBraille;
// zero means 1.
type Options struct {
	Width, Height int
	Style         Style
	Interp        Interp
	Thickness     int
}

func bounds(series []float64) (lo, hi float64) {
	lo, hi = series[0], series[0]
	for _, v := range series[1:] {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return lo, hi
}
