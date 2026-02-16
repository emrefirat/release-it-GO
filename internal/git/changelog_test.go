package git

import (
	"strings"
	"testing"

	"release-it-go/internal/config"
)

func TestGenerateChangelog(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "* feat: add feature (abc1234)\n* fix: bug fix (def5678)", nil
	}

	cfg := &config.GitConfig{Changelog: "* %s (%h)"}
	g := newTestGitWithConfig(cfg, false)

	changelog, err := g.GenerateChangelog("v1.0.0", "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(changelog, "feat: add feature") {
		t.Errorf("expected changelog to contain commit message, got: %s", changelog)
	}
}

func TestGenerateChangelog_DefaultFormat(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "* some commit (abc1234)", nil
	}

	cfg := &config.GitConfig{} // empty changelog format
	g := newTestGitWithConfig(cfg, false)

	_, err := g.GenerateChangelog("v1.0.0", "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(capturedArgs, " ")
	if !strings.Contains(cmd, "* %s (%h)") {
		t.Errorf("expected default format, got args: %v", capturedArgs)
	}
}

func TestGenerateChangelog_NoFromTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "* initial commit", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	_, err := g.GenerateChangelog("", "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(capturedArgs, " ")
	// Should use just "HEAD" not "..HEAD"
	if strings.Contains(cmd, "..HEAD") {
		t.Errorf("should not contain range when fromTag is empty, got: %v", capturedArgs)
	}
}

func TestGenerateChangelog_DefaultToRef(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	_, err := g.GenerateChangelog("v1.0.0", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(capturedArgs, " ")
	if !strings.Contains(cmd, "v1.0.0..HEAD") {
		t.Errorf("expected default toRef=HEAD, got args: %v", capturedArgs)
	}
}

func TestGetCommitsWithHashSinceTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "abc1234||feat: add feature\ndef5678||fix: bug fix\nghi9012||chore: update deps", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	commits, err := g.GetCommitsWithHashSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(commits))
	}
	if commits[0].Hash != "abc1234" {
		t.Errorf("expected hash 'abc1234', got %q", commits[0].Hash)
	}
	if commits[0].Subject != "feat: add feature" {
		t.Errorf("expected subject 'feat: add feature', got %q", commits[0].Subject)
	}
	if commits[2].Hash != "ghi9012" {
		t.Errorf("expected hash 'ghi9012', got %q", commits[2].Hash)
	}
}

func TestGetCommitsWithHashSinceTag_NoTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "abc1234||initial commit", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	commits, err := g.GetCommitsWithHashSinceTag("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(capturedArgs, " ")
	if strings.Contains(cmd, "..HEAD") {
		t.Error("should not contain range when tag is empty")
	}
	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
	if commits[0].Hash != "abc1234" {
		t.Errorf("expected hash 'abc1234', got %q", commits[0].Hash)
	}
}

func TestGetCommitsWithHashSinceTag_Empty(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	commits, err := g.GetCommitsWithHashSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commits != nil {
		t.Errorf("expected nil for no commits, got %v", commits)
	}
}

func TestGetCommitsSinceTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "feat: add feature\nfix: bug fix\nchore: update deps", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	commits, err := g.GetCommitsSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 3 {
		t.Errorf("expected 3 commits, got %d", len(commits))
	}
	if commits[0] != "feat: add feature" {
		t.Errorf("expected first commit 'feat: add feature', got %q", commits[0])
	}
}

func TestGetCommitsSinceTag_NoTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "initial commit", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	commits, err := g.GetCommitsSinceTag("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(capturedArgs, " ")
	if strings.Contains(cmd, "..HEAD") {
		t.Error("should not contain range when tag is empty")
	}
	if len(commits) != 1 {
		t.Errorf("expected 1 commit, got %d", len(commits))
	}
}

func TestGetCommitsSinceTag_Empty(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	commits, err := g.GetCommitsSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commits != nil {
		t.Errorf("expected nil for no commits, got %v", commits)
	}
}
