package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/app"
)

func main() {
	// 테마 플래그 설정
	themeFlag := flag.String("theme", "log", "Theme to use (e.g., 'simple', 'log', 'matrix')")
	listThemesFlag := flag.Bool("list-themes", false, "List available themes")
	flag.Parse()

	if *listThemesFlag {
		fmt.Println("Available themes:")
		for name := range app.Themes {
			fmt.Printf("- %s\n", name)
		}
		return
	}

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

	// 플래그에 따라 테마 선택
	theme, ok := app.Themes[*themeFlag]
	if !ok {
		log.Fatalf("Invalid theme: %s. Check available themes.", *themeFlag)
	}

	// 게임 생성 및 실행
	game, err := app.NewGame(s, theme)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	game.Run()
}
