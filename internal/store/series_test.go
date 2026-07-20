package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRoundSeriesRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	s.AppendRound(Round{
		TS: time.Now(), Theme: "cozy", Mode: "ta15", Lang: "en", Source: "words",
		WPM: 76, Acc: 99, DurS: 15,
		RawWPM: 80, CPM: 408, WPMSeries: []float64{60, 74, 76},
	})

	rounds := s.LoadHistory()
	if len(rounds) != 1 {
		t.Fatalf("got %d rounds, want 1", len(rounds))
	}
	r := rounds[0]
	if r.RawWPM != 80 || r.CPM != 408 || len(r.WPMSeries) != 3 || r.WPMSeries[2] != 76 {
		t.Errorf("series fields did not round-trip: %+v", r)
	}
}

func TestOldHistoryLineDecodes(t *testing.T) {
	dir := t.TempDir()
	old := `{"ts":"2026-01-01T00:00:00Z","theme":"simple","mode":"normal","lang":"en","wpm":50,"acc":90,"dur_s":10,"source":"builtin"}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, "history.jsonl"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	rounds := New(dir).LoadHistory()
	if len(rounds) != 1 {
		t.Fatalf("got %d rounds, want 1", len(rounds))
	}
	r := rounds[0]
	if r.WPM != 50 || r.RawWPM != 0 || r.CPM != 0 || r.WPMSeries != nil {
		t.Errorf("old line should decode with zero-valued new fields: %+v", r)
	}
}
