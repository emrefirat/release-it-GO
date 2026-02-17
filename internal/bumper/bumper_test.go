package bumper

import (
	"os"
	"path/filepath"
	"testing"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
)

func TestNewBumper(t *testing.T) {
	cfg := &config.BumperConfig{Enabled: true}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)
	if b == nil {
		t.Fatal("expected non-nil Bumper")
	}
}

func TestBumper_ReadVersion_NilConfig(t *testing.T) {
	logger := applog.NewLogger(0, false)
	b := NewBumper(nil, logger, false)
	v, err := b.ReadVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "" {
		t.Errorf("expected empty version, got %q", v)
	}
}

func TestBumper_ReadVersion_NoIn(t *testing.T) {
	cfg := &config.BumperConfig{Enabled: true}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)
	v, err := b.ReadVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "" {
		t.Errorf("expected empty version, got %q", v)
	}
}

func TestBumper_ReadVersion_JSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	_ = os.WriteFile(file, []byte(`{"version": "1.2.3"}`), 0644)

	cfg := &config.BumperConfig{
		Enabled: true,
		In:      &config.BumperFile{File: file, Path: "version"},
	}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)

	v, err := b.ReadVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "1.2.3" {
		t.Errorf("expected 1.2.3, got %q", v)
	}
}

func TestBumper_WriteVersion_DryRun(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "version.txt")
	_ = os.WriteFile(file, []byte("1.0.0\n"), 0644)

	cfg := &config.BumperConfig{
		Enabled: true,
		Out: []config.BumperFile{
			{File: file, ConsumeWholeFile: true},
		},
	}
	logger := applog.NewLogger(0, true)
	b := NewBumper(cfg, logger, true)

	err := b.WriteVersion("2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should NOT be updated in dry-run
	data, _ := os.ReadFile(file)
	if string(data) != "1.0.0\n" {
		t.Errorf("file should not be modified in dry-run, got %q", string(data))
	}
}

func TestBumper_WriteVersion_Text(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "VERSION")
	_ = os.WriteFile(file, []byte("1.0.0\n"), 0644)

	cfg := &config.BumperConfig{
		Enabled: true,
		Out: []config.BumperFile{
			{File: file, ConsumeWholeFile: true},
		},
	}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)

	err := b.WriteVersion("2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "2.0.0\n" {
		t.Errorf("expected '2.0.0\\n', got %q", string(data))
	}
}

func TestBumper_WriteVersion_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	file1 := filepath.Join(dir, "VERSION")
	file2 := filepath.Join(dir, "version.txt")
	_ = os.WriteFile(file1, []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(file2, []byte("1.0.0\n"), 0644)

	cfg := &config.BumperConfig{
		Enabled: true,
		Out: []config.BumperFile{
			{File: file1, ConsumeWholeFile: true},
			{File: file2, ConsumeWholeFile: true},
		},
	}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)

	err := b.WriteVersion("3.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data1, _ := os.ReadFile(file1)
	data2, _ := os.ReadFile(file2)
	if string(data1) != "3.0.0\n" {
		t.Errorf("file1: expected '3.0.0\\n', got %q", string(data1))
	}
	if string(data2) != "3.0.0\n" {
		t.Errorf("file2: expected '3.0.0\\n', got %q", string(data2))
	}
}

func TestBumper_WriteVersion_Glob(t *testing.T) {
	dir := t.TempDir()
	file1 := filepath.Join(dir, "a.txt")
	file2 := filepath.Join(dir, "b.txt")
	_ = os.WriteFile(file1, []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(file2, []byte("1.0.0\n"), 0644)

	cfg := &config.BumperConfig{
		Enabled: true,
		Out: []config.BumperFile{
			{File: filepath.Join(dir, "*.txt"), ConsumeWholeFile: true},
		},
	}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)

	err := b.WriteVersion("2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data1, _ := os.ReadFile(file1)
	data2, _ := os.ReadFile(file2)
	if string(data1) != "2.0.0\n" {
		t.Errorf("file1: expected '2.0.0\\n', got %q", string(data1))
	}
	if string(data2) != "2.0.0\n" {
		t.Errorf("file2: expected '2.0.0\\n', got %q", string(data2))
	}
}

func TestBumper_WriteVersion_Prefix(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "VERSION")
	_ = os.WriteFile(file, []byte("^1.0.0\n"), 0644)

	cfg := &config.BumperConfig{
		Enabled: true,
		Out: []config.BumperFile{
			{File: file, Prefix: "^", ConsumeWholeFile: true},
		},
	}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)

	err := b.WriteVersion("2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "^2.0.0\n" {
		t.Errorf("expected '^2.0.0\\n', got %q", string(data))
	}
}

func TestBumper_WriteVersion_NoOut(t *testing.T) {
	cfg := &config.BumperConfig{Enabled: true}
	logger := applog.NewLogger(0, false)
	b := NewBumper(cfg, logger, false)

	err := b.WriteVersion("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContainsGlobChar(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"file.txt", false},
		{"*.txt", true},
		{"path/?.json", true},
		{"[abc].yaml", true},
		{"normal/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := containsGlobChar(tt.input); got != tt.expected {
				t.Errorf("containsGlobChar(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestResolveGlob_NoGlob(t *testing.T) {
	files, err := resolveGlob("/some/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 || files[0] != "/some/file.txt" {
		t.Errorf("expected [/some/file.txt], got %v", files)
	}
}

func TestResolveGlob_NoMatch(t *testing.T) {
	_, err := resolveGlob("/nonexistent/path/*.xyz123")
	if err == nil {
		t.Error("expected error for no matching files")
	}
}
