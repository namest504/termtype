package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/app"
	"termtype/internal/domain"
	"termtype/internal/themes"
)

var version = "dev"

type gameMode struct {
	name  string
	limit time.Duration
}

var gameModes = []gameMode{
	{"Normal", 0},
	{"Time Attack · 30s", 30 * time.Second},
	{"Time Attack · 60s", 60 * time.Second},
}

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range []rune(text) {
		s.SetContent(x+i, y, r, nil, style)
	}
}

func selectTheme(s tcell.Screen) (domain.Theme, time.Duration, error) {
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

		modeRow := 3 + len(themeNames) + 1
		drawText(s, 2, modeRow, tcell.StyleDefault.Foreground(tcell.ColorYellow),
			"Mode: "+gameModes[modeIndex].name)
		drawText(s, 2, modeRow+2, tcell.StyleDefault.Foreground(tcell.ColorGray),
			"↑/↓ theme · Tab mode · Enter start · Esc quit")
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return nil, 0, fmt.Errorf("theme selection cancelled")
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
			case tcell.KeyEnter:
				return themes.Themes[themeNames[selectedIndex]], gameModes[modeIndex].limit, nil
			}
		}
	}
}

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	vFlag := flag.Bool("v", false, "Print version information (shorthand)")
	flag.Parse()

	if *versionFlag || *vFlag {
		fmt.Println(version)
		os.Exit(0)
	}

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

	// Select theme and mode
	theme, timeLimit, err := selectTheme(s)
	if err != nil {
		return // user cancelled; the deferred Fini restores the terminal
	}

	// Create and run the game
	game, err := app.NewGame(s, theme, timeLimit)
	if err != nil {
		return
	}

	game.Run()
}
