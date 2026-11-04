package changelog

import (
	"github.com/emrefirat/release-it-GO/internal/git"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateChangelog_ConventionalFormat(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890", Type: "feat", Description: "add feature"},
	}

	opts := Options{KeepAChangelog: false}
	result := GenerateChangelog(commits, "v1.0.0", "", opts)

	if !strings.Contains(result, "### Features") {
		t.Error("expected conventional format with Features section")
	}
}

func TestGenerateChangelog_KeepAChangelogFormat(t *testing.T) {
	commits := []*Commit{
		{Hash: "abc1234567890", Type: "feat", Description: "add feature"},
	}

	opts := Options{KeepAChangelog: true}
	result := GenerateChangelog(commits, "1.0.0", "", opts)

	if !strings.Contains(result, "### Added") {
		t.Error("expected keep-a-changelog format with Added section")
	}
}

func TestUpdateChangelogFile_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "CHANGELOG.md")

	newContent := "## v1.0.0\n\n### Features\n\n* initial release\n"
	err := UpdateChangelogFile(filePath, newContent, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Changelog") {
		t.Error("expected default header")
	}
	if !strings.Contains(content, "## v1.0.0") {
		t.Error("expected version content")
	}
}

func TestUpdateChangelogFile_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "CHANGELOG.md")

	existing := "# Changelog\n\n## v1.0.0\n\n### Features\n\n* initial release\n"
	err := os.WriteFile(filePath, []byte(existing), 0644)
	if err != nil {
		t.Fatalf("failed to write existing file: %v", err)
	}

	newContent := "## v1.1.0\n\n### Bug Fixes\n\n* fix bug\n"
	err = UpdateChangelogFile(filePath, newContent, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	content := string(data)

	// New version should come before old version
	v110Idx := strings.Index(content, "## v1.1.0")
	v100Idx := strings.Index(content, "## v1.0.0")

	if v110Idx == -1 {
		t.Fatal("expected v1.1.0 in changelog")
	}
	if v100Idx == -1 {
		t.Fatal("expected v1.0.0 in changelog")
	}
	if v110Idx > v100Idx {
		t.Error("v1.1.0 should come before v1.0.0 (prepend)")
	}

	// Header should be preserved
	if !strings.HasPrefix(content, "# Changelog") {
		t.Error("expected header to be preserved at top")
	}
}

func TestUpdateChangelogFile_CustomHeader(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "CHANGELOG.md")

	newContent := "## v1.0.0\n\n* initial\n"
	err := UpdateChangelogFile(filePath, newContent, "# My Custom Changelog")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	if !strings.Contains(string(data), "# My Custom Changelog") {
		t.Error("expected custom header")
	}
}

func TestInsertAfterHeader(t *testing.T) {
	tests := []struct {
		name       string
		existing   string
		header     string
		newSection string
		wantOrder  []string
	}{
		{
			name:       "insert after header",
			existing:   "# Changelog\n\n## v1.0.0\n\nold content",
			header:     "# Changelog",
			newSection: "## v1.1.0\n\nnew content",
			wantOrder:  []string{"# Changelog", "## v1.1.0", "## v1.0.0"},
		},
		{
			name:       "no header found",
			existing:   "## v1.0.0\n\nold content",
			header:     "# Changelog",
			newSection: "## v1.1.0\n\nnew content",
			wantOrder:  []string{"# Changelog", "## v1.1.0", "## v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := insertAfterHeader(tt.existing, tt.header, tt.newSection)

			prevIdx := -1
			for _, want := range tt.wantOrder {
				idx := strings.Index(result, want)
				if idx == -1 {
					t.Errorf("expected %q in result", want)
					continue
				}
				if idx <= prevIdx {
					t.Errorf("%q should come after previous item", want)
				}
				prevIdx = idx
			}
		})
	}
}

func TestUpdateChangelogFile_PreservesExistingContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "CHANGELOG.md")

	existing := "# Changelog\n\n## v1.0.0\n\n### Features\n\n* feature one\n* feature two\n"
	err := os.WriteFile(filePath, []byte(existing), 0644)
	if err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	newContent := "## v1.1.0\n\n### Bug Fixes\n\n* fix one\n"
	err = UpdateChangelogFile(filePath, newContent, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "feature one") {
		t.Error("expected existing content to be preserved")
	}
	if !strings.Contains(content, "feature two") {
		t.Error("expected existing content to be preserved")
	}
	if !strings.Contains(content, "fix one") {
		t.Error("expected new content to be added")
	}
}

func TestGenerateChangelog_AddVersionURL_GatesCompareLink(t *testing.T) {
	commits := []*Commit{{Type: "feat", Description: "thing", Hash: "abc1234"}}
	repo := &git.RepoInfo{Host: "github.com", Owner: "o", Repository: "r"}

	with := GenerateChangelog(commits, "1.1.0", "1.0.0", Options{RepoInfo: repo, AddVersionURL: true})
	if !strings.Contains(with, "compare/1.0.0...1.1.0") {
		t.Errorf("addVersionUrl: true must emit the compare link, got:\n%s", with)
	}

	without := GenerateChangelog(commits, "1.1.0", "1.0.0", Options{RepoInfo: repo, AddVersionURL: false})
	if strings.Contains(without, "compare/") {
		t.Errorf("addVersionUrl: false was documented but always ignored; got:\n%s", without)
	}
	if !strings.Contains(without, "abc1234") {
		t.Error("commit hash links must not depend on addVersionUrl")
	}
}

func TestInsertAfterHeader_KeepAChangelog_UnreleasedStaysFirst(t *testing.T) {
	existing := "# Changelog\n\nAll notable changes.\n\n## [Unreleased]\n\n- wip item\n\n## [1.0.0] - 2026-01-01\n\n### Added\n- first\n"
	result := insertAfterHeader(existing, "# Changelog", "## [1.1.0] - 2026-02-01\n\n### Added\n- second\n")

	order := []string{"# Changelog", "## [Unreleased]", "- wip item", "## [1.1.0]", "## [1.0.0]"}
	prev := -1
	for _, want := range order {
		idx := strings.Index(result, want)
		if idx == -1 {
			t.Fatalf("expected %q in result:\n%s", want, result)
		}
		if idx <= prev {
			t.Errorf("%q must come after the previous heading:\n%s", want, result)
		}
		prev = idx
	}
	if strings.Count(result, "## [Unreleased]") != 1 {
		t.Errorf("Unreleased heading duplicated:\n%s", result)
	}
}

func TestInsertAfterHeader_CRLFFileKeepsLineEndings(t *testing.T) {
	existing := "# Changelog\r\n\r\n## [1.0.0]\r\n\r\n- old\r\n"
	result := insertAfterHeader(existing, "# Changelog", "## [1.1.0]\n\n- new\n")

	if strings.Contains(strings.ReplaceAll(result, "\r\n", ""), "\n") {
		t.Errorf("mixed line endings in result: %q", result)
	}
	if !strings.Contains(result, "# Changelog\r\n\r\n## [1.1.0]\r\n\r\n- new\r\n\r\n## [1.0.0]") {
		t.Errorf("unexpected layout: %q", result)
	}
}

func TestInsertAfterHeader_HeaderMatchesWholeLineOnly(t *testing.T) {
	// "## Changelog" contains "# Changelog" but is not the header.
	existing := "## Changelog of the old era\n\n## v0.9.0\n"
	result := insertAfterHeader(existing, "# Changelog", "## v1.0.0\n")

	if !strings.HasPrefix(result, "# Changelog\n\n## v1.0.0\n") {
		t.Errorf("header must be prepended when only a substring matches:\n%s", result)
	}
}
