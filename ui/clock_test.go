package ui

import (
	"testing"
	"time"
)

func TestDisplayedSecondsUsesTimestamps(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Second)

	if got := displayedSeconds(false, start, time.Time{}, start.Add(2500*time.Millisecond)); got != 2 {
		t.Fatalf("count up = %d, want 2", got)
	}
	// A late refresh still reports the real elapsed second, not "one more than last time".
	if got := displayedSeconds(false, start, time.Time{}, start.Add(3200*time.Millisecond)); got != 3 {
		t.Fatalf("late count up = %d, want 3", got)
	}
	if got := displayedSeconds(true, start, end, start.Add(1500*time.Millisecond)); got != 9 {
		t.Fatalf("count down = %d, want 9", got)
	}
	if got := displayedSeconds(true, start, end, end); got != 0 {
		t.Fatalf("at end = %d, want 0", got)
	}
	if got := displayedSeconds(true, start, end, end.Add(2*time.Second)); got != 0 {
		t.Fatalf("after end = %d, want 0", got)
	}
}

func TestCountdownPercent(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(10 * time.Second)
	if got := countdownPercent(start, end, start.Add(2500*time.Millisecond)); got != 25 {
		t.Fatalf("percent = %d, want 25", got)
	}
	if got := countdownPercent(start, end, end.Add(time.Second)); got != 100 {
		t.Fatalf("after end percent = %d, want 100", got)
	}
}

func TestStepClockSeconds(t *testing.T) {
	if got := stepClockSeconds(59, 1, 1); got != 60 {
		t.Fatalf("seconds carry = %d, want 60", got)
	}
	if got := stepClockSeconds(59*60, 60, 1); got != 3600 {
		t.Fatalf("minutes carry = %d, want 3600", got)
	}
	if got := stepClockSeconds(3600, 1, -1); got != 3599 {
		t.Fatalf("borrow = %d, want 3599", got)
	}
	if got := stepClockSeconds(0, 1, -1); got != 0 {
		t.Fatalf("below zero = %d, want 0", got)
	}
	if got := stepClockSeconds(clockMaxSeconds, 1, 1); got != clockMaxSeconds {
		t.Fatalf("above a day = %d, want max", got)
	}
}
