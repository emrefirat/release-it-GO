package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPluginCompat_ConventionalChangelog(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.json")
	content := `{
		"git": {
			"commitMessage": "chore(release): v${version} [skip-ci]",
			"tagName": "${version}",
			"tagAnnotation": "Release v${version}"
		},
		"gitlab": {
			"release": true,
			"releaseName": "v${version}"
		},
		"npm": {
			"publish": false
		},
		"plugins": {
			"@release-it/conventional-changelog": {
				"preset": "angular",
				"infile": "CHANGELOG.md",
				"changelog": true,
				"commit": true
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Git settings from config
	if cfg.Git.CommitMessage != "chore(release): v${version} [skip-ci]" {
		t.Errorf("commitMessage = %q", cfg.Git.CommitMessage)
	}
	if cfg.Git.TagName != "${version}" {
		t.Errorf("tagName = %q", cfg.Git.TagName)
	}

	// GitLab settings
	if !cfg.GitLab.Release {
		t.Error("gitlab.release should be true")
	}
	if cfg.GitLab.ReleaseName != "v${version}" {
		t.Errorf("releaseName = %q", cfg.GitLab.ReleaseName)
	}

	// Plugin compat: conventional-changelog settings mapped to changelog config
	if !cfg.Changelog.Enabled {
		t.Error("changelog.enabled should be true (from plugin changelog: true)")
	}
	if cfg.Changelog.Preset != "angular" {
		t.Errorf("changelog.preset = %q, want 'angular'", cfg.Changelog.Preset)
	}
	if cfg.Changelog.Infile != "CHANGELOG.md" {
		t.Errorf("changelog.infile = %q, want 'CHANGELOG.md'", cfg.Changelog.Infile)
	}
}

func TestApplyPluginCompat_KeepAChangelog(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.json")
	content := `{
		"plugins": {
			"@release-it/keep-a-changelog": {
				"filename": "CHANGES.md",
				"head": "# My Changelog"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if !cfg.Changelog.KeepAChangelog {
		t.Error("changelog.keepAChangelog should be true")
	}
	if cfg.Changelog.Infile != "CHANGES.md" {
		t.Errorf("changelog.infile = %q, want 'CHANGES.md'", cfg.Changelog.Infile)
	}
	if cfg.Changelog.Header != "# My Changelog" {
		t.Errorf("changelog.header = %q, want '# My Changelog'", cfg.Changelog.Header)
	}
}

func TestApplyPluginCompat_NoPlugins(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.json")
	content := `{
		"git": {
			"tagName": "v${version}"
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Defaults should remain
	if cfg.Changelog.Preset != "angular" {
		t.Errorf("default preset = %q, want 'angular'", cfg.Changelog.Preset)
	}
}

func TestApplyPluginCompat_OldNpmFormat(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.json")
	// Real-world npm release-it config with type mismatches
	content := `{
		"versionFile": "version.json",
		"git": {
			"changelog": "git log --pretty=format:\"* %s (%h)\" ${latestTag}...HEAD",
			"changelogFile": "CHANGELOG.md",
			"requireCleanWorkingDir": true,
			"requireUpstream": false,
			"requireCommits": false,
			"requireBranch": [],
			"addUntrackedFiles": false,
			"commit": true,
			"commitMessage": "chore(release): v${version} [skip ci]",
			"commitArgs": [],
			"tag": true,
			"tagName": "v${version}",
			"tagAnnotation": "Release ${version}",
			"tagArgs": [],
			"push": true,
			"pushArgs": ["--follow-tags"],
			"pushRepo": ""
		},
		"github": {
			"release": false,
			"releaseName": "Release ${version}",
			"preRelease": false,
			"draft": false,
			"tokenRef": "GITHUB_TOKEN",
			"assets": [],
			"timeout": 30000000000,
			"web": false,
			"autoGenerate": false
		},
		"gitlab": {
			"release": true,
			"releaseName": "Release ${version}",
			"tokenRef": "GITLAB_TOKEN",
			"assets": {
				"links": []
			}
		},
		"npm": {
			"publish": false
		},
		"hooks": {},
		"plugins": {}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify key fields loaded correctly
	if cfg.Git.CommitMessage != "chore(release): v${version} [skip ci]" {
		t.Errorf("commitMessage = %q", cfg.Git.CommitMessage)
	}
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("tagName = %q", cfg.Git.TagName)
	}
	if !cfg.Git.Push {
		t.Error("push should be true")
	}
	if !cfg.GitLab.Release {
		t.Error("gitlab.release should be true")
	}
	if cfg.Git.RequireBranch != "" {
		t.Errorf("requireBranch = %q, want empty", cfg.Git.RequireBranch)
	}
}

func TestApplyPluginCompat_RequireBranchArray(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.json")
	content := `{
		"git": {
			"requireBranch": ["main", "master"]
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Git.RequireBranch != "main,master" {
		t.Errorf("requireBranch = %q, want 'main,master'", cfg.Git.RequireBranch)
	}
}

func TestApplyPluginCompat_PluginDoesNotOverrideExplicitConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.json")
	// User explicitly sets changelog.enabled: true
	// Plugin does NOT set "changelog" field at all
	content := `{
		"changelog": {"enabled": true},
		"plugins": {
			"@release-it/conventional-changelog": {
				"preset": "conventionalcommits",
				"infile": "CHANGELOG.md"
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// User's explicit config should be preserved
	if !cfg.Changelog.Enabled {
		t.Error("changelog.enabled should remain true (user config takes priority over plugin)")
	}
	if cfg.Changelog.Preset != "conventionalcommits" {
		t.Errorf("changelog.preset = %q, want 'conventionalcommits'", cfg.Changelog.Preset)
	}
}

func TestApplyPluginCompat_PluginExplicitlyDisablesChangelog(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.json")
	// Plugin explicitly sets "changelog": false
	content := `{
		"plugins": {
			"@release-it/conventional-changelog": {
				"preset": "angular",
				"changelog": false
			}
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Changelog.Enabled {
		t.Error("changelog.enabled should be false (plugin explicitly set changelog: false)")
	}
}

func TestLoadConfigFromBytes_JSONNormalization(t *testing.T) {
	// requireBranch as array should be normalized to comma-separated string
	data := []byte(`{"git": {"requireBranch": ["main", "develop"]}}`)

	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("LoadConfigFromBytes failed: %v", err)
	}

	if cfg.Git.RequireBranch != "main,develop" {
		t.Errorf("requireBranch = %q, want 'main,develop'", cfg.Git.RequireBranch)
	}
}

func TestLoadConfigFromBytes_JSONNormalization_PluginsRemoved(t *testing.T) {
	// "plugins" and "npm" keys should be removed during normalization
	data := []byte(`{"git": {"tagName": "v${version}"}, "npm": {"publish": false}, "plugins": {}}`)

	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("LoadConfigFromBytes failed: %v", err)
	}

	if cfg.Git.TagName != "v${version}" {
		t.Errorf("tagName = %q, want 'v${version}'", cfg.Git.TagName)
	}
}

func TestApplyPluginCompat_YAMLIgnored(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".release-it.yaml")
	content := `git:
  tagName: "v${version}"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// YAML doesn't have plugin compat processing, defaults remain
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("tagName = %q", cfg.Git.TagName)
	}
}
