// Package chart renders float series as terminal-cell charts. It is pure
// computation: no screen, no styles — callers color cells by Kind.
package chart

import "fmt"

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

// brailleBits holds the braille dot bit for the pixel at (px, py) inside one
// cell: 2 pixel columns × 4 pixel rows per character cell.
var brailleBits = [4][2]int{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}

// labelW is the y-axis label gutter: "999 " right-aligned, then the tick.
const labelW = 4

// sample maps series onto n points. InterpLinear draws straight segments
// between samples; InterpSmooth uses monotone cubic interpolation.
func sample(series []float64, n int, ip Interp) []float64 {
	if ip == InterpSmooth && len(series) >= 3 {
		return monotoneCubic(series, n)
	}
	out := make([]float64, n)
	for c := 0; c < n; c++ {
		pos := float64(c) * float64(len(series)-1) / float64(n-1)
		i := int(pos)
		v := series[i]
		if frac := pos - float64(i); i+1 < len(series) {
			v = series[i]*(1-frac) + series[i+1]*frac
		}
		out[c] = v
	}
	return out
}

// pixelRows maps sampled values onto pixel rows using the given lo/hi
// bounds (the TRUE series bounds, not the sampled array's own bounds — the
// caller must scale consistently with whatever bounds the axis labels
// use). Row 0 is the top (max). A flat series (hi == lo) sits on the
// middle row.
func pixelRows(values []float64, pxRows int, lo, hi float64) []int {
	out := make([]int, len(values))
	for i, v := range values {
		if hi == lo {
			out[i] = pxRows / 2
			continue
		}
		out[i] = int((hi - v) / (hi - lo) * float64(pxRows-1))
	}
	return out
}

// Render draws series as a line chart with a labeled y-axis into a
// Height×Width cell grid. Fewer than two samples, or a rect too small to
// hold the axis, returns nil.
func Render(series []float64, o Options) [][]Cell {
	cols := o.Width - labelW - 1
	if len(series) < 2 || cols < 2 || o.Height < 2 {
		return nil
	}
	grid := make([][]Cell, o.Height)
	for i := range grid {
		grid[i] = make([]Cell, o.Width)
		for j := range grid[i] {
			grid[i][j] = Cell{Rune: ' '}
		}
	}

	lo, hi := bounds(series)
	tick := '┤'
	if o.Style == StyleASCII {
		tick = '|'
	}
	for row := 0; row < o.Height; row++ {
		v := hi
		if hi != lo {
			v = hi - (hi-lo)*float64(row)/float64(o.Height-1)
		}
		if row == 0 || row == o.Height-1 || row == (o.Height-1)/2 {
			label := fmt.Sprintf("%*.0f", labelW-1, v)
			for i, r := range []rune(label) {
				grid[row][i] = Cell{Rune: r, Kind: KindLabel}
			}
		}
		grid[row][labelW] = Cell{Rune: tick, Kind: KindAxis}
	}

	switch o.Style {
	case StyleASCII:
		renderASCII(grid, series, cols, o.Height)
	default:
		renderBraille(grid, series, cols, o, lo, hi)
	}
	return grid
}

// renderBraille plots into a 2×4-per-cell pixel grid, joining neighbor
// pixel columns vertically so steep segments stay connected. lo, hi are
// the TRUE series bounds (matching the axis labels), NOT the bounds of the
// downsampled array — a narrow peak can fall between sample points and be
// smoothed away, so scaling against the sample's own bounds would diverge
// from what the axis reports.
func renderBraille(grid [][]Cell, series []float64, cols int, o Options, lo, hi float64) {
	pxRows := pixelRows(sample(series, cols*2, o.Interp), o.Height*4, lo, hi)
	masks := make([][]int, o.Height)
	for i := range masks {
		masks[i] = make([]int, cols)
	}
	prev := pxRows[0]
	for c, row := range pxRows {
		lo, hi := row, row
		if c > 0 {
			if prev < row {
				lo = prev + 1
			} else if prev > row {
				hi = prev - 1
			}
		}
		for py := lo; py <= hi; py++ {
			masks[py/4][c/2] |= brailleBits[py%4][c%2]
		}
		prev = row
	}
	for cy := range masks {
		for cx, mask := range masks[cy] {
			if mask != 0 {
				grid[cy][labelW+1+cx] = Cell{Rune: rune(0x2800 + mask), Kind: KindLine}
			}
		}
	}
}

// renderASCII is the --ascii fallback: a marker per cell column with
// vertical fill between steep neighbors. Samples are picked per column
// without interpolation, matching the pre-refactor behavior.
func renderASCII(grid [][]Cell, series []float64, cols, height int) {
	rows := make([]int, cols)
	lo, hi := bounds(series)
	for c := 0; c < cols; c++ {
		i := c * (len(series) - 1) / (cols - 1)
		if hi == lo {
			rows[c] = height / 2
			continue
		}
		rows[c] = int((hi - series[i]) / (hi - lo) * float64(height-1))
	}
	for c, row := range rows {
		grid[row][labelW+1+c] = Cell{Rune: '*', Kind: KindLine}
		if c > 0 {
			a, b := rows[c-1], row
			if a > b {
				a, b = b, a
			}
			for between := a + 1; between < b; between++ {
				grid[between][labelW+1+c] = Cell{Rune: '|', Kind: KindLine}
			}
		}
	}
}
