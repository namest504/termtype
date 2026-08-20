package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/chart"
	"github.com/namest504/termtype/internal/domain"
)

// chartOpts is the app-wide chart rendering default, set once from config
// at startup (mirroring SetASCII). Width/Height are per-call.
var chartOpts = chart.Options{Style: chart.StyleBraille, Interp: chart.InterpLinear, Thickness: 1}

// SetChartOptions selects the chart style/interpolation/thickness used by
// DrawLineChart. Width and Height on o are ignored.
func SetChartOptions(o chart.Options) { chartOpts = o }

// DrawLineChart draws series as a line chart with a labeled y-axis inside
// the rect at (x, y) spanning width×height cells, using the app-wide chart
// options. --ascii mode overrides the style with the ASCII fallback.
func DrawLineChart(r domain.Renderer, x, y, width, height int, series []float64, axisStyle, lineStyle tcell.Style) {
	o := chartOpts
	o.Width, o.Height = width, height
	if IsASCII() {
		o.Style = chart.StyleASCII
	}
	for cy, row := range chart.Render(series, o) {
		for cx, c := range row {
			if c.Kind == chart.KindNone {
				continue
			}
			st := axisStyle
			if c.Kind == chart.KindLine {
				st = lineStyle
			}
			r.SetContent(x+cx, y+cy, c.Rune, st)
		}
	}
}
