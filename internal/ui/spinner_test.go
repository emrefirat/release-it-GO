package ui

import (
	"testing"
	"time"
)

func TestNewSpinner(t *testing.T) {
	s := NewSpinner(false)
	if s == nil {
		t.Fatal("expected non-nil spinner")
	}
	if s.isCI {
		t.Error("expected isCI to be false")
	}

	s = NewSpinner(true)
	if !s.isCI {
		t.Error("expected isCI to be true")
	}
}

func TestSpinner_StartStop_CI(t *testing.T) {
	s := NewSpinner(true)
	s.Start("Testing")

	if !s.active {
		t.Error("expected spinner to be active")
	}

	s.Stop(true)

	if s.active {
		t.Error("expected spinner to be inactive")
	}
}

func TestSpinner_StartStop_NonCI(t *testing.T) {
	s := NewSpinner(false)
	s.Start("Testing")

	if !s.active {
		t.Error("expected spinner to be active")
	}

	// Give spinner goroutine time to start
	time.Sleep(100 * time.Millisecond)

	s.Stop(true)

	if s.active {
		t.Error("expected spinner to be inactive")
	}
}

func TestSpinner_StopFailure(t *testing.T) {
	s := NewSpinner(true)
	s.Start("Testing")
	s.Stop(false)

	if s.active {
		t.Error("expected spinner to be inactive")
	}
}

func TestSpinner_StopWithoutStart(t *testing.T) {
	s := NewSpinner(true)
	// Should not panic
	s.Stop(true)
}

func TestSpinner_Update(t *testing.T) {
	s := NewSpinner(true)
	s.Start("Initial message")
	s.Update("Updated message")

	s.mu.Lock()
	msg := s.message
	s.mu.Unlock()

	if msg != "Updated message" {
		t.Errorf("expected 'Updated message', got %q", msg)
	}

	s.Stop(true)
}
