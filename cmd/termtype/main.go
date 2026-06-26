package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/app"
	"termtype/internal/domain"
	"termtype/internal/themes"
)

var version = "dev"

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range []rune(text) {
		s.SetContent(x+i, y, r, nil, style)
	}
}

func selectTheme(s tcell.Screen) (domain.Theme, error) {
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

	for {
		s.Clear()
		drawText(s, 2, 1, tcell.StyleDefault, "Select a theme:")

		for i, name := range themeNames {
			style := tcell.StyleDefault
			if i == selectedIndex {
				style = style.Reverse(true)
			}
			drawText(s, 4, 3+i, style, name)
		}
		s.Show()

		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return nil, fmt.Errorf("theme selection cancelled")
			case tcell.KeyUp:
				if selectedIndex > 0 {
					selectedIndex--
				}
			case tcell.KeyDown:
				if selectedIndex < len(themeNames)-1 {
					selectedIndex++
				}
			case tcell.KeyEnter:
				return themes.Themes[themeNames[selectedIndex]], nil
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
	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s.SetStyle(defStyle)
	s.EnablePaste()
	s.Clear()

	// Select theme
	theme, err := selectTheme(s)
	if err != nil {
		s.Fini()
		fmt.Println(err)
		return
	}

	// Create and run the game
	game, err := app.NewGame(s, theme)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	game.Run()
}
