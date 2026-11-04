package git

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/emrefirat/release-it-GO/internal/config"
	"github.com/emrefirat/release-it-GO/internal/testutil"
)

// initScratchRepo creates an empty git repository and makes it the working
// directory for the test. Only the real executor is exercised here, so the
// global/system git config is isolated like the other real-git packages do.
func initScratchRepo(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", "")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "")
	t.Setenv("GIT_TERMINAL_PROMPT", "")
	t.Setenv("GIT_CONFIG_COUNT", "")
	t.Setenv("GIT_CONFIG_KEY_0", "")
	t.Setenv("GIT_CONFIG_VALUE_0", "")
	if err := testutil.IsolateGit(); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	if out, err := exec.Command("git", "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
}

func TestCommandExecutor_StderrIsNotPartOfTheOutput(t *testing.T) {
	initScratchRepo(t)
	// A normal debugging step: every git command then prints trace lines on
	// stderr, which CombinedOutput merged into the parsed result.
	t.Setenv("GIT_TRACE", "1")

	out, err := commandExecutor("git", "rev-parse", "--is-inside-work-tree")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "true" {
		t.Errorf("stdout must be returned alone, got %q", out)
	}
}

func TestCommandExecutor_ErrorCarriesStderr(t *testing.T) {
	initScratchRepo(t)

	out, err := commandExecutor("git", "rev-parse", "--verify", "--quiet", "refs/heads/definitely-missing")
	if err == nil {
		t.Fatal("expected an error for a missing ref")
	}
	if out != "" {
		t.Errorf("stdout should be empty, got %q", out)
	}
	// With --quiet git prints nothing; make sure a loud failure keeps its text.
	_, err = commandExecutor("git", "rev-parse", "--verify", "refs/heads/definitely-missing")
	if err == nil || !strings.Contains(err.Error(), "fatal") {
		t.Errorf("error must include git's stderr, got: %v", err)
	}
}

func TestGit_BranchAndCleanTree_UnaffectedByGitTrace(t *testing.T) {
	initScratchRepo(t)
	t.Setenv("GIT_TRACE", "1")

	g := newTestGitWithConfig(&config.GitConfig{RequireCleanWorkingDir: true}, false)
	if err := g.checkCleanWorkingDir(); err != nil {
		t.Errorf("empty repository must be clean even with GIT_TRACE on, got: %v", err)
	}
	count, err := g.GetCommitCountSinceTag("")
	if err == nil && count != 0 {
		t.Errorf("empty repository has 0 commits, got %d", count)
	}
}

func TestGetCommitCountSinceTag_RejectsNonNumericOutput(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()
	commandExecutor = func(name string, args ...string) (string, error) {
		return "12:34:56.789 trace: built-in: git rev-list\n3", nil
	}

	g := newTestGitWithConfig(&config.GitConfig{}, false)
	if _, err := g.GetCommitCountSinceTag("v1.0.0"); err == nil {
		t.Error("garbage output used to be digit-scraped into a huge count; it must be an error now")
	}
}
