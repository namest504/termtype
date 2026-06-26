package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"termtype/internal/app"
	"termtype/internal/domain"
	"termtype/internal/themes"
	"termtype/internal/ui"
)

var version = "dev"

type gameMode struct {
	name  string
	limit time.Duration
}

var gameModes = []gameMode{
	{"Normal", 0},
	{"Time Attack (30s)", 30 * time.Second},
	{"Time Attack (60s)", 60 * time.Second},
}

// resolveASCII decides whether to use the plain-ASCII glyph set. An explicit
// --ascii flag or TERMTYPE_ASCII env var wins; otherwise we keep Unicode unless
// the locale is clearly not UTF-8 (e.g. "C"/"POSIX"), since most modern
// terminals are UTF-8 even when the locale is unset.
func resolveASCII(flagSet bool) bool {
	if flagSet {
		return true
	}
	switch strings.ToLower(os.Getenv("TERMTYPE_ASCII")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	loc := os.Getenv("LC_ALL")
	if loc == "" {
		loc = os.Getenv("LC_CTYPE")
	}
	if loc == "" {
		loc = os.Getenv("LANG")
	}
	if loc == "" {
		return false // unknown locale → assume a modern UTF-8 terminal
	}
	loc = strings.ToLower(loc)
	return !strings.Contains(loc, "utf-8") && !strings.Contains(loc, "utf8")
}

type language struct {
	name      string
	sentences []string
}

var languages = []language{
	{"English", domain.Sentences},
	{"한국어 (Korean)", domain.KoreanSentences},
}

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	col := x
	for _, r := range []rune(text) {
		s.SetContent(col, y, r, nil, style)
		col += runewidth.RuneWidth(r)
	}
}

func selectTheme(s tcell.Screen) (domain.Theme, time.Duration, []string, error) {
	var themeNames []string
	for name := range themes.Themes {
		themeNames = append(themeNames, name)
	}
	sort.Slice(themeNames, func(i, j int) bool {
		if themeNames[i] == "log" {
			return true
		}
		if themeNames[j] == "log" {
			return false
		}
		return themeNames[i] < themeNames[j]
	})

	selectedIndex := 0
	modeIndex := 0
	langIndex := 0

	for {
		s.Clear()
		drawText(s, 2, 1, tcell.StyleDefault.Bold(true), "Select a theme:")

		for i, name := range themeNames {
			style := tcell.StyleDefault
			if i == selectedIndex {
				style = style.Reverse(true)
			}
			drawText(s, 4, 3+i, style, name)
		}

		gl := ui.Glyphs()
		w, _ := s.Size()
		modeRow := 3 + len(themeNames) + 1
		drawText(s, 2, modeRow, tcell.StyleDefault.Foreground(tcell.ColorYellow),
			"Mode: "+gameModes[modeIndex].name)
		drawText(s, 2, modeRow+1, tcell.StyleDefault.Foreground(tcell.ColorTeal),
			"Language: "+languages[langIndex].name)

		// Pick the widest help line that fits the terminal.
		sep := " " + gl.Sep + " "
		full := strings.Join([]string{
			gl.ArrowUD + " theme", "Tab mode", gl.ArrowLR + " language",
			gl.Enter + " start", "Esc quit",
		}, sep)
		compact := strings.Join([]string{gl.ArrowUD + " theme", "Tab mode", gl.ArrowLR + " lang"}, " ")
		help := full
		if runewidth.StringWidth(help) > w-2 {
			help = compact
		}
		if runewidth.StringWidth(help) > w-2 {
			help = gl.Enter + " start"
		}
		drawText(s, 2, modeRow+3, tcell.StyleDefault.Foreground(tcell.ColorGray), help)
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return nil, 0, nil, fmt.Errorf("theme selection cancelled")
			case tcell.KeyUp:
				if selectedIndex > 0 {
					selectedIndex--
				}
			case tcell.KeyDown:
				if selectedIndex < len(themeNames)-1 {
					selectedIndex++
				}
			case tcell.KeyTab:
				modeIndex = (modeIndex + 1) % len(gameModes)
			case tcell.KeyLeft, tcell.KeyRight:
				langIndex = (langIndex + 1) % len(languages)
			case tcell.KeyEnter:
				return themes.Themes[themeNames[selectedIndex]], gameModes[modeIndex].limit, languages[langIndex].sentences, nil
			}
		}
	}
}

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	vFlag := flag.Bool("v", false, "Print version information (shorthand)")
	asciiFlag := flag.Bool("ascii", false, "Use ASCII-only symbols (for terminals/fonts that can't render the Unicode glyphs)")
	flag.Parse()

	if *versionFlag || *vFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	ui.SetASCII(resolveASCII(*asciiFlag))

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	// Restore the terminal on any return path or panic in the game loop.
	defer s.Fini()

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s.SetStyle(defStyle)
	s.EnablePaste()
	s.Clear()

	// Select theme, mode, and language
	theme, timeLimit, sentences, err := selectTheme(s)
	if err != nil {
		return // user cancelled; the deferred Fini restores the terminal
	}

	// Create and run the game
	game, err := app.NewGame(s, theme, timeLimit, sentences)
	if err != nil {
		return
	}

	game.Run()
}
