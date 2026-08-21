package chart

import "testing"

// 단조 구간에서 보간값도 단조 — 오버슈트 없음이 스펙 요구사항이다.
func TestMonotoneCubicNoOvershoot(t *testing.T) {
	series := []float64{0, 10, 12, 60, 62, 100}
	got := monotoneCubic(series, 101)
	lo, hi := bounds(series)
	prev := got[0]
	for i, v := range got {
		if v < lo-1e-9 || v > hi+1e-9 {
			t.Fatalf("overshoot at %d: %v outside [%v,%v]", i, v, lo, hi)
		}
		if v < prev-1e-9 {
			t.Fatalf("monotone increasing series produced a dip at %d: %v < %v", i, v, prev)
		}
		prev = v
	}
}

func TestMonotoneCubicHitsSamples(t *testing.T) {
	series := []float64{20, 50, 30, 80}
	got := monotoneCubic(series, 7) // n = 2*(len-1)+1 → 짝수 인덱스가 원본 샘플
	for i, want := range series {
		if diff := got[i*2] - want; diff > 1e-9 || diff < -1e-9 {
			t.Fatalf("sample %d: got %v, want %v", i, got[i*2], want)
		}
	}
}

func TestMonotoneCubicFlat(t *testing.T) {
	for _, v := range monotoneCubic([]float64{5, 5, 5}, 10) {
		if v != 5 {
			t.Fatalf("flat series interpolated to %v", v)
		}
	}
}

func TestSampleSmoothFallsBackBelowThree(t *testing.T) {
	got := sample([]float64{0, 10}, 3, InterpSmooth)
	want := []float64{0, 5, 10}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("2-sample smooth should be linear, got %v", got)
		}
	}
}
