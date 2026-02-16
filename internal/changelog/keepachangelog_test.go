package changelog

import (
	"strings"
	"testing"
)

func TestRenderKeepAChangelog_Basic(t *testing.T) {
	commits := []*Commit{
		{Type: "feat", Description: "add OAuth2 support", Scope: "auth"},
		{Type: "fix", Description: "rate limiting issue", Scope: "api"},
		{Type: "refactor", Description: "updated dependency versions"},
	}

	result := RenderKeepAChangelog(commits, "1.2.0", "2026-02-16")

	if !strings.Contains(result, "## [1.2.0] - 2026-02-16") {
		t.Error("expected version header with date")
	}
	if !strings.Contains(result, "### Added") {
		t.Error("expected Added section")
	}
	if !strings.Contains(result, "### Fixed") {
		t.Error("expected Fixed section")
	}
	if !strings.Contains(result, "### Changed") {
		t.Error("expected Changed section")
	}
	if !strings.Contains(result, "add OAuth2 support (auth)") {
		t.Error("expected feat commit with scope in Added section")
	}
	if !strings.Contains(result, "rate limiting issue (api)") {
		t.Error("expected fix commit with scope in Fixed section")
	}
}

func TestRenderKeepAChangelog_NoScope(t *testing.T) {
	commits := []*Commit{
		{Type: "feat", Description: "add dark mode"},
	}

	result := RenderKeepAChangelog(commits, "1.0.0", "2026-01-01")

	if !strings.Contains(result, "- add dark mode\n") {
		t.Errorf("expected entry without scope, got: %s", result)
	}
}

func TestRenderKeepAChangelog_DefaultDate(t *testing.T) {
	commits := []*Commit{
		{Type: "feat", Description: "feature"},
	}

	result := RenderKeepAChangelog(commits, "1.0.0", "")

	// Should have a date in YYYY-MM-DD format
	if !strings.Contains(result, "## [1.0.0] - 20") {
		t.Error("expected auto-generated date")
	}
}

func TestRenderKeepAChangelog_PerfAndRevert(t *testing.T) {
	commits := []*Commit{
		{Type: "perf", Description: "optimize query"},
		{Type: "revert", Description: "revert feature X"},
	}

	result := RenderKeepAChangelog(commits, "1.1.0", "2026-02-16")

	if !strings.Contains(result, "### Changed") {
		t.Error("expected Changed section for perf")
	}
	if !strings.Contains(result, "### Removed") {
		t.Error("expected Removed section for revert")
	}
}

func TestRenderKeepAChangelog_NonChangelogTypes(t *testing.T) {
	commits := []*Commit{
		{Type: "docs", Description: "update readme"},
		{Type: "ci", Description: "update CI"},
		{Type: "test", Description: "add tests"},
	}

	result := RenderKeepAChangelog(commits, "1.0.0", "2026-01-01")

	if strings.Contains(result, "### Added") {
		t.Error("should not have Added section for non-changelog types")
	}
	if strings.Contains(result, "update readme") {
		t.Error("docs should not appear in keep-a-changelog")
	}
}
