package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAssets(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	for _, name := range []string{"app.zip", "app.tar.gz", "readme.md"} {
		os.WriteFile(filepath.Join(tmpDir, name), []byte("test"), 0644)
	}

	t.Run("glob match", func(t *testing.T) {
		files, err := ResolveAssets([]string{filepath.Join(tmpDir, "*.zip")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("multiple patterns", func(t *testing.T) {
		files, err := ResolveAssets([]string{
			filepath.Join(tmpDir, "*.zip"),
			filepath.Join(tmpDir, "*.tar.gz"),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("expected 2 files, got %d", len(files))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		files, err := ResolveAssets([]string{filepath.Join(tmpDir, "*.exe")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("empty patterns", func(t *testing.T) {
		files, err := ResolveAssets(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("invalid pattern", func(t *testing.T) {
		_, err := ResolveAssets([]string{"[invalid"})
		if err == nil {
			t.Error("expected error for invalid glob pattern")
		}
	})
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"zip file", "app.zip", "application/zip"},
		{"tar.gz file", "app.tar.gz", "application/gzip"},
		{"tgz file", "app.tgz", "application/gzip"},
		{"dmg file", "app.dmg", "application/x-apple-diskimage"},
		{"deb file", "app.deb", "application/vnd.debian.binary-package"},
		{"rpm file", "app.rpm", "application/x-rpm"},
		{"exe file", "app.exe", "application/vnd.microsoft.portable-executable"},
		{"sha256 file", "app.sha256", "text/plain"},
		{"sig file", "app.sig", "application/pgp-signature"},
		{"asc file", "app.asc", "application/pgp-signature"},
		{"unknown file", "app.unknown123", "application/octet-stream"},
		{"uppercase", "APP.ZIP", "application/zip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectContentType(tt.filePath)
			if got != tt.want {
				t.Errorf("DetectContentType(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}
