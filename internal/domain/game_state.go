package domain

import (
	"time"
)

// GameState is the struct that manages the game state
type GameState struct {
	Sentences      []string
	TargetSentence string
	UserInput      string
	StartTime      time.Time
	TimerStarted   bool
	IsFinished     bool
	Wpm            float64
	Accuracy       float64

	// Per-theme custom state
	CustomState interface{}
}

// ResetCommon holds the common reset logic
func (gs *GameState) ResetCommon() {
	gs.UserInput = ""
	gs.TimerStarted = false
	gs.IsFinished = false
	gs.Wpm = 0
	gs.Accuracy = 0
}
