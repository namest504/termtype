package main

import (
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/themes"
	"github.com/namest504/termtype/internal/ui"
)

// sortedThemeNames returns the registry names in menu order: cozy leads
// (the default), log keeps second place, the rest follow alphabetically.
func sortedThemeNames() []string {
	var names []string
	for name := range themes.Themes {
		names = append(names, name)
	}
	rank := func(name string) int {
		switch name {
		case "cozy":
			return 0
		case "log":
			return 1
		}
		return 2
	}
	sort.Slice(names, func(i, j int) bool {
		if ri, rj := rank(names[i]), rank(names[j]); ri != rj {
			return ri < rj
		}
		return names[i] < names[j]
	})
	return names
}

// menuAction is what a key press asks the menu loop to do.
type menuAction int

const (
	actNone menuAction = iota
	actStart
	actSettings
	actHistory
	actQuit
)

// menuModel is the main-menu state: a theme carousel that can expand into
// a full list. Drawing is separate so transitions are unit-testable.
type menuModel struct {
	themes   []string
	idx      int  // carousel position (the picked theme)
	expanded bool // theme list unfolded below the carousel
	sel      int  // list selection while expanded
}

func newMenuModel(cfgTheme string) menuModel {
	names := sortedThemeNames()
	return menuModel{
		themes: names,
		idx:    indexOf(len(names), func(i int) bool { return names[i] == cfgTheme }),
	}
}

func (m *menuModel) handleKey(ev *tcell.EventKey) menuAction {
	if m.expanded {
		switch ev.Key() {
		case tcell.KeyUp:
			if m.sel > 0 {
				m.sel--
			}
		case tcell.KeyDown:
			if m.sel < len(m.themes)-1 {
				m.sel++
			}
		case tcell.KeyEnter:
			m.idx = m.sel
			m.expanded = false
		case tcell.KeyEscape:
			m.expanded = false
		case tcell.KeyCtrlC:
			return actQuit
		}
		return actNone
	}
	switch ev.Key() {
	case tcell.KeyLeft:
		m.idx = cycleIdx(m.idx, -1, len(m.themes))
	case tcell.KeyRight:
		m.idx = cycleIdx(m.idx, 1, len(m.themes))
	case tcell.KeyDown:
		m.expanded, m.sel = true, m.idx
	case tcell.KeyEnter:
		return actStart
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return actQuit
	case tcell.KeyRune:
		switch ev.Rune() {
		case 's', 'S':
			return actSettings
		case 'h', 'H':
			return actHistory
		}
	}
	return actNone
}

// drawMenu renders the carousel main screen; summary is the read-only
// "Mode · Text · Language" line built by the caller from config.
func drawMenu(s tcell.Screen, m menuModel, summary string) {
	s.Clear()
	w, _ := s.Size()
	gl := ui.Glyphs()
	centered := func(y int, style tcell.Style, text string) {
		x := (w - runewidth.StringWidth(text)) / 2
		if x < 0 {
			x = 0
		}
		drawText(s, x, y, style, ui.Truncate(text, w))
	}

	centered(1, tcell.StyleDefault.Bold(true), "termtype")

	l, r := "‹", "›"
	if ui.IsASCII() {
		l, r = "<", ">"
	}
	centered(3, tcell.StyleDefault.Reverse(true), "  "+l+"  "+m.themes[m.idx]+"  "+r+"  ")
	centered(5, tcell.StyleDefault.Foreground(tcell.ColorGray), summary)

	helpY := 7
	if m.expanded {
		for i, name := range m.themes {
			style := tcell.StyleDefault
			if i == m.sel {
				style = style.Reverse(true)
			}
			centered(7+i, style, " "+name+" ")
		}
		helpY = 7 + len(m.themes) + 1
		centered(helpY, tcell.StyleDefault.Foreground(tcell.ColorGray),
			gl.ArrowUD+" pick "+gl.Sep+" "+gl.Enter+" select "+gl.Sep+" Esc close")
		s.Show()
		return
	}

	full := strings.Join([]string{
		gl.Enter + " start", "s settings", "h history", "Esc quit",
	}, " "+gl.Sep+" ")
	compact := gl.Enter + " start " + gl.Sep + " s settings"
	help := full
	if runewidth.StringWidth(help) > w-2 {
		help = compact
	}
	if runewidth.StringWidth(help) > w-2 {
		help = gl.Enter + " start"
	}
	centered(helpY, tcell.StyleDefault.Foreground(tcell.ColorGray), help)
	s.Show()
}
