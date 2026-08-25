package ui

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// spinnerFrames defines the animation frames for the spinner.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner provides a simple terminal spinner for long-running operations.
type Spinner struct {
	message string
	active  bool
	mu      sync.Mutex
	done    chan struct{}
	isCI    bool
}

// NewSpinner creates a new spinner instance.
func NewSpinner(isCI bool) *Spinner {
	return &Spinner{
		isCI: isCI,
	}
}

// Start begins the spinner animation with the given message.
func (s *Spinner) Start(message string) {
	s.mu.Lock()

	// A previous animator may still be running if Stop was skipped; close its
	// channel so two goroutines never animate at once.
	if s.active && s.done != nil {
		close(s.done)
	}

	s.message = message
	s.active = true
	done := make(chan struct{})
	s.done = done
	isCI := s.isCI
	s.mu.Unlock()

	// In CI mode, don't print start message; only Stop() prints the result line
	if isCI {
		return
	}

	// The goroutine captures its own done channel — re-reading s.done without
	// the mutex raced with Start's reassignment.
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			s.mu.Lock()
			msg := s.message
			s.mu.Unlock()

			frame := spinnerFrames[i%len(spinnerFrames)]
			fmt.Fprintf(os.Stderr, "\r%s %s...", frame, msg)
			i++

			select {
			case <-done:
				return
			case <-ticker.C:
			}
		}
	}()
}

// Stop stops the spinner and shows a success or failure indicator.
func (s *Spinner) Stop(success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	close(s.done)

	if s.isCI {
		if success {
			fmt.Fprintf(os.Stderr, "  %s %s\n", FormatSuccess(IconSuccess), s.message)
		} else {
			fmt.Fprintf(os.Stderr, "  %s %s\n", FormatError(IconFail), s.message)
		}
		return
	}

	if success {
		fmt.Fprintf(os.Stderr, "\r%s %s\n", FormatSuccess("✓"), s.message)
	} else {
		fmt.Fprintf(os.Stderr, "\r%s %s\n", FormatError("✗"), s.message)
	}
}

// Update changes the spinner message while it's running.
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}
