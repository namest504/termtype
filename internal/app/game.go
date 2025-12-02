package app

import (
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

// Game 전체를 관리하는 구조체
type Game struct {
	screen   tcell.Screen
	renderer *ui.Renderer
	state    *domain.GameState
	theme    domain.Theme
}

// 새로운 게임 생성
func NewGame(s tcell.Screen, theme domain.Theme) (*Game, error) {
	state := &domain.GameState{Sentences: domain.Sentences}
	theme.ResetState(state)

	return &Game{screen: s, renderer: ui.NewRenderer(s), state: state, theme: theme}, nil
}

// 게임 실행 (실시간 Ticker 포함)
func (g *Game) Run() {
	ticker := time.NewTicker(1 * time.Second) // 1초마다 Tick
	defer ticker.Stop()

	eventChan := make(chan tcell.Event)
	go func() {
		for {
			eventChan <- g.screen.PollEvent()
		}
	}()

	g.theme.UpdateScreen(g.renderer, g.state)

	for {
		select {
		case ev := <-eventChan:
			switch ev := ev.(type) {
			case *tcell.EventResize:
				g.screen.Sync()
				g.theme.UpdateScreen(g.renderer, g.state)
			case *tcell.EventKey:
				g.handleKeyEvent(ev)
				g.theme.UpdateScreen(g.renderer, g.state)
			}
		case <-ticker.C:
			if !g.state.IsFinished {
				g.theme.OnTick(g.state)
				g.theme.UpdateScreen(g.renderer, g.state)
			}
		}
	}
}

// 키 이벤트 처리
func (g *Game) handleKeyEvent(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyEscape {
		g.screen.Fini()
		os.Exit(0)
	}

	if g.state.IsFinished {
		if ev.Key() == tcell.KeyEnter {
			g.theme.ResetState(g.state)
		}
		return
	}

	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(g.state.UserInput) > 0 {
			runes := []rune(g.state.UserInput)
			g.state.UserInput = string(runes[:len(runes)-1])
		}
	case tcell.KeyRune:
		if !g.state.TimerStarted {
			g.state.StartTime = time.Now()
			g.state.TimerStarted = true
		}
		g.state.UserInput += string(ev.Rune())
	}

	// 타이핑 완료 체크
	if len(g.state.UserInput) >= len(g.state.TargetSentence) {
		g.state.IsFinished = true
		duration := time.Since(g.state.StartTime).Minutes()

		if len(g.state.UserInput) > len(g.state.TargetSentence) {
			g.state.UserInput = g.state.UserInput[:len(g.state.TargetSentence)]
		}

		if duration > 0 {
			g.state.Wpm = (float64(len(g.state.UserInput)) / 5.0) / duration
		}

		correctChars := 0
		for i, r := range []rune(g.state.TargetSentence) {
			if i < len([]rune(g.state.UserInput)) && []rune(g.state.UserInput)[i] == r {
				correctChars++
			}
		}
		g.state.Accuracy = (float64(correctChars) / float64(len(g.state.TargetSentence))) * 100
	}
}