package changelog

import (
	"strings"
	"testing"

	"github.com/emfi/release-it-go/internal/git"
)

func TestRenderConventional_SingleFeat(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890", Type: "feat", Description: "add login", Scope: "auth"},
	}

	result := RenderConventional(commits, "v1.2.0", "v1.1.0", nil)

	if !strings.Contains(result, "## v1.2.0") {
		t.Error("expected version header")
	}
	if !strings.Contains(result, "### Features") {
		t.Error("expected Features section")
	}
	if !strings.Contains(result, "**auth:**") {
		t.Error("expected scope in entry")
	}
	if !strings.Contains(result, "add login") {
		t.Error("expected description in entry")
	}
}

func TestRenderConventional_WithRepoInfo(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890", Type: "feat", Description: "add feature"},
	}

	repoInfo := &git.RepoInfo{
		Host:       "github.com",
		Owner:      "emfi",
		Repository: "release-it-go",
	}

	result := RenderConventional(commits, "v1.2.0", "v1.1.0", repoInfo)

	if !strings.Contains(result, "https://github.com/emfi/release-it-go/compare/v1.1.0...v1.2.0") {
		t.Error("expected compare URL in header")
	}
	if !strings.Contains(result, "https://github.com/emfi/release-it-go/commit/abc1234567890") {
		t.Error("expected commit URL")
	}
	if !strings.Contains(result, "abc1234") {
		t.Error("expected short hash")
	}
}

func TestRenderConventional_MixedTypes(t *testing.T) {
	commits := []*Commit{
		{Hash: "aaa1111111111", Type: "feat", Description: "add feature"},
		{Hash: "bbb2222222222", Type: "fix", Description: "fix bug"},
		{Hash: "ccc3333333333", Type: "perf", Description: "optimize query"},
		{Hash: "ddd4444444444", Type: "docs", Description: "update readme"},
	}

	result := RenderConventional(commits, "v1.2.0", "", nil)

	if !strings.Contains(result, "### Features") {
		t.Error("expected Features section")
	}
	if !strings.Contains(result, "### Bug Fixes") {
		t.Error("expected Bug Fixes section")
	}
	if !strings.Contains(result, "### Performance Improvements") {
		t.Error("expected Performance Improvements section")
	}
	// docs should NOT appear
	if strings.Contains(result, "update readme") {
		t.Error("docs commits should not appear in conventional changelog")
	}
}

func TestRenderConventional_BreakingChanges(t *testing.T) {
	commits := []*Commit{
		{
			Hash:            "aaa1111111111",
			Type:            "feat",
			Scope:           "auth",
			Description:     "new API",
			BreakingChange:  true,
			BreakingMessage: "OAuth1 support removed",
		},
	}

	result := RenderConventional(commits, "v2.0.0", "v1.0.0", nil)

	if !strings.Contains(result, "### BREAKING CHANGES") {
		t.Error("expected BREAKING CHANGES section")
	}
	if !strings.Contains(result, "OAuth1 support removed") {
		t.Error("expected breaking change message")
	}
	if !strings.Contains(result, "**auth:**") {
		t.Error("expected scope in breaking change")
	}
}

func TestRenderConventional_NoPrevVersion(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890", Type: "feat", Description: "initial feature"},
	}

	repoInfo := &git.RepoInfo{
		Host:       "github.com",
		Owner:      "emfi",
		Repository: "release-it-go",
	}

	result := RenderConventional(commits, "v1.0.0", "", repoInfo)

	// No compare URL when no prev version
	if strings.Contains(result, "compare") {
		t.Error("should not have compare URL without prev version")
	}
	if !strings.Contains(result, "## v1.0.0") {
		t.Error("expected version header")
	}
}

func TestRenderConventional_NoCommitsForSections(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890", Type: "chore", Description: "update deps"},
	}

	result := RenderConventional(commits, "v1.0.1", "", nil)

	if strings.Contains(result, "### Features") {
		t.Error("should not have Features section")
	}
	if strings.Contains(result, "### Bug Fixes") {
		t.Error("should not have Bug Fixes section")
	}
}

func TestShortHash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"abc1234567890", "abc1234"},
		{"abc", "abc"},
		{"", ""},
		{"1234567", "1234567"},
		{"12345678", "1234567"},
	}

	for _, tt := range tests {
		got := shortHash(tt.input)
		if got != tt.want {
			t.Errorf("shortHash(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
