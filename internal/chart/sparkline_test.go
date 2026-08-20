package chart

import "testing"

func TestSparklineScaling(t *testing.T) {
	got := Sparkline([]float64{0, 100}, false)
	if got != "▁█" {
		t.Fatalf("got %q, want ▁█", got)
	}
}

func TestSparklineFlat(t *testing.T) {
	got := Sparkline([]float64{50, 50, 50}, false)
	if got != "▅▅▅" {
		t.Fatalf("flat series should sit mid-level, got %q", got)
	}
}

func TestSparklineASCII(t *testing.T) {
	got := Sparkline([]float64{0, 100}, true)
	if got != ".%" {
		t.Fatalf("got %q, want .%%", got)
	}
}

func TestSparklineEmpty(t *testing.T) {
	if got := Sparkline(nil, false); got != "" {
		t.Fatalf("empty input should yield empty string, got %q", got)
	}
}
