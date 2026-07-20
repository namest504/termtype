package domain

import (
	"testing"
	"time"
)

func TestSampleWPMGuards(t *testing.T) {
	gs := &GameState{TargetSentence: "abcd"}

	gs.SampleWPM() // before the timer starts
	if len(gs.WPMSamples) != 0 {
		t.Fatal("sampled before typing started")
	}

	gs.TimerStarted = true
	gs.StartTime = time.Now().Add(-2 * time.Second)
	gs.UserInput = "ab"
	gs.SampleWPM()
	if len(gs.WPMSamples) != 1 || gs.WPMSamples[0] <= 0 {
		t.Fatalf("expected one positive sample, got %v", gs.WPMSamples)
	}

	gs.Paused = true
	gs.SampleWPM()
	gs.Paused = false
	gs.IsFinished = true
	gs.SampleWPM()
	if len(gs.WPMSamples) != 1 {
		t.Fatalf("sampled while paused or finished: %v", gs.WPMSamples)
	}
}

func TestFinalizeMetrics(t *testing.T) {
	gs := &GameState{
		TargetSentence: "abcd",
		UserInput:      "abXd", // 3 of 4 correct
		TypedRunes:     6,      // two runes were backspaced along the way
		TimerStarted:   true,
		StartTime:      time.Now().Add(-30 * time.Second),
	}
	gs.Finalize()

	if gs.FinalDurS < 29 || gs.FinalDurS > 31 {
		t.Errorf("FinalDurS = %.1f, want ~30", gs.FinalDurS)
	}
	// raw: 6 runes / 5 per word over half a minute = 2.4
	if gs.FinalRawWPM < 2.3 || gs.FinalRawWPM > 2.5 {
		t.Errorf("FinalRawWPM = %.2f, want ~2.4", gs.FinalRawWPM)
	}
	// cpm: 3 correct runes over half a minute = 6
	if gs.FinalCPM < 5.9 || gs.FinalCPM > 6.1 {
		t.Errorf("FinalCPM = %.2f, want ~6", gs.FinalCPM)
	}
}

func TestResetCommonClearsMetrics(t *testing.T) {
	gs := &GameState{
		TypedRunes:  9,
		WPMSamples:  []float64{60, 70},
		FinalRawWPM: 80,
		FinalCPM:    400,
		FinalDurS:   15,
	}
	gs.ResetCommon()
	if gs.TypedRunes != 0 || gs.WPMSamples != nil ||
		gs.FinalRawWPM != 0 || gs.FinalCPM != 0 || gs.FinalDurS != 0 {
		t.Errorf("metrics not cleared: %+v", gs)
	}
}
