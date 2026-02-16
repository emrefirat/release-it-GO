package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestWriteConfigJSON_DefaultConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go.json")

	cfg := DefaultConfig()

	if err := WriteConfigJSON(cfg, path); err != nil {
		t.Fatalf("WriteConfigJSON failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	// Default config should produce an empty JSON object (all fields match defaults)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parsing written JSON: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty JSON for default config, got %d keys: %v", len(result), result)
	}
}

func TestWriteConfigJSON_NonDefaultFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go.json")

	cfg := DefaultConfig()
	cfg.GitHub.Release = true
	cfg.Git.TagName = "v${version}"
	cfg.Git.RequireBranch = "main"

	if err := WriteConfigJSON(cfg, path); err != nil {
		t.Fatalf("WriteConfigJSON failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parsing written JSON: %v", err)
	}

	// github.release=true is non-default
	ghMap, ok := result["github"].(map[string]interface{})
	if !ok {
		t.Fatal("expected github section in output")
	}
	if release, ok := ghMap["release"].(bool); !ok || !release {
		t.Error("expected github.release=true in output")
	}

	// git.tagName="v${version}" differs from default "${version}"
	gitMap, ok := result["git"].(map[string]interface{})
	if !ok {
		t.Fatal("expected git section in output")
	}
	if tagName, ok := gitMap["tagName"].(string); !ok || tagName != "v${version}" {
		t.Errorf("expected git.tagName=v${version}, got %v", gitMap["tagName"])
	}

	// git.requireBranch="main" differs from default ""
	if rb, ok := gitMap["requireBranch"].(string); !ok || rb != "main" {
		t.Errorf("expected git.requireBranch=main, got %v", gitMap["requireBranch"])
	}
}

func TestWriteConfigJSON_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go.json")

	cfg := DefaultConfig()
	if err := WriteConfigJSON(cfg, path); err != nil {
		t.Fatalf("WriteConfigJSON failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	// Check file is readable/writable
	perm := info.Mode().Perm()
	if perm&0600 != 0600 {
		t.Errorf("expected at least 0600 permissions, got %o", perm)
	}
}

func TestWriteConfigJSON_InvalidPath(t *testing.T) {
	cfg := DefaultConfig()
	err := WriteConfigJSON(cfg, "/nonexistent/dir/.release-it-go.json")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestWriteConfigJSON_ChangedChangelog(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go.json")

	cfg := DefaultConfig()
	cfg.Changelog.KeepAChangelog = true

	if err := WriteConfigJSON(cfg, path); err != nil {
		t.Fatalf("WriteConfigJSON failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parsing written JSON: %v", err)
	}

	clMap, ok := result["changelog"].(map[string]interface{})
	if !ok {
		t.Fatal("expected changelog section in output")
	}
	if kac, ok := clMap["keepAChangelog"].(bool); !ok || !kac {
		t.Error("expected changelog.keepAChangelog=true in output")
	}
}

func TestWriteConfigYAML_DefaultConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go.yaml")

	cfg := DefaultConfig()

	if err := WriteConfigYAML(cfg, path); err != nil {
		t.Fatalf("WriteConfigYAML failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	// Default config should produce an empty YAML object
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("parsing written YAML: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty YAML for default config, got %d keys: %v", len(result), result)
	}
}

func TestWriteConfigYAML_NonDefaultFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go.yaml")

	cfg := DefaultConfig()
	cfg.GitHub.Release = true
	cfg.Git.TagName = "v${version}"
	cfg.Git.RequireBranch = "main"

	if err := WriteConfigYAML(cfg, path); err != nil {
		t.Fatalf("WriteConfigYAML failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("parsing written YAML: %v", err)
	}

	// github.release=true is non-default
	ghMap, ok := result["github"].(map[string]interface{})
	if !ok {
		t.Fatal("expected github section in output")
	}
	if release, ok := ghMap["release"].(bool); !ok || !release {
		t.Error("expected github.release=true in output")
	}

	// git section
	gitMap, ok := result["git"].(map[string]interface{})
	if !ok {
		t.Fatal("expected git section in output")
	}
	if tagName, ok := gitMap["tagName"].(string); !ok || tagName != "v${version}" {
		t.Errorf("expected git.tagName=v${version}, got %v", gitMap["tagName"])
	}
	if rb, ok := gitMap["requireBranch"].(string); !ok || rb != "main" {
		t.Errorf("expected git.requireBranch=main, got %v", gitMap["requireBranch"])
	}
}

func TestWriteFullExampleYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go-full.yaml")

	if err := WriteFullExampleYAML(path); err != nil {
		t.Fatalf("WriteFullExampleYAML failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	content := string(data)

	// Should be valid YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("parsing written YAML: %v", err)
	}

	// Full config should have all top-level sections
	expectedSections := []string{"git", "github", "gitlab", "hooks", "changelog", "bumper", "calver", "notification"}
	for _, section := range expectedSections {
		if _, ok := result[section]; !ok {
			t.Errorf("expected section %q in full config", section)
		}
	}

	// Should contain comment lines (YAML advantage over JSON)
	if !strings.Contains(content, "# ") {
		t.Error("expected YAML comments in full example")
	}
	if !strings.Contains(content, "# Git commit, tag and push settings") {
		t.Error("expected section documentation comments")
	}

	// Should NOT have runtime flags
	for _, flag := range []string{"ci:", "dry-run:", "verbose:", "increment:", "preReleaseId:"} {
		if strings.Contains(content, "\n"+flag) {
			t.Errorf("full example should not contain runtime flag %q", flag)
		}
	}
}

func TestFullExampleYAML_IsLoadable(t *testing.T) {
	cfg, err := LoadConfigFromBytes([]byte(fullExampleYAML), "yaml")
	if err != nil {
		t.Fatalf("fullExampleYAML is not loadable: %v", err)
	}

	if !cfg.GitHub.Release {
		t.Error("expected github.release=true in full example")
	}
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("expected git.tagName=v${version}, got %s", cfg.Git.TagName)
	}
	if len(cfg.Notification.Webhooks) == 0 {
		t.Error("expected notification.webhooks to have entries")
	}
}

func TestToMinimalMap_EmptyForDefaults(t *testing.T) {
	cfg := DefaultConfig()
	m := toMinimalMap(cfg)

	if len(m) != 0 {
		t.Errorf("expected empty map for default config, got %v", m)
	}
}

func TestToConfigMap_ForceFieldsIncludesDefaults(t *testing.T) {
	cfg := DefaultConfig()

	// Without force: default config produces empty map
	m := toConfigMap(cfg, nil)
	if len(m) != 0 {
		t.Errorf("expected empty map without force, got %v", m)
	}

	// With force: specified fields appear even though they match defaults
	force := ForceFields{
		"git":       {"commit": true, "tag": true, "push": true},
		"changelog": {"enabled": true, "infile": true},
	}
	m = toConfigMap(cfg, force)

	gitMap, ok := m["git"].(map[string]interface{})
	if !ok {
		t.Fatal("expected git section with force fields")
	}
	for _, key := range []string{"commit", "tag", "push"} {
		if _, ok := gitMap[key]; !ok {
			t.Errorf("expected git.%s in forced output", key)
		}
	}

	clMap, ok := m["changelog"].(map[string]interface{})
	if !ok {
		t.Fatal("expected changelog section with force fields")
	}
	for _, key := range []string{"enabled", "infile"} {
		if _, ok := clMap[key]; !ok {
			t.Errorf("expected changelog.%s in forced output", key)
		}
	}

	// Sections without force fields should still be absent
	if _, ok := m["github"]; ok {
		t.Error("github section should not appear without force or diff")
	}
}
