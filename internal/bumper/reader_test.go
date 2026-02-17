package bumper

import (
	"os"
	"path/filepath"
	"testing"

	"release-it-go/internal/config"
)

func TestReadVersionFromFile_JSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	_ = os.WriteFile(file, []byte(`{"version": "1.2.3"}`), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "1.2.3" {
		t.Errorf("expected 1.2.3, got %q", v)
	}
}

func TestReadVersionFromFile_JSON_Nested(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.json")
	_ = os.WriteFile(file, []byte(`{"tool": {"poetry": {"version": "2.0.0"}}}`), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "tool.poetry.version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "2.0.0" {
		t.Errorf("expected 2.0.0, got %q", v)
	}
}

func TestReadVersionFromFile_YAML(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "chart.yaml")
	_ = os.WriteFile(file, []byte("version: 3.1.0\nname: myapp\n"), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "3.1.0" {
		t.Errorf("expected 3.1.0, got %q", v)
	}
}

func TestReadVersionFromFile_TOML(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "pyproject.toml")
	content := "[tool]\n[tool.poetry]\nversion = \"4.5.6\"\n"
	_ = os.WriteFile(file, []byte(content), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "tool.poetry.version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "4.5.6" {
		t.Errorf("expected 4.5.6, got %q", v)
	}
}

func TestReadVersionFromFile_INI(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "setup.cfg")
	content := "[metadata]\nname = mypackage\nversion = 1.0.0\n"
	_ = os.WriteFile(file, []byte(content), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "[metadata].version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "1.0.0" {
		t.Errorf("expected 1.0.0, got %q", v)
	}
}

func TestReadVersionFromFile_INI_NoSection(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.ini")
	content := "version = 2.0.0\nname = test\n"
	_ = os.WriteFile(file, []byte(content), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "2.0.0" {
		t.Errorf("expected 2.0.0, got %q", v)
	}
}

func TestReadVersionFromFile_Text(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "VERSION")
	_ = os.WriteFile(file, []byte("5.0.0\n"), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "5.0.0" {
		t.Errorf("expected 5.0.0, got %q", v)
	}
}

func TestReadVersionFromFile_ExplicitType(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "custom.dat")
	_ = os.WriteFile(file, []byte(`{"version": "9.9.9"}`), 0644)

	v, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "version", Type: "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "9.9.9" {
		t.Errorf("expected 9.9.9, got %q", v)
	}
}

func TestReadVersionFromFile_FileNotFound(t *testing.T) {
	_, err := ReadVersionFromFile(config.BumperFile{File: "/nonexistent/file.json", Path: "version"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadVersionFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(file, []byte(`{invalid json}`), 0644)

	_, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "version"})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReadVersionFromFile_MissingKey(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	_ = os.WriteFile(file, []byte(`{"name": "test"}`), 0644)

	_, err := ReadVersionFromFile(config.BumperFile{File: file, Path: "version"})
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestReadINI_MissingKey(t *testing.T) {
	_, err := readINI([]byte("[metadata]\nname = test\n"), "[metadata].version")
	if err == nil {
		t.Error("expected error for missing INI key")
	}
}

func TestParseINIPath(t *testing.T) {
	tests := []struct {
		path    string
		section string
		key     string
	}{
		{"[metadata].version", "metadata", "version"},
		{"version", "", "version"},
		{"[section].key", "section", "key"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			section, key := parseINIPath(tt.path)
			if section != tt.section || key != tt.key {
				t.Errorf("parseINIPath(%q) = (%q, %q), want (%q, %q)",
					tt.path, section, key, tt.section, tt.key)
			}
		})
	}
}

func TestExtractNestedValue_NonMap(t *testing.T) {
	_, err := extractNestedValue("string value", "key")
	if err == nil {
		t.Error("expected error for non-map value")
	}
}
