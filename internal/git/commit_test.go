package git

import (
	"strings"
	"testing"

	"release-it-go/internal/config"
)

func TestStage_UpdateOnly(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{AddUntrackedFiles: false}
	g := newTestGitWithConfig(cfg, false)

	err := g.Stage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "--update") {
		t.Errorf("expected --update flag, got args: %v", executedArgs)
	}
}

func TestStage_WithUntracked(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{AddUntrackedFiles: true}
	g := newTestGitWithConfig(cfg, false)

	err := g.Stage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if strings.Contains(cmd, "--update") {
		t.Errorf("should not contain --update when AddUntrackedFiles=true, got: %v", executedArgs)
	}
	if !strings.Contains(cmd, "add .") {
		t.Errorf("expected 'add .', got: %v", executedArgs)
	}
}

func TestStage_DryRun(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	called := false
	commandExecutor = func(name string, args ...string) (string, error) {
		called = true
		return "", nil
	}

	cfg := &config.GitConfig{AddUntrackedFiles: false}
	g := newTestGitWithConfig(cfg, true)

	err := g.Stage()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected stage to be skipped in dry-run mode")
	}
}

func TestCommit(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	err := g.Commit("Release 1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "--message") {
		t.Error("expected --message flag")
	}
	if !strings.Contains(cmd, "Release 1.0.0") {
		t.Error("expected commit message in args")
	}
}

func TestCommit_WithArgs(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{
		CommitArgs: []string{"--no-verify", "--signoff"},
	}
	g := newTestGitWithConfig(cfg, false)

	err := g.Commit("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "--no-verify") {
		t.Error("expected --no-verify in args")
	}
	if !strings.Contains(cmd, "--signoff") {
		t.Error("expected --signoff in args")
	}
}

func TestStageFile(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	err := g.StageFile("CHANGELOG.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "add CHANGELOG.md") {
		t.Errorf("expected 'add CHANGELOG.md', got: %v", executedArgs)
	}
}

func TestCommit_DryRun(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	called := false
	commandExecutor = func(name string, args ...string) (string, error) {
		called = true
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, true)

	err := g.Commit("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected commit to be skipped in dry-run mode")
	}
}
