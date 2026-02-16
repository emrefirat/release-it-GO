package git

import (
	"strings"
	"testing"

	"github.com/emfi/release-it-go/internal/config"
)

func TestPush_Default(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := config.DefaultConfig()
	g := newTestGitWithConfig(&cfg.Git, false)

	err := g.Push()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "push") {
		t.Error("expected 'push' in args")
	}
	if !strings.Contains(cmd, "--follow-tags") {
		t.Error("expected default --follow-tags in args")
	}
}

func TestPush_CustomRepo(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{
		PushRepo: "upstream",
		PushArgs: []string{"--follow-tags"},
	}
	g := newTestGitWithConfig(cfg, false)

	err := g.Push()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "upstream") {
		t.Error("expected 'upstream' in args")
	}
}

func TestPush_CustomArgs(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{
		PushArgs: []string{"--force", "--tags"},
	}
	g := newTestGitWithConfig(cfg, false)

	err := g.Push()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "--force") {
		t.Error("expected --force in args")
	}
	if !strings.Contains(cmd, "--tags") {
		t.Error("expected --tags in args")
	}
}

func TestPush_DryRun(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	called := false
	commandExecutor = func(name string, args ...string) (string, error) {
		called = true
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, true)

	err := g.Push()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected push to be skipped in dry-run mode")
	}
}
