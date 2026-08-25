package bumper

import (
	"encoding/json"
	yaml "go.yaml.in/yaml/v3"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"release-it-go/internal/config"
)

func TestWriteVersionToFile_JSON(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	_ = os.WriteFile(file, []byte(`{"name": "test", "version": "1.0.0"}`), 0644)

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
	_ = os.WriteFile(file, []byte(`{"tool": {"version": "1.0.0"}}`), 0644)

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
	_ = os.WriteFile(file, []byte("name: myapp\nversion: 1.0.0\n"), 0644)

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
	_ = os.WriteFile(file, []byte(content), 0644)

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
	_ = os.WriteFile(file, []byte(content), 0644)

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
	_ = os.WriteFile(file, []byte("1.0.0\n"), 0644)

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
	_ = os.WriteFile(file, []byte("old content\n"), 0644)

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
	_ = os.WriteFile(file, []byte(content), 0644)

	err := WriteVersionToFile(config.BumperFile{File: file, Path: "[metadata].version"}, "2.0.0")
	if err == nil {
		t.Error("expected error for missing INI key")
	}
}

func TestWriteVersionToFile_JSON_MissingPath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "package.json")
	_ = os.WriteFile(file, []byte(`{"name": "test"}`), 0644)

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

func TestWriteJSON_PreservesFormattingAndKeyOrder(t *testing.T) {
	input := "{\n    \"name\": \"my-app\",\n    \"version\": \"1.0.0\",\n    \"description\": \"a & b\",\n    \"scripts\": {\n        \"build\": \"tsc\"\n    }\n}\n"
	want := strings.Replace(input, "1.0.0", "2.0.0", 1)

	got, err := writeJSON([]byte(input), "version", "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The release commit must not contain a massive unrelated diff: key
	// order, 4-space indent, and raw & must all survive.
	if string(got) != want {
		t.Errorf("formatting destroyed:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestWriteYAML_PreservesCommentsAndOrder(t *testing.T) {
	input := "# my chart\nname: my-app # app name\nversion: 1.0.0\nnotes: keep\n"
	want := strings.Replace(input, "1.0.0", "2.0.0", 1)

	got, err := writeYAML([]byte(input), "version", "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != want {
		t.Errorf("comments/order destroyed:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestWriteTOML_PreservesLayoutAndComments(t *testing.T) {
	input := "# cargo manifest\n[package]\nname = \"my-app\" # the name\nversion = \"1.0.0\"\nedition = \"2021\"\n"
	want := strings.Replace(input, "1.0.0", "2.0.0", 1)

	got, err := writeTOML([]byte(input), "package.version", "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != want {
		t.Errorf("layout destroyed:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestWriteJSON_NestedPath_OnlyTargetChanges(t *testing.T) {
	input := "{\n  \"version\": \"1.0.0\",\n  \"nested\": {\n    \"version\": \"1.0.0\"\n  }\n}\n"

	got, err := writeJSON([]byte(input), "nested.version", "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var tree map[string]interface{}
	if err := json.Unmarshal(got, &tree); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, got)
	}
	if tree["version"] != "1.0.0" {
		t.Errorf("top-level version must stay 1.0.0, got %v", tree["version"])
	}
	nested := tree["nested"].(map[string]interface{})
	if nested["version"] != "2.0.0" {
		t.Errorf("nested.version must become 2.0.0, got %v", nested["version"])
	}
}

func TestWriteYAML_NonStringValue_FallsBackButUpdates(t *testing.T) {
	// version parses as a float — the textual value ("1.0") can't be matched
	// as the string "1"; the full re-marshal fallback must still update it.
	input := "version: 1.0\nkeep: yes\n"

	got, err := writeYAML([]byte(input), "version", "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var tree map[string]interface{}
	if err := yaml.Unmarshal(got, &tree); err != nil {
		t.Fatalf("output is not valid YAML: %v\n%s", err, got)
	}
	if tree["version"] != "2.0.0" {
		t.Errorf("version = %v, want 2.0.0 via fallback", tree["version"])
	}
}
