package chart

import "math"

// monotoneCubic interpolates series onto n points with the Fritsch–Carlson
// monotone cubic scheme: the curve is smooth but never overshoots its
// samples, so a WPM graph never shows a peak the player didn't hit.
func monotoneCubic(series []float64, n int) []float64 {
	m := len(series)
	// secant slopes and tangents
	d := make([]float64, m-1)
	for i := range d {
		d[i] = series[i+1] - series[i]
	}
	t := make([]float64, m)
	t[0], t[m-1] = d[0], d[m-2]
	for i := 1; i < m-1; i++ {
		if d[i-1]*d[i] <= 0 {
			t[i] = 0
		} else {
			t[i] = (d[i-1] + d[i]) / 2
		}
	}
	// Fritsch–Carlson limiter keeps each segment monotone.
	for i := 0; i < m-1; i++ {
		if d[i] == 0 {
			t[i], t[i+1] = 0, 0
			continue
		}
		a, b := t[i]/d[i], t[i+1]/d[i]
		if s := a*a + b*b; s > 9 {
			tau := 3 / math.Sqrt(s)
			t[i] = tau * a * d[i]
			t[i+1] = tau * b * d[i]
		}
	}
	out := make([]float64, n)
	for c := 0; c < n; c++ {
		pos := float64(c) * float64(m-1) / float64(n-1)
		i := int(pos)
		if i >= m-1 {
			i = m - 2
		}
		x := pos - float64(i)
		h00 := (1 + 2*x) * (1 - x) * (1 - x)
		h10 := x * (1 - x) * (1 - x)
		h01 := x * x * (3 - 2*x)
		h11 := x * x * (x - 1)
		out[c] = h00*series[i] + h10*t[i] + h01*series[i+1] + h11*t[i+1]
	}
	return out
}
