package chart

import "testing"

func braille2(o Options, series []float64) [][]Cell {
	return Render(series, o)
}

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
	rows := pixelRows([]float64{5, 5, 5, 5}, 8)
	for _, r := range rows {
		if r != 4 {
			t.Fatalf("flat series pixel row = %d, want 4", r)
		}
	}
}
