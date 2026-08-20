package chart

import "testing"

func TestRenderNilOnDegenerate(t *testing.T) {
	o := Options{Width: 20, Height: 5}
	if Render([]float64{42}, o) != nil {
		t.Fatal("single sample should render nil")
	}
	if Render([]float64{1, 2}, Options{Width: 6, Height: 5}) != nil {
		t.Fatal("too-narrow rect should render nil")
	}
	if Render([]float64{1, 2}, Options{Width: 20, Height: 1}) != nil {
		t.Fatal("too-short rect should render nil")
	}
}

func TestRenderDimensions(t *testing.T) {
	o := Options{Width: 20, Height: 5}
	g := Render([]float64{10, 20, 30}, o)
	if len(g) != 5 || len(g[0]) != 20 {
		t.Fatalf("grid is %dx%d, want 5x20", len(g), len(g[0]))
	}
}

func TestRenderAxisAndLabels(t *testing.T) {
	o := Options{Width: 20, Height: 5}
	g := Render([]float64{0, 100}, o)
	const labelW = 4
	for row := 0; row < 5; row++ {
		if g[row][labelW].Kind != KindAxis || g[row][labelW].Rune != '┤' {
			t.Fatalf("row %d col %d should be axis tick, got %q/%v",
				row, labelW, g[row][labelW].Rune, g[row][labelW].Kind)
		}
	}
	// top row label reads "100", bottom "  0" (right-aligned width 3)
	top := string([]rune{g[0][0].Rune, g[0][1].Rune, g[0][2].Rune})
	if top != "100" {
		t.Fatalf("top label %q, want 100", top)
	}
	if g[0][0].Kind != KindLabel {
		t.Fatalf("label cell kind = %v, want KindLabel", g[0][0].Kind)
	}
}

func TestRenderLineCellsTagged(t *testing.T) {
	o := Options{Width: 20, Height: 5}
	g := Render([]float64{0, 50, 100}, o)
	found := false
	for _, row := range g {
		for cx, c := range row {
			if c.Kind == KindLine {
				if cx <= 4 {
					t.Fatalf("line cell inside axis area at col %d", cx)
				}
				found = true
			}
		}
	}
	if !found {
		t.Fatal("no line cells rendered")
	}
}

func TestRenderFlatSeriesMidRow(t *testing.T) {
	o := Options{Width: 20, Height: 5}
	g := Render([]float64{50, 50, 50, 50}, o)
	for row := 0; row < 5; row++ {
		hasLine := false
		for _, c := range g[row] {
			if c.Kind == KindLine {
				hasLine = true
			}
		}
		// flat series → all line pixels land in the middle pixel row band
		if hasLine && row != 2 {
			t.Fatalf("flat series drew on cell row %d, want only row 2", row)
		}
	}
}

func TestRenderASCIIStyle(t *testing.T) {
	o := Options{Width: 20, Height: 5, Style: StyleASCII}
	g := Render([]float64{0, 100, 0}, o)
	stars := 0
	for _, row := range g {
		for _, c := range row {
			if c.Kind == KindLine && c.Rune == '*' {
				stars++
			}
			if c.Rune >= 0x2800 && c.Rune <= 0x28FF {
				t.Fatal("ASCII style must not emit braille runes")
			}
		}
	}
	if stars == 0 {
		t.Fatal("ASCII style should draw * markers")
	}
	if g[0][4].Rune != '|' {
		t.Fatalf("ASCII tick should be '|', got %q", g[0][4].Rune)
	}
}

