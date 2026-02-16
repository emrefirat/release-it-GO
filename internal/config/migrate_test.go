package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLegacyConfig_Found(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	if err := os.WriteFile(LegacyConfigFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	path, found := DetectLegacyConfig()
	if !found {
		t.Error("expected legacy config to be found")
	}
	if path != LegacyConfigFile {
		t.Errorf("expected path %s, got %s", LegacyConfigFile, path)
	}
}

func TestDetectLegacyConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	_, found := DetectLegacyConfig()
	if found {
		t.Error("expected legacy config not to be found")
	}
}

func TestDetectNativeConfig_Found(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	if err := os.WriteFile(NativeConfigFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	if !DetectNativeConfig() {
		t.Error("expected native config to be found")
	}
}

func TestDetectNativeConfig_NotFound(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	if DetectNativeConfig() {
		t.Error("expected native config not to be found")
	}
}

func TestMigrateLegacyConfig_Basic(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create legacy config with npm-specific fields
	legacyContent := `{
		"git": {
			"requireBranch": ["main", "master"],
			"tagName": "v${version}"
		},
		"github": {
			"release": true
		},
		"plugins": {
			"@release-it/conventional-changelog": {
				"preset": "angular",
				"infile": "CHANGELOG.md",
				"changelog": true
			}
		},
		"npm": {
			"publish": false
		}
	}`

	if err := os.WriteFile(LegacyConfigFile, []byte(legacyContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := MigrateLegacyConfig(LegacyConfigFile); err != nil {
		t.Fatalf("MigrateLegacyConfig failed: %v", err)
	}

	// Check backup was created
	backupPath := LegacyConfigFile + ".bak"
	if !fileExists(backupPath) {
		t.Error("backup file was not created")
	}

	// Check backup content matches original
	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backupData) != legacyContent {
		t.Error("backup content does not match original")
	}

	// Check native config was created
	if !fileExists(NativeConfigFile) {
		t.Fatal("native config file was not created")
	}

	// Load and verify the migrated config
	cfg, err := LoadConfig(filepath.Join(dir, NativeConfigFile))
	if err != nil {
		t.Fatalf("loading migrated config: %v", err)
	}

	if !cfg.GitHub.Release {
		t.Error("expected github.release=true after migration")
	}
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("expected git.tagName=v${version}, got %s", cfg.Git.TagName)
	}
}

func TestMigrateLegacyConfig_NonexistentFile(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	err := MigrateLegacyConfig("nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMigrateLegacyConfig_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	if err := os.WriteFile(LegacyConfigFile, []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	err := MigrateLegacyConfig(LegacyConfigFile)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}

	// Backup should still be created
	if !fileExists(LegacyConfigFile + ".bak") {
		t.Error("backup should be created even on parse error")
	}
}
