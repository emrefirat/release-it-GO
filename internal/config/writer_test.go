package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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

func TestWriteFullExampleJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it-go-full.json")

	if err := WriteFullExampleJSON(path); err != nil {
		t.Fatalf("WriteFullExampleJSON failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading written file: %v", err)
	}

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("parsing written JSON: %v", err)
	}

	// Full config should have all top-level sections
	expectedSections := []string{"git", "github", "gitlab", "hooks", "changelog", "bumper", "calver", "notification"}
	for _, section := range expectedSections {
		if _, ok := result[section]; !ok {
			t.Errorf("expected section %q in full config", section)
		}
	}

	// Should NOT have runtime flags
	for _, flag := range []string{"ci", "dry-run", "verbose", "increment", "preReleaseId"} {
		if _, ok := result[flag]; ok {
			t.Errorf("full example should not contain runtime flag %q", flag)
		}
	}
}

func TestFullExampleJSON_IsLoadable(t *testing.T) {
	// Verify the curated JSON can be loaded as a valid config
	cfg, err := LoadConfigFromBytes([]byte(fullExampleJSON), "json")
	if err != nil {
		t.Fatalf("fullExampleJSON is not loadable: %v", err)
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
