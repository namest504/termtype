package ui

import "github.com/namest504/termtype/internal/chart"

// Sparkline renders values as one block glyph per value. It delegates to
// the chart package using the active glyph mode.
func Sparkline(values []float64) string {
	return chart.Sparkline(values, IsASCII())
}
