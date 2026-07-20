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
