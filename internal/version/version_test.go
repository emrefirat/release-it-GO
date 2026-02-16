package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetVersionFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{"simple version", "1.2.3\n", "1.2.3", false},
		{"with v prefix", "v1.2.3\n", "1.2.3", false},
		{"with whitespace", "  1.2.3  \n", "1.2.3", false},
		{"empty file", "", "", true},
		{"whitespace only", "   \n", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			filePath := filepath.Join(dir, "VERSION")
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := GetVersionFromFile(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersionFromFile error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVersionFromFile = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetVersionFromFile_NonexistentFile(t *testing.T) {
	_, err := GetVersionFromFile("/nonexistent/VERSION")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestDetectVersion_FallsBackToVersionFile(t *testing.T) {
	// Mock runGit to fail (no git tags)
	origRunGit := runGit
	defer func() { runGit = origRunGit }()
	runGit = func(args ...string) (string, error) {
		return "", os.ErrNotExist
	}

	// Create a VERSION file in a temp dir and change to it
	dir := t.TempDir()
	versionFile := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(versionFile, []byte("2.5.0\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	got := DetectVersion(VersionOptions{})
	if got != "2.5.0" {
		t.Errorf("DetectVersion = %q, want %q", got, "2.5.0")
	}
}

func TestDetectVersion_FallsBackToDefault(t *testing.T) {
	origRunGit := runGit
	defer func() { runGit = origRunGit }()
	runGit = func(args ...string) (string, error) {
		return "", os.ErrNotExist
	}

	// Change to temp dir with no VERSION file
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	got := DetectVersion(VersionOptions{})
	if got != FallbackVersion {
		t.Errorf("DetectVersion = %q, want %q", got, FallbackVersion)
	}
}

func TestGetLatestTagVersion_WithMockedGit(t *testing.T) {
	origRunGit := runGit
	defer func() { runGit = origRunGit }()

	runGit = func(args ...string) (string, error) {
		return "v1.5.0\n", nil
	}

	got, err := GetLatestTagVersion(VersionOptions{})
	if err != nil {
		t.Fatalf("GetLatestTagVersion error: %v", err)
	}
	if got != "1.5.0" {
		t.Errorf("GetLatestTagVersion = %q, want %q", got, "1.5.0")
	}
}

func TestGetLatestTagVersion_WithTagMatch(t *testing.T) {
	origRunGit := runGit
	defer func() { runGit = origRunGit }()

	runGit = func(args ...string) (string, error) {
		return "v1.5.0\n", nil
	}

	_, err := GetLatestTagVersion(VersionOptions{TagMatch: "v2*"})
	if err == nil {
		t.Error("expected error when tag doesn't match pattern")
	}
}

func TestGetLatestTagVersion_WithTagExclude(t *testing.T) {
	origRunGit := runGit
	defer func() { runGit = origRunGit }()

	runGit = func(args ...string) (string, error) {
		return "v1.5.0-beta.1\n", nil
	}

	_, err := GetLatestTagVersion(VersionOptions{TagExclude: "*-*"})
	if err == nil {
		t.Error("expected error when tag matches exclude pattern")
	}
}

func TestGetLatestTagVersion_AllRefs(t *testing.T) {
	origRunGit := runGit
	defer func() { runGit = origRunGit }()

	runGit = func(args ...string) (string, error) {
		if args[0] == "tag" {
			return "v2.0.0\nv1.5.0\nv1.0.0\n", nil
		}
		return "", os.ErrNotExist
	}

	got, err := GetLatestTagVersion(VersionOptions{GetFromAllRefs: true})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if got != "2.0.0" {
		t.Errorf("got %q, want %q", got, "2.0.0")
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		s       string
		pattern string
		want    bool
	}{
		{"anything", "*", true},
		{"v1.0.0", "v*", true},
		{"v1.0.0", "v2*", false},
		{"v1.0.0-beta", "*-*", true},
		{"v1.0.0", "*-*", false},
		{"v1.0.0", "v1.0.0", true},
		{"v1.0.0", "v1.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.pattern, func(t *testing.T) {
			got := matchPattern(tt.s, tt.pattern)
			if got != tt.want {
				t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.s, tt.pattern, got, tt.want)
			}
		})
	}
}
