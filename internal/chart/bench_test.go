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
	opts := []struct {
		name string
		opts Options
	}{
		{"braille-1px", Options{Width: 64, Height: 10, Style: StyleBraille, Interp: InterpSmooth, Thickness: 1}},
		{"braille-3px", Options{Width: 64, Height: 10, Style: StyleBraille, Interp: InterpSmooth, Thickness: 3}},
		{"box", Options{Width: 64, Height: 10, Style: StyleBox, Interp: InterpSmooth}},
		{"ascii", Options{Width: 64, Height: 10, Style: StyleASCII}},
	}
	for _, tc := range opts {
		b.Run(tc.name, func(b *testing.B) {
			for b.Loop() {
				Render(benchSeries, tc.opts)
			}
		})
	}
}

func BenchmarkMonotoneCubic(b *testing.B) {
	for b.Loop() {
		monotoneCubic(benchSeries, 128)
	}
}
