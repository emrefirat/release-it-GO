package config

import (
	"os"
	"path/filepath"
	"strings"
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

// --- npm config keys that used to fail the strict unknown-key check ---

func TestLoadConfigFromBytes_SchemaKey_AcceptedSilently(t *testing.T) {
	data := []byte(`{"$schema": "https://unpkg.com/release-it/schema/release-it.json", "git": {"tagName": "v${version}"}}`)
	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("$schema is editor metadata and must not be an error: %v", err)
	}
	if len(cfg.Warnings) != 0 {
		t.Errorf("$schema must not even warn, got %v", cfg.Warnings)
	}
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("tagName = %q", cfg.Git.TagName)
	}
}

func TestLoadConfigFromBytes_NpmOnlyKeys_WarnInsteadOfFailing(t *testing.T) {
	data := []byte(`{"preReleaseBase": 1, "github": {"release": true, "update": true}}`)
	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("npm-only keys must be ignored with a warning, got error: %v", err)
	}
	if !cfg.GitHub.Release {
		t.Error("github.release should still load")
	}
	joined := strings.Join(cfg.Warnings, "\n")
	for _, want := range []string{`"preReleaseBase"`, `"github.update"`} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings should mention %s, got:\n%s", want, joined)
		}
	}
}

func TestLoadConfigFromBytes_PluginScopedHooks_WarnAndKeepTheRest(t *testing.T) {
	data := []byte(`{"hooks": {
		"before:init": ["echo init"],
		"after:git:bump": ["go generate ./..."],
		"before:github:init": ["echo gh"],
		"before:npm:release": ["npm test"],
		"after:git:release": ["echo pushed"]
	}}`)
	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("plugin-scoped hooks must not be fatal: %v", err)
	}
	if len(cfg.Hooks.BeforeInit) != 1 || len(cfg.Hooks.AfterGitRelease) != 1 {
		t.Errorf("supported hooks must still load: %+v", cfg.Hooks)
	}
	if len(cfg.Warnings) != 3 {
		t.Fatalf("expected 3 warnings, got %d: %v", len(cfg.Warnings), cfg.Warnings)
	}
	joined := strings.Join(cfg.Warnings, "\n")
	if !strings.Contains(joined, `"hooks.after:git:bump"`) || !strings.Contains(joined, `use "after:bump"`) {
		t.Errorf("after:git:bump warning should point at after:bump, got:\n%s", joined)
	}
	if !strings.Contains(joined, `"hooks.before:github:init"`) || !strings.Contains(joined, `use "before:init"`) {
		t.Errorf("before:github:init warning should point at before:init, got:\n%s", joined)
	}
	if !strings.Contains(joined, `"hooks.before:npm:release"`) || !strings.Contains(joined, "npm publishing") {
		t.Errorf("npm hook warning should explain npm is unsupported, got:\n%s", joined)
	}
}

func TestLoadConfigFromBytes_PluginScopedHooks_YAML(t *testing.T) {
	data := []byte("hooks:\n  after:git:bump:\n    - go generate ./...\n")
	cfg, err := LoadConfigFromBytes(data, "yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %v", cfg.Warnings)
	}
}

func TestApplyPluginCompat_Bumper(t *testing.T) {
	data := []byte(`{"plugins": {"@release-it/bumper": {
		"in": "VERSION",
		"out": [
			{"file": "package.json", "path": "version"},
			{"file": "chart/Chart.yaml", "path": "appVersion", "prefix": "v"},
			"VERSION"
		]
	}}}`)
	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Bumper.Enabled {
		t.Fatal("bumper plugin must enable the built-in bumper")
	}
	if cfg.Bumper.In == nil || cfg.Bumper.In.File != "VERSION" {
		t.Errorf("bumper.in = %+v, want VERSION", cfg.Bumper.In)
	}
	if len(cfg.Bumper.Out) != 3 {
		t.Fatalf("bumper.out = %+v, want 3 entries", cfg.Bumper.Out)
	}
	if cfg.Bumper.Out[0].File != "package.json" || cfg.Bumper.Out[0].Path != "version" {
		t.Errorf("out[0] = %+v", cfg.Bumper.Out[0])
	}
	if cfg.Bumper.Out[1].Prefix != "v" || cfg.Bumper.Out[1].Path != "appVersion" {
		t.Errorf("out[1] = %+v", cfg.Bumper.Out[1])
	}
	if cfg.Bumper.Out[2].File != "VERSION" {
		t.Errorf("out[2] = %+v", cfg.Bumper.Out[2])
	}
	if len(cfg.Warnings) != 0 {
		t.Errorf("a mapped plugin must not warn, got %v", cfg.Warnings)
	}
}

func TestApplyPluginCompat_BumperObjectIn_SingleOut(t *testing.T) {
	data := []byte(`{"plugins": {"@release-it/bumper": {
		"in": {"file": "package.json", "path": "version"},
		"out": {"file": "VERSION", "consumeWholeFile": true}
	}}}`)
	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Bumper.In == nil || cfg.Bumper.In.Path != "version" {
		t.Errorf("bumper.in = %+v", cfg.Bumper.In)
	}
	if len(cfg.Bumper.Out) != 1 || !cfg.Bumper.Out[0].ConsumeWholeFile {
		t.Errorf("bumper.out = %+v", cfg.Bumper.Out)
	}
}

func TestApplyPluginCompat_UnknownPlugin_Warns(t *testing.T) {
	data := []byte(`{"plugins": {"release-it-calver-plugin": {"format": "yy.mm"}}}`)
	cfg, err := LoadConfigFromBytes(data, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Warnings) != 1 || !strings.Contains(cfg.Warnings[0], "release-it-calver-plugin") {
		t.Errorf("unsupported plugins must be reported, got %v", cfg.Warnings)
	}
}

func TestLoadConfigFromBytes_TopLevelTypo_SuggestsKey(t *testing.T) {
	_, err := LoadConfigFromBytes([]byte(`{"incremnt": "minor"}`), "json")
	if err == nil {
		t.Fatal("expected an unknown-key error")
	}
	want := `unknown config key "incremnt" (did you mean "increment"?)`
	if err.Error() != want {
		t.Errorf("error = %q\nwant  %q", err.Error(), want)
	}
}
