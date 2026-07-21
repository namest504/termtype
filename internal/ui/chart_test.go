package ui

import "testing"

func TestChartRowsMapping(t *testing.T) {
	rows := chartRows([]float64{60, 94}, 10, 6)
	if len(rows) != 10 {
		t.Fatalf("got %d columns, want 10", len(rows))
	}
	if rows[0] != 5 {
		t.Errorf("minimum should sit on the bottom row, got %d", rows[0])
	}
	if rows[9] != 0 {
		t.Errorf("maximum should sit on the top row, got %d", rows[9])
	}
	for _, r := range rows {
		if r < 0 || r > 5 {
			t.Errorf("row %d out of range", r)
		}
	}
}

func TestChartRowsFlatSeries(t *testing.T) {
	rows := chartRows([]float64{76, 76, 76}, 8, 6)
	for _, r := range rows {
		if r != 3 {
			t.Errorf("flat series should sit on the middle row, got %d", r)
		}
	}
}

func TestChartRowsDegenerate(t *testing.T) {
	if chartRows([]float64{76}, 10, 6) != nil {
		t.Error("a single sample should draw nothing")
	}
	if chartRows(nil, 10, 6) != nil {
		t.Error("an empty series should draw nothing")
	}
	if chartRows([]float64{1, 2}, 0, 6) != nil {
		t.Error("zero columns should draw nothing")
	}
}

func TestChartPixelRowsInterpolates(t *testing.T) {
	// Two samples over four pixel columns: the middle columns interpolate.
	rows := chartPixelRows([]float64{0, 30}, 4, 31)
	want := []int{30, 20, 10, 0}
	for i, w := range want {
		if rows[i] != w {
			t.Errorf("pixel col %d = %d, want %d (rows %v)", i, rows[i], w, rows)
		}
	}
}

func TestChartPixelRowsFlatAndDegenerate(t *testing.T) {
	for _, r := range chartPixelRows([]float64{76, 76}, 8, 8) {
		if r != 4 {
			t.Errorf("flat series pixel row = %d, want middle 4", r)
		}
	}
	if chartPixelRows([]float64{1}, 8, 8) != nil {
		t.Error("single sample should draw nothing")
	}
}

func TestBrailleBitsDistinct(t *testing.T) {
	seen := map[int]bool{}
	sum := 0
	for _, row := range brailleBits {
		for _, b := range row {
			if seen[b] {
				t.Errorf("bit %#x repeated", b)
			}
			seen[b] = true
			sum |= b
		}
	}
	if sum != 0xFF {
		t.Errorf("bits cover %#x, want 0xFF", sum)
	}
}

func TestChartRowsCompress(t *testing.T) {
	series := make([]float64, 120) // longer than the chart is wide
	for i := range series {
		series[i] = float64(i)
	}
	rows := chartRows(series, 30, 8)
	if len(rows) != 30 {
		t.Fatalf("got %d columns, want 30", len(rows))
	}
	if rows[0] != 7 || rows[29] != 0 {
		t.Errorf("compressed series should still span bottom to top, got %d..%d", rows[0], rows[29])
	}
}
