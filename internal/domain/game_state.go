package domain

import (
	"time"
)

// GameState holds the state of a typing round.
type GameState struct {
	Sentences      []string
	TargetSentence string
	UserInput      string
	StartTime      time.Time
	TimerStarted   bool
	IsFinished     bool
	Wpm            float64
	Accuracy       float64

	// Mode
	TimeLimit time.Duration // 0 = normal; > 0 = time-attack countdown
	TimedOut  bool          // time attack: the limit was reached before completion

	// Pause
	Paused      bool
	pausedAt    time.Time
	pausedTotal time.Duration

	// Per-theme custom state
	CustomState interface{}
}

// ResetCommon resets the per-round state. The mode (TimeLimit) is preserved.
func (gs *GameState) ResetCommon() {
	gs.UserInput = ""
	gs.TimerStarted = false
	gs.IsFinished = false
	gs.Wpm = 0
	gs.Accuracy = 0
	gs.TimedOut = false
	gs.Paused = false
	gs.pausedTotal = 0
}

// Elapsed returns active typing time, excluding any paused spans.
func (gs *GameState) Elapsed() time.Duration {
	if !gs.TimerStarted {
		return 0
	}
	end := time.Now()
	if gs.Paused {
		end = gs.pausedAt
	}
	d := end.Sub(gs.StartTime) - gs.pausedTotal
	if d < 0 {
		return 0
	}
	return d
}

// Remaining returns the time left in time-attack mode (0 in normal mode).
func (gs *GameState) Remaining() time.Duration {
	if gs.TimeLimit <= 0 {
		return 0
	}
	r := gs.TimeLimit - gs.Elapsed()
	if r < 0 {
		return 0
	}
	return r
}

// TogglePause pauses or resumes the active timer. It is a no-op before typing
// starts or after the round is finished.
func (gs *GameState) TogglePause() {
	if gs.IsFinished || !gs.TimerStarted {
		return
	}
	if gs.Paused {
		gs.pausedTotal += time.Since(gs.pausedAt)
		gs.Paused = false
	} else {
		gs.pausedAt = time.Now()
		gs.Paused = true
	}
}

// Finalize ends the round and computes WPM and accuracy from the current input.
// It works for both a fully typed sentence and a partial time-attack result.
func (gs *GameState) Finalize() {
	inputRunes := []rune(gs.UserInput)
	targetRunes := []rune(gs.TargetSentence)
	if len(inputRunes) > len(targetRunes) {
		inputRunes = inputRunes[:len(targetRunes)]
		gs.UserInput = string(inputRunes)
	}

	duration := gs.Elapsed().Minutes()
	if duration > 0 {
		gs.Wpm = (float64(len(inputRunes)) / 5.0) / duration
	}

	correct := 0
	for i := 0; i < len(inputRunes) && i < len(targetRunes); i++ {
		if inputRunes[i] == targetRunes[i] {
			correct++
		}
	}
	if len(inputRunes) > 0 {
		gs.Accuracy = (float64(correct) / float64(len(inputRunes))) * 100
	}

	gs.IsFinished = true
}
