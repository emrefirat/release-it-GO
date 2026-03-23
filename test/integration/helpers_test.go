package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGitRepo creates a new git repo in the given directory with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	// Create initial file and commit
	writeFile(t, filepath.Join(dir, "README.md"), "# Test Project\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "chore: initial commit")
}

// createCommits creates commits with the given messages in the repo.
// Each commit creates a unique file to ensure there are changes to commit.
func createCommits(t *testing.T, dir string, messages []string) {
	t.Helper()
	for _, msg := range messages {
		filename := fmt.Sprintf("file_%d_%d.txt", os.Getpid(), commitCounter)
		commitCounter++
		writeFile(t, filepath.Join(dir, filename), fmt.Sprintf("content for: %s\n", msg))
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-m", msg)
	}
}

// commitCounter ensures unique file names across calls within the same test.
var commitCounter int

// createTag creates a git tag in the repo.
func createTag(t *testing.T, dir string, tag string) {
	t.Helper()
	runGit(t, dir, "tag", "-a", tag, "-m", fmt.Sprintf("Release %s", tag))
}

// assertTagExists verifies a tag exists in the repo.
func assertTagExists(t *testing.T, dir string, tag string) {
	t.Helper()
	out := runGit(t, dir, "tag", "-l", tag)
	if strings.TrimSpace(out) != tag {
		t.Errorf("expected tag %q to exist, got %q", tag, out)
	}
}

// assertTagNotExists verifies a tag does not exist in the repo.
func assertTagNotExists(t *testing.T, dir string, tag string) {
	t.Helper()
	out := runGit(t, dir, "tag", "-l", tag)
	if strings.TrimSpace(out) == tag {
		t.Errorf("expected tag %q to not exist", tag)
	}
}

// assertChangelogContains checks that CHANGELOG.md contains the expected string.
func assertChangelogContains(t *testing.T, dir string, expected string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("failed to read CHANGELOG.md: %v", err)
	}
	if !strings.Contains(string(content), expected) {
		t.Errorf("CHANGELOG.md does not contain %q, got:\n%s", expected, string(content))
	}
}

// getLatestCommitMsg returns the latest commit message in the repo.
func getLatestCommitMsg(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(runGit(t, dir, "log", "-1", "--pretty=format:%s"))
}

// runGit executes a git command in the given directory and returns stdout.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

// assertFileContains checks that a file contains the expected string.
func assertFileContains(t *testing.T, path string, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if !strings.Contains(string(content), expected) {
		t.Errorf("%s does not contain %q", filepath.Base(path), expected)
	}
}

// assertFileNotExists checks that a file does not exist.
func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Errorf("expected %s to not exist", filepath.Base(path))
	}
}

// assertWorkingDirClean checks that git status is clean.
func assertWorkingDirClean(t *testing.T, dir string) {
	t.Helper()
	out := runGit(t, dir, "status", "--porcelain")
	if strings.TrimSpace(out) != "" {
		t.Errorf("expected clean working directory, got:\n%s", out)
	}
}

// assertTagCount checks the number of tags in the repo.
func assertTagCount(t *testing.T, dir string, expected int) {
	t.Helper()
	out := runGit(t, dir, "tag", "-l")
	tags := strings.Split(strings.TrimSpace(out), "\n")
	if tags[0] == "" {
		tags = []string{}
	}
	if len(tags) != expected {
		t.Errorf("expected %d tags, got %d: %v", expected, len(tags), tags)
	}
}

// writeFile writes content to a file, creating directories as needed.
func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}
