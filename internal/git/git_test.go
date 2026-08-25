package git

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
)

// newTestGit creates a Git instance for testing with default config.
func newTestGit(dryRun bool) *Git {
	cfg := config.DefaultConfig()
	logger := applog.NewLogger(0, false)
	return NewGit(&cfg.Git, logger, dryRun)
}

func TestNewGit(t *testing.T) {
	cfg := config.DefaultConfig()
	logger := applog.NewLogger(0, false)
	g := NewGit(&cfg.Git, logger, true)

	if g == nil {
		t.Fatal("NewGit returned nil")
	}
	if g.dryRun != true {
		t.Error("expected dryRun to be true")
	}
}

func TestIsWriteOperation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		isWrite bool
	}{
		{"empty args", []string{}, false},
		{"status is read-only", []string{"status", "--porcelain"}, false},
		{"log is read-only", []string{"log", "--oneline"}, false},
		{"describe is read-only", []string{"describe", "--tags"}, false},
		{"rev-parse is read-only", []string{"rev-parse", "HEAD"}, false},
		{"remote is read-only", []string{"remote", "-v"}, false},
		{"diff is read-only", []string{"diff"}, false},
		{"show is read-only", []string{"show", "HEAD"}, false},
		{"branch is read-only", []string{"branch"}, false},
		{"commit is write", []string{"commit", "-m", "msg"}, true},
		{"push is write", []string{"push"}, true},
		{"add is write", []string{"add", "."}, true},
		{"tag create is write", []string{"tag", "--annotate", "v1.0.0"}, true},
		{"tag -l is read-only", []string{"tag", "-l"}, false},
		{"tag --list is read-only", []string{"tag", "--list"}, false},
		{"unknown command is write", []string{"unknown-cmd"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWriteOperation(tt.args)
			if got != tt.isWrite {
				t.Errorf("isWriteOperation(%v) = %v, want %v", tt.args, got, tt.isWrite)
			}
		})
	}
}

func TestRun_DryRun(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	called := false
	commandExecutor = func(name string, args ...string) (string, error) {
		called = true
		return "output", nil
	}

	g := newTestGit(true)

	// Write operation should NOT call commandExecutor in dry-run
	_, err := g.run("commit", "-m", "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected write operation to be skipped in dry-run mode")
	}

	// Read operation should still call commandExecutor in dry-run
	called = false
	_, err = g.run("status", "--porcelain")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected read operation to execute in dry-run mode")
	}
}

func TestRun_Normal(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "success", nil
	}

	g := newTestGit(false)
	out, err := g.run("commit", "-m", "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out != "success" {
		t.Errorf("expected 'success', got %q", out)
	}
}

func TestRun_Error(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "error message", fmt.Errorf("exit status 1")
	}

	g := newTestGit(false)
	_, err := g.run("push")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "git push") {
		t.Errorf("error should contain command, got: %v", err)
	}
}

func TestRunSilent(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "  output with spaces  ", nil
	}

	g := newTestGit(false)
	out, err := g.runSilent("status")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if out != "output with spaces" {
		t.Errorf("expected trimmed output, got %q", out)
	}
}

func TestRun_ErrorPreservesRootCauseAndStderr(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	rootErr := fmt.Errorf("exit status 128")
	commandExecutor = func(name string, args ...string) (string, error) {
		return "fatal: not a git repository", rootErr
	}

	g := newTestGit(false)
	_, err := g.run("status")
	if err == nil {
		t.Fatal("expected error")
	}
	// Project convention: errors.Is must reach the root cause (%w), and the
	// user-facing text must carry git's stderr — the old wrappers each kept
	// only one half.
	if !errors.Is(err, rootErr) {
		t.Errorf("errors.Is cannot reach the root cause: %v", err)
	}
	if !strings.Contains(err.Error(), "fatal: not a git repository") {
		t.Errorf("error text lost git stderr: %v", err)
	}
	if !strings.Contains(err.Error(), "git status") {
		t.Errorf("error text lost the command: %v", err)
	}
}

func TestRunSilent_ErrorIncludesStderr(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	rootErr := fmt.Errorf("exit status 1")
	commandExecutor = func(name string, args ...string) (string, error) {
		return "error: pathspec 'x' did not match", rootErr
	}

	g := newTestGit(false)
	_, err := g.runSilent("checkout", "x")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, rootErr) {
		t.Errorf("errors.Is cannot reach the root cause: %v", err)
	}
	if !strings.Contains(err.Error(), "pathspec") {
		t.Errorf("error text lost git stderr: %v", err)
	}
}
