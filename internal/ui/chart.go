package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
)

// chartRows maps a series onto a cols×rows grid: one row index per column,
// row 0 at the top holding the maximum. The series is stretched or
// compressed to fill the columns. A flat series sits on the middle row.
func chartRows(series []float64, cols, rows int) []int {
	if len(series) < 2 || cols < 1 || rows < 1 {
		return nil
	}
	min, max := series[0], series[0]
	for _, v := range series {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	out := make([]int, cols)
	for c := 0; c < cols; c++ {
		i := 0
		if cols > 1 {
			i = c * (len(series) - 1) / (cols - 1)
		}
		if max == min {
			out[c] = rows / 2
			continue
		}
		out[c] = int((max - series[i]) / (max - min) * float64(rows-1))
	}
	return out
}

// chartBounds returns the series minimum and maximum.
func chartBounds(series []float64) (min, max float64) {
	min, max = series[0], series[0]
	for _, v := range series {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// chartPixelRows maps the series onto pxCols pixel columns of pxRows pixel
// rows, linearly interpolating between samples so the curve is smooth at
// sub-cell resolution. Row 0 is the top (the maximum).
func chartPixelRows(series []float64, pxCols, pxRows int) []int {
	if len(series) < 2 || pxCols < 2 || pxRows < 1 {
		return nil
	}
	min, max := chartBounds(series)
	out := make([]int, pxCols)
	for c := 0; c < pxCols; c++ {
		if max == min {
			out[c] = pxRows / 2
			continue
		}
		pos := float64(c) * float64(len(series)-1) / float64(pxCols-1)
		i := int(pos)
		v := series[i]
		if frac := pos - float64(i); i+1 < len(series) {
			v = series[i]*(1-frac) + series[i+1]*frac
		}
		out[c] = int((max - v) / (max - min) * float64(pxRows-1))
	}
	return out
}

// brailleBits holds the braille dot bit for the pixel at (px, py) inside one
// cell: 2 pixel columns × 4 pixel rows per character cell.
var brailleBits = [4][2]int{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}

// DrawLineChart draws series as a line chart with a labeled y-axis inside
// the rect at (x, y) spanning width×height cells. In Unicode mode the line
// is drawn with braille dots at 2×4 pixels per cell, which reads as a
// smooth curve; ASCII mode falls back to one marker per column. Fewer than
// two samples, or a rect too small to hold the axis, draws nothing.
func DrawLineChart(r domain.Renderer, x, y, width, height int, series []float64, axisStyle, lineStyle tcell.Style) {
	const labelW = 4 // "999 " — right-aligned wpm labels
	cols := width - labelW - 1
	if len(series) < 2 || cols < 2 || height < 2 {
		return
	}

	gl := Glyphs()
	min, max := chartBounds(series)

	// Y-axis: a tick every row, labels on the top, middle, and bottom rows.
	for row := 0; row < height; row++ {
		v := max
		if max != min {
			v = max - (max-min)*float64(row)/float64(height-1)
		}
		if row == 0 || row == height-1 || row == (height-1)/2 {
			label := fmt.Sprintf("%*.0f", labelW-1, v)
			r.DrawText(x+labelW-1-runewidth.StringWidth(label), y+row, axisStyle, label)
		}
		r.DrawText(x+labelW, y+row, axisStyle, gl.ChartTick)
	}

	if IsASCII() {
		drawASCIILine(r, x+labelW+1, y, cols, height, series, lineStyle)
		return
	}

	// Braille line: plot into a 2×4-per-cell pixel grid, joining neighbor
	// columns vertically so steep segments stay connected.
	pxRows := chartPixelRows(series, cols*2, height*4)
	grid := make([][]int, height)
	for i := range grid {
		grid[i] = make([]int, cols)
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
			grid[py/4][c/2] |= brailleBits[py%4][c%2]
		}
		prev = row
	}
	for cy := range grid {
		for cx, mask := range grid[cy] {
			if mask != 0 {
				r.SetContent(x+labelW+1+cx, y+cy, rune(0x2800+mask), lineStyle)
			}
		}
	}
}

// drawASCIILine is the --ascii fallback: a marker per cell column with
// vertical fill between steep neighbors.
func drawASCIILine(r domain.Renderer, x, y, cols, height int, series []float64, lineStyle tcell.Style) {
	gl := Glyphs()
	rows := chartRows(series, cols, height)
	dot := []rune(gl.ChartDot)[0]
	bar := []rune(gl.ChartTick)[0]
	for c, row := range rows {
		r.SetContent(x+c, y+row, dot, lineStyle)
		if c > 0 {
			lo, hi := rows[c-1], row
			if lo > hi {
				lo, hi = hi, lo
			}
			for between := lo + 1; between < hi; between++ {
				r.SetContent(x+c, y+between, bar, lineStyle)
			}
		}
	}
}
