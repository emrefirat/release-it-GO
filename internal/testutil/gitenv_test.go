package testutil

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestIsolateGit_GlobalConfigInvisible_DefaultBranchMain(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	// Restore the process environment for other tests in this package.
	for _, key := range []string{"GIT_CONFIG_GLOBAL", "GIT_CONFIG_NOSYSTEM", "GIT_TERMINAL_PROMPT", "GIT_CONFIG_COUNT", "GIT_CONFIG_KEY_0", "GIT_CONFIG_VALUE_0"} {
		t.Setenv(key, os.Getenv(key))
	}

	if err := IsolateGit(); err != nil {
		t.Fatalf("IsolateGit: %v", err)
	}

	// git must see no global settings at all, whatever the developer has.
	out, err := exec.Command("git", "config", "--global", "--list").CombinedOutput()
	if err == nil && strings.TrimSpace(string(out)) != "" {
		t.Errorf("global config leaked into the isolated environment:\n%s", out)
	}

	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	branch, err := exec.Command("git", "-C", dir, "symbolic-ref", "--short", "HEAD").Output()
	if err != nil {
		t.Fatalf("symbolic-ref: %v", err)
	}
	if got := strings.TrimSpace(string(branch)); got != "main" {
		t.Errorf("default branch = %q, want main regardless of the developer's init.defaultBranch", got)
	}
}
