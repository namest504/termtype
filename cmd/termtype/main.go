package main

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/app"
)

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	for i, r := range []rune(text) {
		s.SetContent(x+i, y, r, nil, style)
	}
}

func selectTheme(s tcell.Screen) (app.Theme, error) {
	var themes []string
	for name := range app.Themes {
		themes = append(themes, name)
	}
	sort.Strings(themes)

	selectedIndex := 0

	for {
		s.Clear()
		drawText(s, 2, 1, tcell.StyleDefault, "Select a theme:")

		for i, name := range themes {
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
				if selectedIndex < len(themes)-1 {
					selectedIndex++
				}
			case tcell.KeyEnter:
				return app.Themes[themes[selectedIndex]], nil
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// 화면 초기화
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

	// 테마 선택
	theme, err := selectTheme(s)
	if err != nil {
		s.Fini()
		fmt.Println(err)
		return
	}

	// 게임 생성 및 실행
	game, err := app.NewGame(s, theme)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	game.Run()
}