package app

import (
	"bufio"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

// 게임 상태를 관리하는 구조체
type GameState struct {
	sentences      []string
	targetSentence string
	userInput      string
	startTime      time.Time
	timerStarted   bool
	isFinished     bool
	wpm            float64
	accuracy       float64

	// 테마별 커스텀 상태
	CustomState interface{}
}

// Game 전체를 관리하는 구조체
type Game struct {
	screen   tcell.Screen
	renderer *Renderer
	state    *GameState
	theme    Theme
}

// 새로운 게임 생성
func NewGame(s tcell.Screen, theme Theme) (*Game, error) {
	sentences, err := loadSentences("configs/sentences.txt")
	if err != nil {
		return nil, err
	}

	state := &GameState{sentences: sentences}
	theme.ResetState(state)

	return &Game{screen: s, renderer: NewRenderer(s), state: state, theme: theme}, nil
}

// 파일에서 문장 불러오기
func loadSentences(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sentences []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			sentences = append(sentences, line)
		}
	}
	return sentences, scanner.Err()
}

// 공통 리셋 로직
func (gs *GameState) resetCommon() {
	gs.userInput = ""
	gs.timerStarted = false
	gs.isFinished = false
	gs.wpm = 0
	gs.accuracy = 0
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
			if !g.state.isFinished {
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

	if g.state.isFinished {
		if ev.Key() == tcell.KeyEnter {
			g.theme.ResetState(g.state)
		}
		return
	}

	switch ev.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(g.state.userInput) > 0 {
			runes := []rune(g.state.userInput)
			g.state.userInput = string(runes[:len(runes)-1])
		}
	case tcell.KeyRune:
		if !g.state.timerStarted {
			g.state.startTime = time.Now()
			g.state.timerStarted = true
		}
		g.state.userInput += string(ev.Rune())
	}

	// 타이핑 완료 체크
	if len(g.state.userInput) >= len(g.state.targetSentence) {
		g.state.isFinished = true
		duration := time.Since(g.state.startTime).Minutes()

		if len(g.state.userInput) > len(g.state.targetSentence) {
			g.state.userInput = g.state.userInput[:len(g.state.targetSentence)]
		}

		if duration > 0 {
			g.state.wpm = (float64(len(g.state.userInput)) / 5.0) / duration
		}

		correctChars := 0
		for i, r := range []rune(g.state.targetSentence) {
			if i < len([]rune(g.state.userInput)) && []rune(g.state.userInput)[i] == r {
				correctChars++
			}
		}
		g.state.accuracy = (float64(correctChars) / float64(len(g.state.targetSentence))) * 100
	}
}