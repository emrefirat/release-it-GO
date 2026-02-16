package changelog

import (
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
