package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	want := Config{Theme: "matrix", Mode: "ta30", Lang: "ko", Ghost: true}
	s.SaveConfig(want)
	if got := s.LoadConfig(); got != want {
		t.Errorf("LoadConfig() = %+v, want %+v", got, want)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	s := New(t.TempDir())
	if got := s.LoadConfig(); got != (Config{}) {
		t.Errorf("LoadConfig() with no file = %+v, want zero value", got)
	}
}

func TestHistoryAppendLoad(t *testing.T) {
	s := New(t.TempDir())
	r1 := Round{TS: time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC), Theme: "simple",
		Mode: "normal", Lang: "en", WPM: 60.5, Acc: 97.2, DurS: 12.3, Source: "builtin"}
	r2 := r1
	r2.WPM = 65.0
	s.AppendRound(r1)
	s.AppendRound(r2)
	got := s.LoadHistory()
	if len(got) != 2 || got[0].WPM != 60.5 || got[1].WPM != 65.0 {
		t.Errorf("LoadHistory() = %+v, want the two appended rounds in order", got)
	}
	if !got[0].TS.Equal(r1.TS) || got[0].Theme != "simple" || got[0].Source != "builtin" {
		t.Errorf("LoadHistory()[0] = %+v, want %+v", got[0], r1)
	}
}

func TestLoadHistorySkipsCorruptLines(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.AppendRound(Round{TS: time.Now(), WPM: 50, DurS: 10})
	f, err := os.OpenFile(filepath.Join(dir, "history.jsonl"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("not json\n{}\n")
	f.Close()
	s.AppendRound(Round{TS: time.Now(), WPM: 55, DurS: 10})
	if got := s.LoadHistory(); len(got) != 2 {
		t.Errorf("LoadHistory() returned %d rounds, want 2 (corrupt lines skipped)", len(got))
	}
}

func TestNoopStore(t *testing.T) {
	var nilStore *Store
	for name, s := range map[string]*Store{"nil": nilStore, "empty": {}} {
		s.SaveConfig(Config{Theme: "x"})
		s.AppendRound(Round{TS: time.Now()})
		if got := s.LoadHistory(); got != nil {
			t.Errorf("%s store LoadHistory() = %v, want nil", name, got)
		}
		if got := s.LoadConfig(); got != (Config{}) {
			t.Errorf("%s store LoadConfig() = %+v, want zero value", name, got)
		}
	}
}

func TestChartStyleDefault(t *testing.T) {
	if got := (Config{}).ChartStyle(); got != "braille2" {
		t.Fatalf("empty style should default to braille2, got %q", got)
	}
	if got := (Config{Style: "box"}).ChartStyle(); got != "box" {
		t.Fatalf("explicit style should pass through, got %q", got)
	}
}
