package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig_GitDefaults(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"git.commit", cfg.Git.Commit, true},
		{"git.commitMessage", cfg.Git.CommitMessage, "chore: release v${version}"},
		{"git.tag", cfg.Git.Tag, true},
		{"git.tagName", cfg.Git.TagName, "${version}"},
		{"git.tagAnnotation", cfg.Git.TagAnnotation, "Release ${version}"},
		{"git.push", cfg.Git.Push, true},
		{"git.pushRepo", cfg.Git.PushRepo, "origin"},
		{"git.requireCleanWorkingDir", cfg.Git.RequireCleanWorkingDir, true},
		{"git.requireUpstream", cfg.Git.RequireUpstream, true},
		{"git.requireCommits", cfg.Git.RequireCommits, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %v, expected %v", tt.got, tt.expected)
			}
		})
	}

	if len(cfg.Git.PushArgs) != 1 || cfg.Git.PushArgs[0] != "--follow-tags" {
		t.Errorf("git.pushArgs = %v, expected [--follow-tags]", cfg.Git.PushArgs)
	}
}

func TestDefaultConfig_GitHubDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.GitHub.Release {
		t.Error("github.release should be false by default")
	}
	if cfg.GitHub.ReleaseName != "Release ${version}" {
		t.Errorf("github.releaseName = %q, expected %q", cfg.GitHub.ReleaseName, "Release ${version}")
	}
	if !cfg.GitHub.MakeLatest {
		t.Error("github.makeLatest should be true by default")
	}
	if cfg.GitHub.Host != "api.github.com" {
		t.Errorf("github.host = %q, expected %q", cfg.GitHub.Host, "api.github.com")
	}
	if cfg.GitHub.TokenRef != "GITHUB_TOKEN" {
		t.Errorf("github.tokenRef = %q, expected %q", cfg.GitHub.TokenRef, "GITHUB_TOKEN")
	}
}

func TestDefaultConfig_GitLabDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.GitLab.Release {
		t.Error("gitlab.release should be false by default")
	}
	if cfg.GitLab.TokenRef != "GITLAB_TOKEN" {
		t.Errorf("gitlab.tokenRef = %q, expected %q", cfg.GitLab.TokenRef, "GITLAB_TOKEN")
	}
	if cfg.GitLab.TokenHeader != "Private-Token" {
		t.Errorf("gitlab.tokenHeader = %q, expected %q", cfg.GitLab.TokenHeader, "Private-Token")
	}
}

func TestDefaultConfig_ChangelogDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Changelog.Enabled {
		t.Error("changelog.enabled should be true by default")
	}
	if cfg.Changelog.Preset != "angular" {
		t.Errorf("changelog.preset = %q, expected %q", cfg.Changelog.Preset, "angular")
	}
	if cfg.Changelog.Infile != "CHANGELOG.md" {
		t.Errorf("changelog.infile = %q, expected %q", cfg.Changelog.Infile, "CHANGELOG.md")
	}
	if cfg.Changelog.Header != "# Changelog" {
		t.Errorf("changelog.header = %q, expected %q", cfg.Changelog.Header, "# Changelog")
	}
}

func TestDefaultConfig_CalVerDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CalVer.Format != "yy.mm.minor" {
		t.Errorf("calver.format = %q, expected %q", cfg.CalVer.Format, "yy.mm.minor")
	}
	if cfg.CalVer.Increment != "calendar" {
		t.Errorf("calver.increment = %q, expected %q", cfg.CalVer.Increment, "calendar")
	}
	if cfg.CalVer.FallbackIncrement != "minor" {
		t.Errorf("calver.fallbackIncrement = %q, expected %q", cfg.CalVer.FallbackIncrement, "minor")
	}
}

func TestDefaultConfig_TopLevelDefaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CI {
		t.Error("ci should be false by default")
	}
	if cfg.DryRun {
		t.Error("dry-run should be false by default")
	}
	if cfg.Verbose != 0 {
		t.Errorf("verbose = %d, expected 0", cfg.Verbose)
	}
}

