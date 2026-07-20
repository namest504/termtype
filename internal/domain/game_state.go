package domain

import (
	"math/rand"
	"time"
)

// GameState holds the state of a typing round.
type GameState struct {
	Sentences      []string
	TargetGen      func() string // when set, targets come from here, not the pool
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

// RandomSentence returns the next typing target: the TargetGen stream when
// one is set (the "words" source), otherwise a random pick from the game's
// sentence pool. It falls back to the default English pool if no pool has
// been set, so themes can pick a sentence without knowing which language the
// game was started in.
func (gs *GameState) RandomSentence() string {
	if gs.TargetGen != nil {
		return gs.TargetGen()
	}
	pool := gs.Sentences
	if len(pool) == 0 {
		pool = Sentences
	}
	return pool[rand.Intn(len(pool))]
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

// computeStats returns WPM and accuracy for the current input without mutating
// state. Over-typed input is measured only up to the target length. WPM uses the
// standard 5-characters-per-word convention.
func (gs *GameState) computeStats() (wpm, accuracy float64) {
	inputRunes := []rune(gs.UserInput)
	targetRunes := []rune(gs.TargetSentence)
	n := len(inputRunes)
	if n > len(targetRunes) {
		n = len(targetRunes)
		inputRunes = inputRunes[:n]
	}

	if duration := gs.Elapsed().Minutes(); duration > 0 {
		wpm = (float64(n) / 5.0) / duration
	}

	correct := 0
	for i := 0; i < n && i < len(targetRunes); i++ {
		if inputRunes[i] == targetRunes[i] {
			correct++
		}
	}
	if n > 0 {
		accuracy = (float64(correct) / float64(n)) * 100
	}
	return wpm, accuracy
}

// LiveStats returns the in-progress WPM and accuracy, for display while typing.
func (gs *GameState) LiveStats() (wpm, accuracy float64) {
	return gs.computeStats()
}

// Finalize ends the round and computes WPM and accuracy from the current input.
// It works for both a fully typed sentence and a partial time-attack result.
func (gs *GameState) Finalize() {
	inputRunes := []rune(gs.UserInput)
	targetRunes := []rune(gs.TargetSentence)
	if len(inputRunes) > len(targetRunes) {
		gs.UserInput = string(inputRunes[:len(targetRunes)])
	}

	gs.Wpm, gs.Accuracy = gs.computeStats()
	gs.IsFinished = true
}
