package ui

import "testing"

func TestSparklineScaling(t *testing.T) {
	SetASCII(false)
	// min=1 -> lowest glyph, max=8 -> highest, 4.5 -> level 3 of 0..7.
	if got := Sparkline([]float64{1, 8, 4.5}); got != "▁█▄" {
		t.Errorf("Sparkline() = %q, want ▁█▄", got)
	}
}

func TestSparklineFlat(t *testing.T) {
	SetASCII(false)
	// Equal values sit at the middle level so a flat series stays visible.
	if got := Sparkline([]float64{5, 5}); got != "▅▅" {
		t.Errorf("Sparkline() = %q, want ▅▅", got)
	}
}

func TestSparklineASCII(t *testing.T) {
	SetASCII(true)
	defer SetASCII(false)
	if got := Sparkline([]float64{1, 8}); got != ".%" {
		t.Errorf("Sparkline() = %q, want .%%", got)
	}
}

func TestSparklineEmpty(t *testing.T) {
	if got := Sparkline(nil); got != "" {
		t.Errorf("Sparkline(nil) = %q, want empty string", got)
	}
}
