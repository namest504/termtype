package themes

import "testing"

func TestHexWindow(t *testing.T) {
	cases := []struct {
		name               string
		targetLen, itLen   int
		avail              int
		wantStart, wantVis int
	}{
		{"short target fits", 40, 0, 10, 0, 3},        // 3 rows, all visible
		{"start of long target", 400, 0, 5, 0, 5},     // 25 rows, window at top
		{"middle of long target", 400, 200, 5, 10, 5}, // cursor row 12 centered
		{"end of long target", 400, 399, 5, 20, 5},    // pinned to the last rows
		{"input past target end", 400, 500, 5, 20, 5}, // clamped
		{"tiny terminal", 400, 0, 0, 0, 1},            // at least one row
		{"empty target", 0, 0, 5, 0, 1},               // never zero rows
	}
	for _, c := range cases {
		start, vis := hexWindow(c.targetLen, c.itLen, c.avail)
		if start != c.wantStart || vis != c.wantVis {
			t.Errorf("%s: hexWindow(%d, %d, %d) = (%d, %d), want (%d, %d)",
				c.name, c.targetLen, c.itLen, c.avail, start, vis, c.wantStart, c.wantVis)
		}
	}
}
