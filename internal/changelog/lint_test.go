package changelog

import (
	"fmt"
	"strings"
	"testing"
)

func TestLintCommits_AllConventional(t *testing.T) {
	commits := []LintInput{
		{Hash: "abc1234", Subject: "feat: add new feature"},
		{Hash: "def5678", Subject: "fix: resolve bug"},
		{Hash: "ghi9012", Subject: "docs: update readme"},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 3 {
		t.Errorf("expected 3 passed, got %d", len(passed))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
	for _, p := range passed {
		if p.Reason != "conventional commit" {
			t.Errorf("expected reason 'conventional commit', got %q", p.Reason)
		}
	}
}

func TestLintCommits_NonConventional(t *testing.T) {
	commits := []LintInput{
		{Hash: "abc1234", Subject: "fix some bug"},
		{Hash: "def5678", Subject: "update readme"},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 0 {
		t.Errorf("expected 0 passed, got %d", len(passed))
	}
	if len(failed) != 2 {
		t.Errorf("expected 2 failed, got %d", len(failed))
	}
	for _, f := range failed {
		if f.Reason != "not in conventional commit format" {
			t.Errorf("expected reason 'not in conventional commit format', got %q", f.Reason)
		}
		if f.Valid {
			t.Error("expected Valid=false for failed commit")
		}
	}
}

func TestLintCommits_MergeCommit(t *testing.T) {
	commits := []LintInput{
		{Hash: "abc1234", Subject: "Merge branch 'feature' into main"},
		{Hash: "def5678", Subject: "Merge pull request #42 from owner/branch"},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 2 {
		t.Errorf("expected 2 passed, got %d", len(passed))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
	for _, p := range passed {
		if p.Reason != "merge commit" {
			t.Errorf("expected reason 'merge commit', got %q", p.Reason)
		}
	}
}

func TestLintCommits_RevertCommit(t *testing.T) {
	commits := []LintInput{
		{Hash: "abc1234", Subject: "Revert \"feat: add feature\""},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 1 {
		t.Errorf("expected 1 passed, got %d", len(passed))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
	if passed[0].Reason != "revert commit" {
		t.Errorf("expected reason 'revert commit', got %q", passed[0].Reason)
	}
}

func TestLintCommits_Mixed(t *testing.T) {
	commits := []LintInput{
		{Hash: "aaa1111", Subject: "feat: add feature"},
		{Hash: "bbb2222", Subject: "bad commit message"},
		{Hash: "ccc3333", Subject: "Merge branch 'main'"},
		{Hash: "ddd4444", Subject: "fix(core): resolve issue"},
		{Hash: "eee5555", Subject: "another bad one"},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 3 {
		t.Errorf("expected 3 passed, got %d", len(passed))
	}
	if len(failed) != 2 {
		t.Errorf("expected 2 failed, got %d", len(failed))
	}
}

func TestLintCommits_Empty(t *testing.T) {
	passed, failed := LintCommits(nil)
	if len(passed) != 0 {
		t.Errorf("expected 0 passed, got %d", len(passed))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
}

func TestLintCommits_ScopedAndBreaking(t *testing.T) {
	commits := []LintInput{
		{Hash: "aaa1111", Subject: "feat(auth): add login"},
		{Hash: "bbb2222", Subject: "fix(ui)!: breaking change"},
		{Hash: "ccc3333", Subject: "chore: update deps"},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 3 {
		t.Errorf("expected 3 passed, got %d", len(passed))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
}

func TestLintCommits_WhitespaceHandling(t *testing.T) {
	commits := []LintInput{
		{Hash: "aaa1111", Subject: "  feat: add feature  "},
		{Hash: "bbb2222", Subject: "  bad commit  "},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 1 {
		t.Errorf("expected 1 passed, got %d", len(passed))
	}
	if len(failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(failed))
	}
}

func TestLintCommits_InvalidType(t *testing.T) {
	commits := []LintInput{
		{Hash: "aaa1111", Subject: "fic: deneme commit"},
		{Hash: "bbb2222", Subject: "foo: bar baz"},
		{Hash: "ccc3333", Subject: "feature: new thing"},
	}

	passed, failed := LintCommits(commits)
	if len(passed) != 0 {
		t.Errorf("expected 0 passed, got %d", len(passed))
	}
	if len(failed) != 3 {
		t.Errorf("expected 3 failed, got %d", len(failed))
	}
	for _, f := range failed {
		if !strings.HasPrefix(f.Reason, "unknown type:") {
			t.Errorf("expected 'unknown type:' reason, got %q", f.Reason)
		}
	}
}

func TestLintCommits_AllValidTypes(t *testing.T) {
	types := []string{"feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"}
	commits := make([]LintInput, len(types))
	for i, typ := range types {
		commits[i] = LintInput{Hash: fmt.Sprintf("hash%d", i), Subject: typ + ": something"}
	}

	passed, failed := LintCommits(commits)
	if len(passed) != len(types) {
		t.Errorf("expected %d passed, got %d", len(types), len(passed))
	}
	if len(failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(failed))
	}
}
