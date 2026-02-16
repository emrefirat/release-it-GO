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

	t.Run("wildcard all files", func(t *testing.T) {
		files, err := ResolveAssets([]string{filepath.Join(tmpDir, "*")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 3 {
			t.Errorf("expected 3 files, got %d", len(files))
		}
	})

	t.Run("exact file path", func(t *testing.T) {
		files, err := ResolveAssets([]string{filepath.Join(tmpDir, "app.zip")})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
	})

	t.Run("duplicate patterns", func(t *testing.T) {
		files, err := ResolveAssets([]string{
			filepath.Join(tmpDir, "*.zip"),
			filepath.Join(tmpDir, "*.zip"),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Duplicate patterns produce duplicate results
		if len(files) != 2 {
			t.Errorf("expected 2 files (duplicated), got %d", len(files))
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
		{"gz file", "data.gz", "application/gzip"},
		{"bz2 file", "archive.bz2", "application/x-bzip2"},
		{"xz file", "archive.xz", "application/x-xz"},
		{"msi file", "installer.msi", "application/x-msi"},
		{"apk file", "app.apk", "application/vnd.android.package-archive"},
		{"path with directory", "/dist/linux/app.tar.gz", "application/gzip"},
		{"mixed case tar.gz", "APP.TAR.GZ", "application/gzip"},
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
