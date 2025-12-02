package domain

import (
	"time"
)

// GameState 게임 상태를 관리하는 구조체
type GameState struct {
	Sentences      []string
	TargetSentence string
	UserInput      string
	StartTime      time.Time
	TimerStarted   bool
	IsFinished     bool
	Wpm            float64
	Accuracy       float64

	// 테마별 커스텀 상태
	CustomState interface{}
}

// ResetCommon 공통 리셋 로직
func (gs *GameState) ResetCommon() {
	gs.UserInput = ""
	gs.TimerStarted = false
	gs.IsFinished = false
	gs.Wpm = 0
	gs.Accuracy = 0
}