func TestLoadConfig_JSONFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".release-it.json")

	content := `{
		"git": {
			"commit": false,
			"tagName": "v${version}",
			"pushRepo": "upstream"
		},
		"github": {
			"release": true
		}
	}`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Git.Commit {
		t.Error("git.commit should be false")
	}
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("git.tagName = %q, expected %q", cfg.Git.TagName, "v${version}")
	}
	if cfg.Git.PushRepo != "upstream" {
		t.Errorf("git.pushRepo = %q, expected %q", cfg.Git.PushRepo, "upstream")
	}
	if !cfg.GitHub.Release {
		t.Error("github.release should be true")
	}
	// Defaults should be preserved for unset fields
	if cfg.GitHub.Host != "api.github.com" {
		t.Errorf("github.host should keep default, got %q", cfg.GitHub.Host)
	}
}

func TestLoadConfig_YAMLFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".release-it.yaml")

	content := `git:
  commit: false
  tagName: "v${version}"
github:
  release: true
  host: "github.example.com"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Git.Commit {
		t.Error("git.commit should be false")
	}
	if !cfg.GitHub.Release {
		t.Error("github.release should be true")
	}
	if cfg.GitHub.Host != "github.example.com" {
		t.Errorf("github.host = %q, expected %q", cfg.GitHub.Host, "github.example.com")
	}
}

func TestLoadConfig_TOMLFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".release-it.toml")

	content := `[git]
commit = false
tagName = "v${version}"

[github]
release = true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Git.Commit {
		t.Error("git.commit should be false")
	}
	if !cfg.GitHub.Release {
		t.Error("github.release should be true")
	}
}

func TestLoadConfig_NoFile_ReturnsDefaults(t *testing.T) {
	// Change to a temp dir with no config files
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	dir := t.TempDir()
	os.Chdir(dir)

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	expected := DefaultConfig()
	if cfg.Git.Commit != expected.Git.Commit {
		t.Error("should return defaults when no config file found")
	}
}

func TestLoadConfig_InvalidFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".release-it.json")

	if err := os.WriteFile(cfgPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(cfgPath)
	if err == nil {
		t.Error("expected error for invalid JSON file")
	}
}

func TestLoadConfig_NonexistentFile_ReturnsError(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/.release-it.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadConfigFromBytes_JSON(t *testing.T) {
	data := []byte(`{"git": {"commit": false}}`)

	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("LoadConfigFromBytes failed: %v", err)
	}

	if cfg.Git.Commit {
		t.Error("git.commit should be false")
	}
}

func TestLoadConfigFromBytes_YAML(t *testing.T) {
	data := []byte("git:\n  commit: false\n")

	cfg, err := LoadConfigFromBytes(data, "yaml")
	if err != nil {
		t.Fatalf("LoadConfigFromBytes failed: %v", err)
	}

	if cfg.Git.Commit {
		t.Error("git.commit should be false")
	}
}

func TestLoadConfigFromBytes_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid}`)
	_, err := LoadConfigFromBytes(data, "json")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadConfig_NativeConfigPriority(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create both legacy and native config with different values
	legacy := `{"git": {"tagName": "legacy-${version}"}}`
	native := `{"git": {"tagName": "native-${version}"}}`

	if err := os.WriteFile(".release-it.json", []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".release-it-go.json", []byte(native), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Native config should win
	if cfg.Git.TagName != "native-${version}" {
		t.Errorf("expected native config to take priority, got tagName=%s", cfg.Git.TagName)
	}
}

func TestLoadConfig_FallsBackToLegacy(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create only legacy config
	legacy := `{"git": {"tagName": "legacy-${version}"}}`
	if err := os.WriteFile(".release-it.json", []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Git.TagName != "legacy-${version}" {
		t.Errorf("expected legacy config to be loaded, got tagName=%s", cfg.Git.TagName)
	}
}

func TestConfigSearchFiles_NativeFirst(t *testing.T) {
	// Verify the search order has native files before legacy files
	nativeFound := false
	for _, f := range configSearchFiles {
		if f == ".release-it-go.json" {
			nativeFound = true
		}
		if f == ".release-it.json" && !nativeFound {
			t.Error(".release-it.json appears before .release-it-go.json in search order")
		}
	}
	if !nativeFound {
		t.Error(".release-it-go.json not found in configSearchFiles")
	}
}
