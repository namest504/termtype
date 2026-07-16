package store

import (
	"strings"
	"testing"
	"time"
)

func TestBestRespectsBucketAndEligibility(t *testing.T) {
	rounds := []Round{
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 90, DurS: 3}, // too short for PB
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 70, DurS: 10},
		{Mode: "normal", Lang: "ko", Source: "builtin", WPM: 80, DurS: 10}, // other bucket
		{Mode: "ta30", Lang: "en", Source: "builtin", WPM: 85, DurS: 30},   // other bucket
		{Mode: "normal", Lang: "en", Source: "custom", WPM: 95, DurS: 10},  // other bucket
	}
	best, ok := Best(rounds, Key{Mode: "normal", Lang: "en", Source: "builtin"})
	if !ok || best != 70 {
		t.Errorf("Best() = %v, %v; want 70, true", best, ok)
	}
}

func TestBestEmpty(t *testing.T) {
	if _, ok := Best(nil, Key{Mode: "normal", Lang: "en", Source: "builtin"}); ok {
		t.Error("Best() on empty history reported ok")
	}
}

func TestRecentWPMsLastN(t *testing.T) {
	k := Key{Mode: "normal", Lang: "en", Source: "builtin"}
	var rounds []Round
	for i := 1; i <= 12; i++ {
		rounds = append(rounds, Round{Mode: "normal", Lang: "en", Source: "builtin",
			WPM: float64(i), DurS: 10})
	}
	rounds = append(rounds, Round{Mode: "ta30", Lang: "en", Source: "builtin", WPM: 99, DurS: 30})
	got := RecentWPMs(rounds, k, 10)
	if len(got) != 10 || got[0] != 3 || got[9] != 12 {
		t.Errorf("RecentWPMs() = %v, want WPMs 3 through 12", got)
	}
}

func TestModeString(t *testing.T) {
	cases := []struct {
		limit time.Duration
		want  string
	}{{0, "normal"}, {30 * time.Second, "ta30"}, {60 * time.Second, "ta60"}}
	for _, c := range cases {
		if got := ModeString(c.limit); got != c.want {
			t.Errorf("ModeString(%v) = %q, want %q", c.limit, got, c.want)
		}
	}
}

func TestFormatStats(t *testing.T) {
	rounds := []Round{
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 60, Acc: 95, DurS: 10},
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 70, Acc: 97, DurS: 10},
	}
	out := FormatStats(rounds)
	for _, want := range []string{"normal", "en", "runs 2", "best  70.0", "avg  65.0", "acc 96.0%"} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatStats() output %q missing %q", out, want)
		}
	}
}

func TestFormatStatsNoEligibleBest(t *testing.T) {
	rounds := []Round{
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 90, Acc: 99, DurS: 3},
	}
	out := FormatStats(rounds)
	if strings.Contains(out, "best   0.0") {
		t.Errorf("FormatStats() = %q, must not print a 0.0 best for a bucket with no eligible round", out)
	}
	if !strings.Contains(out, "best     -") {
		t.Errorf("FormatStats() = %q, want a dash placeholder for the missing best", out)
	}
}

func TestFormatStatsEmpty(t *testing.T) {
	if out := FormatStats(nil); !strings.Contains(out, "No history") {
		t.Errorf("FormatStats(nil) = %q, want a no-history notice", out)
	}
}
