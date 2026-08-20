package main

import (
	"testing"

	"github.com/namest504/termtype/internal/chart"
	"github.com/namest504/termtype/internal/store"
	"github.com/namest504/termtype/internal/ui"
)

func TestChartOptionsFor(t *testing.T) {
	cases := []struct {
		name  string
		code  string
		style chart.Style
		thick int
	}{
		{"braille1 is thin", "braille1", chart.StyleBraille, 1},
		{"braille2 is medium", "braille2", chart.StyleBraille, 2},
		{"braille3 is thick", "braille3", chart.StyleBraille, 3},
		{"box style", "box", chart.StyleBox, 1},
		{"unknown code falls back to braille2", "unknown", chart.StyleBraille, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := chartOptionsFor(tc.code)
			if o.Style != tc.style || o.Thickness != tc.thick || o.Interp != chart.InterpSmooth {
				t.Errorf("chartOptionsFor(%q) = %+v, want style %v thickness %d", tc.code, o, tc.style, tc.thick)
			}
		})
	}
}

func TestSummaryLine(t *testing.T) {
	prev := ui.IsASCII()
	ui.SetASCII(false)
	t.Cleanup(func() { ui.SetASCII(prev) })
	cases := []struct {
		name string
		cfg  store.Config
		want string
	}{
		{"zero config defaults", store.Config{}, "Normal · Sentences · English"},
		{"time attack korean", store.Config{Mode: "ta30", Lang: "ko"}, "Time Attack (30s) · Sentences · 한국어 (Korean)"},
		{"words pins english", store.Config{Source: "words", Lang: "ko"}, "Normal · Words · English"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := summaryLine(tc.cfg); got != tc.want {
				t.Errorf("summaryLine(%+v) = %q, want %q", tc.cfg, got, tc.want)
			}
		})
	}
}

func TestResolveASCII(t *testing.T) {
	clear := func(t *testing.T) {
		t.Helper()
		for _, k := range []string{"TERMTYPE_ASCII", "LC_ALL", "LC_CTYPE", "LANG"} {
			t.Setenv(k, "")
		}
	}
	cases := []struct {
		name    string
		flagSet bool
		env     map[string]string
		want    bool
	}{
		{"explicit flag wins", true, map[string]string{"LANG": "en_US.UTF-8"}, true},
		{"env var on", false, map[string]string{"TERMTYPE_ASCII": "1"}, true},
		{"env var off beats non-utf8 locale", false, map[string]string{"TERMTYPE_ASCII": "off", "LANG": "C"}, false},
		{"invalid env falls through to locale", false, map[string]string{"TERMTYPE_ASCII": "banana", "LANG": "C"}, true},
		{"lc_all beats lang", false, map[string]string{"LC_ALL": "en_US.UTF-8", "LANG": "C"}, false},
		{"lc_ctype beats lang", false, map[string]string{"LC_CTYPE": "C", "LANG": "en_US.UTF-8"}, true},
		{"posix locale is ascii", false, map[string]string{"LANG": "POSIX"}, true},
		{"utf8 without dash", false, map[string]string{"LANG": "ko_KR.utf8"}, false},
		{"no locale assumes utf8", false, nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clear(t)
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			if got := resolveASCII(tc.flagSet); got != tc.want {
				t.Errorf("resolveASCII(%v) with %v = %v, want %v", tc.flagSet, tc.env, got, tc.want)
			}
		})
	}
}