func TestSampleLinearInterpolates(t *testing.T) {
	got := sample([]float64{0, 10}, 3, InterpLinear)
	want := []float64{0, 5, 10}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sample[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestPixelRowsFlat(t *testing.T) {
	rows := pixelRows([]float64{5, 5, 5, 5}, 8, 5, 5)
	for _, r := range rows {
		if r != 4 {
			t.Fatalf("flat series pixel row = %d, want 4", r)
		}
	}
}

// TestRenderBrailleScalesAgainstTrueSeriesBounds is a regression test for a
// bug where the braille line was scaled against the bounds of the
// downsampled array (sample(series, cols*2, ...)) instead of the true
// series bounds used by the axis labels. When cols*2 < len(series), a
// narrow peak can be smoothed away by interpolation before scaling, so the
// sampled max is well below the true max — the line then reports as
// touching the axis top even though it is nowhere near the labeled max.
//
// series has a single spike (100) at index 3 out of 17 points, everything
// else 0. With Width=9 the plot area is 4 cols, so cols*2 = 8 sample
// points — well under 17, so the spike lands between two sample points and
// is interpolated down to ~28.6 instead of surviving as 100.
func TestRenderBrailleScalesAgainstTrueSeriesBounds(t *testing.T) {
	series := make([]float64, 17)
	series[3] = 100

	o := Options{Width: 9, Height: 10}
	cols := o.Width - labelW - 1
	if cols*2 >= len(series) {
		t.Fatalf("test setup invalid: cols*2=%d must be < len(series)=%d", cols*2, len(series))
	}

	// Expected top pixel row, computed against the TRUE series bounds
	// (what the axis labels use), not the sampled array's own bounds.
	lo, hi := bounds(series)
	values := sample(series, cols*2, InterpLinear)
	peak := values[0]
	for _, v := range values {
		if v > peak {
			peak = v
		}
	}
	pxRows := o.Height * 4
	expectedPixelRow := int((hi - peak) / (hi - lo) * float64(pxRows-1))
	expectedCellRow := expectedPixelRow / 4

	g := Render(series, o)
	topRow := -1
	for row := range g {
		for _, c := range g[row] {
			if c.Kind == KindLine {
				topRow = row
				break
			}
		}
		if topRow != -1 {
			break
		}
	}
	if topRow == -1 {
		t.Fatal("no line cells rendered")
	}
	if topRow != expectedCellRow {
		t.Fatalf("topmost line cell row = %d, want %d (scaled against true series bounds; "+
			"got row 0 would indicate the bug — scaling against the sampled array's own bounds)",
			topRow, expectedCellRow)
	}
}

func brailleDotCount(g [][]Cell) int {
	n := 0
	for _, row := range g {
		for _, c := range row {
			if c.Rune >= 0x2800 && c.Rune <= 0x28FF {
				for mask := int(c.Rune - 0x2800); mask != 0; mask &= mask - 1 {
					n++
				}
			}
		}
	}
	return n
}

func TestThicknessAddsPixels(t *testing.T) {
	series := []float64{0, 30, 60, 100}
	thin := Render(series, Options{Width: 30, Height: 6, Thickness: 1})
	thick := Render(series, Options{Width: 30, Height: 6, Thickness: 2})
	if brailleDotCount(thick) <= brailleDotCount(thin) {
		t.Fatalf("thickness 2 (%d dots) should light more pixels than 1 (%d)",
			brailleDotCount(thick), brailleDotCount(thin))
	}
}

func TestThicknessClampsAtBottom(t *testing.T) {
	// a series hugging the minimum: thick pixels must not run past the grid
	g := Render([]float64{0, 0, 100}, Options{Width: 20, Height: 3, Thickness: 3})
	if g == nil {
		t.Fatal("render returned nil")
	}
	// reaching here without an index-out-of-range panic is the real assertion
}

func TestThicknessZeroMeansOne(t *testing.T) {
	series := []float64{0, 50, 100}
	zero := Render(series, Options{Width: 30, Height: 6, Thickness: 0})
	one := Render(series, Options{Width: 30, Height: 6, Thickness: 1})
	if brailleDotCount(zero) != brailleDotCount(one) {
		t.Fatal("thickness 0 should behave as 1")
	}
}
