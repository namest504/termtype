package domain

import (
	"testing"
	"time"
)

// Finalize over a partial (time-attack) input uses the typed length as the
// accuracy denominator.
func TestFinalize_PartialAccuracy(t *testing.T) {
	gs := &GameState{
		TargetSentence: "abcdefgh",
		UserInput:      "abXd", // typed 4, correct a,b,_,d = 3
		TimerStarted:   true,
		StartTime:      time.Now().Add(-2 * time.Second),
	}
	gs.Finalize()
	if !gs.IsFinished {
		t.Fatal("Finalize should mark the round finished")
	}
	if gs.Accuracy != 75 {
		t.Errorf("Accuracy = %.2f, want 75 (3 of 4 typed)", gs.Accuracy)
	}
	if gs.Wpm <= 0 {
		t.Errorf("Wpm = %.2f, want > 0", gs.Wpm)
	}
}

// Remaining counts down from the limit in time-attack mode.
func TestRemaining_TimeAttack(t *testing.T) {
	gs := &GameState{
		TimeLimit:    60 * time.Second,
		TimerStarted: true,
		StartTime:    time.Now().Add(-10 * time.Second),
	}
	r := gs.Remaining()
	if r < 49*time.Second || r > 51*time.Second {
		t.Errorf("Remaining = %v, want ~50s", r)
	}
}

// Normal mode has no countdown.
func TestRemaining_NormalIsZero(t *testing.T) {
	gs := &GameState{TimerStarted: true, StartTime: time.Now()}
	if r := gs.Remaining(); r != 0 {
		t.Errorf("Remaining = %v, want 0 in normal mode", r)
	}
}

// Pausing excludes the paused span from elapsed time.
func TestPause_ExcludesPausedTime(t *testing.T) {
	gs := &GameState{TimerStarted: true, StartTime: time.Now()}
	gs.TogglePause()
	time.Sleep(25 * time.Millisecond)
	gs.TogglePause()

	if gs.pausedTotal < 20*time.Millisecond {
		t.Errorf("pausedTotal = %v, want >= 20ms", gs.pausedTotal)
	}
	if gs.Elapsed() > 15*time.Millisecond {
		t.Errorf("Elapsed = %v, should exclude the paused span", gs.Elapsed())
	}
}

// Pause is a no-op before typing starts.
func TestPause_NoOpBeforeStart(t *testing.T) {
	gs := &GameState{}
	gs.TogglePause()
	if gs.Paused {
		t.Error("should not pause before the timer has started")
	}
}
