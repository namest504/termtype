package chart

import "testing"

var benchSeries = func() []float64 {
	s := make([]float64, 60)
	for i := range s {
		s[i] = float64(30 + (i*7)%40)
	}
	return s
}()

func BenchmarkRender(b *testing.B) {
	opts := map[string]Options{
		"braille-1px": {Width: 64, Height: 10, Style: StyleBraille, Interp: InterpSmooth, Thickness: 1},
		"braille-3px": {Width: 64, Height: 10, Style: StyleBraille, Interp: InterpSmooth, Thickness: 3},
		"box":         {Width: 64, Height: 10, Style: StyleBox, Interp: InterpSmooth},
		"ascii":       {Width: 64, Height: 10, Style: StyleASCII},
	}
	for name, o := range opts {
		b.Run(name, func(b *testing.B) {
			for b.Loop() {
				Render(benchSeries, o)
			}
		})
	}
}

func BenchmarkMonotoneCubic(b *testing.B) {
	for b.Loop() {
		monotoneCubic(benchSeries, 128)
	}
}
