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
	defer s.mu.Unlock()

	s.message = message
	s.active = true
	s.done = make(chan struct{})

	// In CI mode, don't print start message; only Stop() prints the result line
	if s.isCI {
		return
	}

	go func() {
		i := 0
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()

				frame := spinnerFrames[i%len(spinnerFrames)]
				fmt.Fprintf(os.Stderr, "\r%s %s...", frame, msg)
				i++
				time.Sleep(80 * time.Millisecond)
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
