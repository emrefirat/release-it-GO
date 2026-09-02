package git

import (
	"fmt"
	"strings"
	"testing"

	"github.com/emrefirat/release-it-GO/internal/config"
)

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

func TestGetFullCommitsSinceTag_ParsesMultilineBodies(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "abc1234\x1ffeat: change API\n\nBREAKING CHANGE: removed\nold endpoints\x1e\ndef5678\x1ffix: bug\x1e", nil
	}

	g := newTestGitWithConfig(&config.GitConfig{}, false)
	commits, err := g.GetFullCommitsSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"log", "v1.0.0..HEAD", "--pretty=format:%h%x1f%B%x1e"}
	if strings.Join(capturedArgs, " ") != strings.Join(wantArgs, " ") {
		t.Errorf("args = %v, want %v", capturedArgs, wantArgs)
	}

	if len(commits) != 2 {
		t.Fatalf("len(commits) = %d, want 2", len(commits))
	}
	if commits[0].Hash != "abc1234" {
		t.Errorf("commits[0].Hash = %q, want abc1234", commits[0].Hash)
	}
	wantMsg := "feat: change API\n\nBREAKING CHANGE: removed\nold endpoints"
	if commits[0].Message != wantMsg {
		t.Errorf("commits[0].Message = %q, want %q (multiline body must survive)", commits[0].Message, wantMsg)
	}
	if commits[1].Hash != "def5678" || commits[1].Message != "fix: bug" {
		t.Errorf("commits[1] = %+v, want {def5678 fix: bug}", commits[1])
	}
}

func TestGetFullCommitsSinceTag_NoTag_ReturnsAllCommits(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "abc1234\x1ffeat: first\x1e", nil
	}

	g := newTestGitWithConfig(&config.GitConfig{}, false)
	commits, err := g.GetFullCommitsSinceTag("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, arg := range capturedArgs {
		if strings.Contains(arg, "..HEAD") {
			t.Errorf("empty tag must not produce a range arg, got %v", capturedArgs)
		}
	}
	if len(commits) != 1 || commits[0].Message != "feat: first" {
		t.Errorf("commits = %+v, want one 'feat: first'", commits)
	}
}

func TestGetFullCommitsSinceTag_EmptyOutput(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", nil
	}

	g := newTestGitWithConfig(&config.GitConfig{}, false)
	commits, err := g.GetFullCommitsSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("expected no commits for empty output, got %d", len(commits))
	}
}

// --- GetCommitCountSinceTag tests ---

func TestGetCommitCountSinceTag_WithTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "5", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	count, err := g.GetCommitCountSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

func TestGetCommitCountSinceTag_NoTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		cmd := strings.Join(args, " ")
		if !strings.Contains(cmd, "..") {
			return "12", nil
		}
		return "", fmt.Errorf("unexpected")
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	count, err := g.GetCommitCountSinceTag("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 12 {
		t.Errorf("expected 12, got %d", count)
	}
}

func TestGetCommitCountSinceTag_Zero(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "0", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	count, err := g.GetCommitCountSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestGetCommitCountSinceTag_GitError(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", fmt.Errorf("git error")
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	_, err := g.GetCommitCountSinceTag("v1.0.0")
	if err == nil {
		t.Error("expected error")
	}
}

// --- GetContributorsSinceTag tests ---

func TestGetContributorsSinceTag_WithDuplicates(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "Alice\nBob\nAlice\nCharlie\nBob", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	contributors, err := g.GetContributorsSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contributors) != 3 {
		t.Errorf("expected 3 unique contributors, got %d: %v", len(contributors), contributors)
	}
}

func TestGetContributorsSinceTag_Empty(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	contributors, err := g.GetContributorsSinceTag("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contributors != nil {
		t.Errorf("expected nil, got %v", contributors)
	}
}

func TestGetContributorsSinceTag_NoTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		cmd := strings.Join(args, " ")
		if !strings.Contains(cmd, "..") {
			return "Alice", nil
		}
		return "", fmt.Errorf("unexpected")
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	contributors, err := g.GetContributorsSinceTag("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(contributors) != 1 || contributors[0] != "Alice" {
		t.Errorf("expected [Alice], got %v", contributors)
	}
}

func TestGetContributorsSinceTag_GitError(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", fmt.Errorf("git error")
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	_, err := g.GetContributorsSinceTag("v1.0.0")
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetFullCommitsSinceTag_CommitsPath_ScopesTheLog(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "abc1234\x1ffeat: api\x1e", nil
	}

	// npm's monorepo lever: only commits touching commitsPath count
	g := newTestGitWithConfig(&config.GitConfig{CommitsPath: "packages/api"}, false)
	if _, err := g.GetFullCommitsSinceTag("v1.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	joined := strings.Join(capturedArgs, " ")
	if !strings.HasSuffix(joined, "-- packages/api") {
		t.Errorf("git log must be scoped with a pathspec, got: %s", joined)
	}
}

func TestGetCommitCountSinceTag_CommitsPath_ScopesRevList(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "3", nil
	}

	g := newTestGitWithConfig(&config.GitConfig{CommitsPath: "packages/api"}, false)
	if _, err := g.GetCommitCountSinceTag("v1.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(strings.Join(capturedArgs, " "), "-- packages/api") {
		t.Errorf("rev-list must be scoped with a pathspec, got: %v", capturedArgs)
	}
}
