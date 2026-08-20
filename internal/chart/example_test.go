package chart_test

import (
	"fmt"
	"strings"

	"github.com/namest504/termtype/internal/chart"
)

func ExampleSparkline() {
	fmt.Println(chart.Sparkline([]float64{1, 2, 3, 4, 5, 6, 7, 8}, false))
	// Output: ▁▂▃▄▅▆▇█
}

func ExampleRender() {
	grid := chart.Render([]float64{5, 5}, chart.Options{
		Width: 12, Height: 4, Style: chart.StyleBox,
	})
	var b strings.Builder
	for _, row := range grid {
		var line strings.Builder
		for _, cell := range row {
			line.WriteRune(cell.Rune)
		}
		b.WriteString(strings.TrimRight(line.String(), " "))
		b.WriteRune('\n')
	}
	fmt.Print(b.String())
	// Output:
	//   5 ┤
	//   5 ┤
	//     ┤───────
	//   5 ┤
}
