package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/chart"
	"github.com/namest504/termtype/internal/store"
	"github.com/namest504/termtype/internal/ui"
)

// chartStyles is the single source of truth for the result-graph styles:
// codes are what config.json stores (see Config.ChartStyle), and opts is
// the chart.Options each code renders with. Both the settings screen and
// chartOptionsFor derive from this table so they can never disagree on
// what an unknown/legacy code falls back to.
var chartStyles = []struct {
	code, label string
	opts        chart.Options
}{
	{"braille1", "braille · 1px", chart.Options{Style: chart.StyleBraille, Interp: chart.InterpSmooth, Thickness: 1}},
	{"braille2", "braille · 2px", chart.Options{Style: chart.StyleBraille, Interp: chart.InterpSmooth, Thickness: 2}},
	{"braille3", "braille · 3px", chart.Options{Style: chart.StyleBraille, Interp: chart.InterpSmooth, Thickness: 3}},
	{"box", "box", chart.Options{Style: chart.StyleBox, Interp: chart.InterpSmooth, Thickness: 1}},
}

// defaultChartStyleIdx is the table index for store.Config{}.ChartStyle()
// (currently "braille2"), used as the fallback when a code isn't found.
var defaultChartStyleIdx = indexOf(len(chartStyles), func(i int) bool {
	return chartStyles[i].code == store.Config{}.ChartStyle()
})

const settingsRows = 5 // Mode, Text, Language, Graph, Style

// settingsModel is the settings screen state, kept free of drawing so key
// transitions are unit-testable.
type settingsModel struct {
	row      int
	modeIdx  int
	srcIdx   int
	langIdx  int
	styleIdx int
	graphOn  bool
}

// styleIdxFor finds a style code's index in chartStyles, falling back to
// the braille2 entry when the code is unknown (e.g. a stale/hand-edited
// config) so the settings screen shows the same style chartOptionsFor
// renders.
func styleIdxFor(code string) int {
	idx := indexOf(len(chartStyles), func(i int) bool { return chartStyles[i].code == code })
	if chartStyles[idx].code != code {
		return defaultChartStyleIdx
	}
	return idx
}

func newSettingsModel(cfg store.Config) settingsModel {
	return settingsModel{
		modeIdx:  indexOf(len(gameModes), func(i int) bool { return store.ModeString(gameModes[i].limit) == cfg.Mode }),
		srcIdx:   indexOf(len(textSources), func(i int) bool { return textSources[i].code == cfg.Source }),
		langIdx:  indexOf(len(languages), func(i int) bool { return languages[i].code == cfg.Lang }),
		styleIdx: styleIdxFor(cfg.ChartStyle()),
		graphOn:  cfg.GraphAuto(),
	}
}

// handleKey advances the model. changed means a value moved (caller saves);
// done means Esc closed the screen.
func (m *settingsModel) handleKey(ev *tcell.EventKey) (changed, done bool) {
	switch ev.Key() {
	case tcell.KeyEscape:
		return false, true
	case tcell.KeyUp:
		if m.row > 0 {
			m.row--
		}
	case tcell.KeyDown:
		if m.row < settingsRows-1 {
			m.row++
		}
	case tcell.KeyLeft:
		return m.cycle(-1), false
	case tcell.KeyRight:
		return m.cycle(1), false
	}
	return false, false
}

func cycleIdx(i, d, n int) int { return (i + d + n) % n }

func (m *settingsModel) cycle(d int) bool {
	switch m.row {
	case 0:
		m.modeIdx = cycleIdx(m.modeIdx, d, len(gameModes))
	case 1:
		m.srcIdx = cycleIdx(m.srcIdx, d, len(textSources))
	case 2:
		// The words pool is English-only; the row is pinned while it is active.
		if textSources[m.srcIdx].code == "words" {
			return false
		}
		m.langIdx = cycleIdx(m.langIdx, d, len(languages))
	case 3:
		m.graphOn = !m.graphOn
	case 4:
		m.styleIdx = cycleIdx(m.styleIdx, d, len(chartStyles))
	}
	return true
}

// apply writes the model's values onto cfg, leaving unrelated fields alone.
func (m settingsModel) apply(cfg store.Config) store.Config {
	cfg.Mode = store.ModeString(gameModes[m.modeIdx].limit)
	cfg.Source = textSources[m.srcIdx].code
	cfg.Lang = languages[m.langIdx].code
	cfg.Graph = "on"
	if !m.graphOn {
		cfg.Graph = "off"
	}
	cfg.Style = chartStyles[m.styleIdx].code
	return cfg
}

// runSettings shows the settings screen. Every value change is saved to
// config immediately; Esc returns to the menu.
func runSettings(s tcell.Screen, events <-chan tcell.Event, cfg *store.Config, st *store.Store) {
	m := newSettingsModel(*cfg)
	for {
		drawSettings(s, m)
		switch ev := (<-events).(type) {
		case nil:
			return
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				// quit is the menu's job; treat as Esc here
				return
			}
			changed, done := m.handleKey(ev)
			if changed {
				*cfg = m.apply(*cfg)
				st.SaveConfig(*cfg)
				ui.SetChartOptions(chartOptionsFor(cfg.ChartStyle()))
			}
			if done {
				return
			}
		}
	}
}

func drawSettings(s tcell.Screen, m settingsModel) {
	s.Clear()
	w, _ := s.Size()
	gl := ui.Glyphs()
	drawText(s, 2, 1, tcell.StyleDefault.Bold(true), ui.Truncate("Settings", w-2))

	langName := languages[m.langIdx].name
	langPinned := textSources[m.srcIdx].code == "words"
	if langPinned {
		langName = "English"
	}
	graph := "On"
	if !m.graphOn {
		graph = "Off"
	}
	style := chartStyles[m.styleIdx].label
	if ui.IsASCII() {
		style += " (ascii)"
	}
	rows := []struct {
		name, value string
		dim         bool
	}{
		{"Mode", gameModes[m.modeIdx].name, false},
		{"Text", textSources[m.srcIdx].name, false},
		{"Language", langName, langPinned},
		{"Graph", graph, false},
		{"Style", style, false},
	}
	for i, row := range rows {
		st := tcell.StyleDefault
		if row.dim {
			st = st.Foreground(tcell.ColorGray)
		}
		if i == m.row {
			st = st.Reverse(true)
		}
		line := fmt.Sprintf("%-10s %s %s %s", row.name, "‹", row.value, "›")
		if ui.IsASCII() {
			line = fmt.Sprintf("%-10s < %s >", row.name, row.value)
		}
		drawText(s, 3, 3+i, st, ui.Truncate(line, w-3))
	}
	help := gl.ArrowUD + " select " + gl.Sep + " " + gl.ArrowLR + " change " + gl.Sep + " Esc back"
	drawText(s, 2, 3+settingsRows+1, tcell.StyleDefault.Foreground(tcell.ColorGray), ui.Truncate(help, w-2))
	s.Show()
}
