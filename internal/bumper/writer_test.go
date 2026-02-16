package bumper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emfi/release-it-go/internal/config"
)

func TestWriteVersionToFile_JSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	os.WriteFile(file, []byte(`{"name": "test", "version": "1.0.0"}`), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "version"}, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if !strings.Contains(string(data), `"version": "2.0.0"`) {
		t.Errorf("expected version 2.0.0 in JSON, got %s", string(data))
	}
	// Name should be preserved
	if !strings.Contains(string(data), `"name": "test"`) {
		t.Errorf("expected name to be preserved, got %s", string(data))
	}
}

func TestWriteVersionToFile_JSON_Nested(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.json")
	os.WriteFile(file, []byte(`{"tool": {"version": "1.0.0"}}`), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "tool.version"}, "3.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if !strings.Contains(string(data), `"version": "3.0.0"`) {
		t.Errorf("expected version 3.0.0, got %s", string(data))
	}
}

func TestWriteVersionToFile_YAML(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "chart.yaml")
	os.WriteFile(file, []byte("name: myapp\nversion: 1.0.0\n"), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "version"}, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	content := string(data)
	if !strings.Contains(content, "version: 2.0.0") {
		t.Errorf("expected version 2.0.0, got %s", content)
	}
	if !strings.Contains(content, "name: myapp") {
		t.Errorf("expected name to be preserved, got %s", content)
	}
}

func TestWriteVersionToFile_TOML(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "pyproject.toml")
	content := "[tool]\n[tool.poetry]\nversion = \"1.0.0\"\nname = \"myapp\"\n"
	os.WriteFile(file, []byte(content), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "tool.poetry.version"}, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	result := string(data)
	if !strings.Contains(result, "version = '2.0.0'") && !strings.Contains(result, `version = "2.0.0"`) {
		t.Errorf("expected version 2.0.0, got %s", result)
	}
}

func TestWriteVersionToFile_INI(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "setup.cfg")
	content := "[metadata]\nname = mypackage\nversion = 1.0.0\n"
	os.WriteFile(file, []byte(content), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "[metadata].version"}, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	result := string(data)
	if !strings.Contains(result, "version = 2.0.0") {
		t.Errorf("expected version = 2.0.0, got %s", result)
	}
	if !strings.Contains(result, "name = mypackage") {
		t.Errorf("expected name to be preserved, got %s", result)
	}
}

func TestWriteVersionToFile_Text(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "VERSION")
	os.WriteFile(file, []byte("1.0.0\n"), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file}, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "2.0.0\n" {
		t.Errorf("expected '2.0.0\\n', got %q", string(data))
	}
}

func TestWriteVersionToFile_ConsumeWholeFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "VERSION")
	os.WriteFile(file, []byte("old content\n"), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, ConsumeWholeFile: true}, "3.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "3.0.0\n" {
		t.Errorf("expected '3.0.0\\n', got %q", string(data))
	}
}

func TestWriteVersionToFile_Prefix(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "VERSION")

	err := WriteVersionToFile(config.BumperFile{File: file, Prefix: "^", ConsumeWholeFile: true}, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "^1.0.0\n" {
		t.Errorf("expected '^1.0.0\\n', got %q", string(data))
	}
}

func TestWriteVersionToFile_INI_MissingKey(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "setup.cfg")
	content := "[metadata]\nname = mypackage\n"
	os.WriteFile(file, []byte(content), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "[metadata].version"}, "2.0.0")
	if err == nil {
		t.Error("expected error for missing INI key")
	}
}

func TestWriteVersionToFile_JSON_MissingPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	os.WriteFile(file, []byte(`{"name": "test"}`), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "missing.deep.path"}, "2.0.0")
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestSetNestedValue_EmptyPath(t *testing.T) {
	obj := map[string]interface{}{"key": "value"}
	err := setNestedValue(obj, "", "new")
	if err == nil {
		t.Error("expected error for empty path")
	}
}
