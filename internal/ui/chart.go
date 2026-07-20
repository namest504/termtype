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

// DrawLineChart draws series as a line chart with a labeled y-axis inside
// the rect at (x, y) spanning width×height cells. Fewer than two samples,
// or a rect too small to hold the axis, draws nothing.
func DrawLineChart(r domain.Renderer, x, y, width, height int, series []float64, axisStyle, lineStyle tcell.Style) {
	const labelW = 4 // "999 " — right-aligned wpm labels
	cols := width - labelW - 1
	if len(series) < 2 || cols < 2 || height < 2 {
		return
	}

	gl := Glyphs()
	min, max := chartBounds(series)
	rows := chartRows(series, cols, height)

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

	// The line: a dot per column, with vertical fill between steep neighbors
	// so the curve reads as connected.
	dot := []rune(gl.ChartDot)[0]
	bar := []rune(gl.ChartTick)[0]
	if !IsASCII() {
		bar = []rune("│")[0]
	}
	for c, row := range rows {
		r.SetContent(x+labelW+1+c, y+row, dot, lineStyle)
		if c > 0 {
			lo, hi := rows[c-1], row
			if lo > hi {
				lo, hi = hi, lo
			}
			for between := lo + 1; between < hi; between++ {
				r.SetContent(x+labelW+1+c, y+between, bar, lineStyle)
			}
		}
	}
}
