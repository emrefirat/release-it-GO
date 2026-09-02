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

func TestLintCommits_AutoPassesFixupSquashAmend(t *testing.T) {
	// commitlint's defaults ignore these: they exist to be squashed before
	// merge, and the commit-msg hook must not reject `git commit --fixup`.
	commits := []LintInput{
		{Hash: "a", Subject: "fixup! feat: x"},
		{Hash: "b", Subject: "squash! fix: y"},
		{Hash: "c", Subject: "amend! chore: z"},
	}
	passed, failed := LintCommits(commits)
	if len(failed) != 0 {
		t.Fatalf("fixup!/squash!/amend! must auto-pass, failed: %+v", failed)
	}
	if len(passed) != 3 {
		t.Errorf("passed = %d, want 3", len(passed))
	}
}

func TestSuggestType(t *testing.T) {
	tests := []struct{ in, want string }{
		{"fic", "fix"},     // one-letter typo
		{"Feat", "feat"},   // wrong case
		{"FIX", "fix"},     // wrong case
		{"chroe", "chore"}, // transposition
		{"docs", ""},       // already valid — nothing to suggest
		{"xyzzy", ""},      // nothing close
		{"", ""},
	}
	for _, tt := range tests {
		if got := SuggestType(tt.in); got != tt.want {
			t.Errorf("SuggestType(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLintCommits_UnknownType_CarriesSuggestion(t *testing.T) {
	_, failed := LintCommits([]LintInput{{Hash: "a", Subject: "fic: deneme"}})
	if len(failed) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failed))
	}
	if failed[0].Suggestion != "fix" {
		t.Errorf("Suggestion = %q, want fix", failed[0].Suggestion)
	}
}

func TestValidTypes_MatchesAllowedTypes(t *testing.T) {
	types := ValidTypes()
	if len(types) != len(allowedTypes) {
		t.Fatalf("ValidTypes has %d entries, allowedTypes has %d — they must be one list", len(types), len(allowedTypes))
	}
	seen := map[string]bool{}
	for _, tt := range types {
		if !allowedTypes[tt.Name] {
			t.Errorf("ValidTypes lists %q but the linter does not accept it", tt.Name)
		}
		if tt.Description == "" {
			t.Errorf("%q has no description", tt.Name)
		}
		seen[tt.Name] = true
	}
	if !seen["build"] {
		t.Error("build is accepted by the linter but was missing from the help list")
	}
}
