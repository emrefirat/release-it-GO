package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCfg(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// --- unknown keys: typos must fail loudly, in every format ---

func TestLoadConfig_UnknownKey_IsError_AllFormats(t *testing.T) {
	cases := map[string]string{
		".release-it-go.json": `{"github": {"relase": true}}`,
		".release-it-go.yaml": "github:\n  relase: true\n",
		".release-it-go.toml": "[github]\nrelase = true\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := LoadConfig(writeCfg(t, name, content))
			if err == nil {
				t.Fatal("a misspelled key must not be silently ignored (github.release stayed false)")
			}
			msg := err.Error()
			if !strings.Contains(msg, "relase") {
				t.Errorf("error should name the unknown key, got: %v", err)
			}
			if !strings.Contains(msg, "release") || !strings.Contains(strings.ToLower(msg), "did you mean") {
				t.Errorf("error should suggest the closest valid key, got: %v", err)
			}
		})
	}
}

func TestLoadConfig_HookKeyTypo_SuggestsKebabCase(t *testing.T) {
	// The rest of the config is camelCase; git hook keys are the hook file
	// names. hooks.preCommit is the most likely mistake and used to vanish.
	_, err := LoadConfig(writeCfg(t, ".release-it-go.json", `{"hooks": {"preCommit": ["go vet ./..."]}}`))
	if err == nil {
		t.Fatal("hooks.preCommit must be rejected as unknown")
	}
	if !strings.Contains(err.Error(), "pre-commit") {
		t.Errorf("error should suggest pre-commit, got: %v", err)
	}
}

func TestLoadConfig_LegacyNpmKeys_WarnInsteadOfError(t *testing.T) {
	content := `{
		"npm": {"publish": false},
		"git": {"changelogFile": "HISTORY.md", "commit": false},
		"github": {"release": true}
	}`
	cfg, err := LoadConfig(writeCfg(t, ".release-it.json", content))
	if err != nil {
		t.Fatalf("npm-only keys must keep loading (migration path), got: %v", err)
	}
	if cfg.Git.Commit || !cfg.GitHub.Release {
		t.Error("real settings next to legacy keys must still apply")
	}
	joined := strings.Join(cfg.Warnings, "\n")
	if !strings.Contains(joined, "npm") || !strings.Contains(joined, "changelogFile") {
		t.Errorf("ignored legacy keys must be reported as warnings, got: %v", cfg.Warnings)
	}
}

func TestLoadConfig_YAMLLegacy_GetsNpmCompatToo(t *testing.T) {
	// Compat used to run for JSON only: a YAML legacy file with a
	// requireBranch array hard-failed and plugin settings were ignored.
	content := `git:
  requireBranch: [main, master]
plugins:
  "@release-it/conventional-changelog":
    infile: HISTORY.md
    preset: angular
`
	cfg, err := LoadConfig(writeCfg(t, ".release-it.yaml", content))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Git.RequireBranch != "main,master" {
		t.Errorf("requireBranch = %q, want main,master", cfg.Git.RequireBranch)
	}
	if cfg.Changelog.Infile != "HISTORY.md" {
		t.Errorf("plugin infile not applied for YAML, got %q", cfg.Changelog.Infile)
	}
}

func TestLoadConfig_HookStringValue_IsNotSplitOnCommas(t *testing.T) {
	content := "hooks:\n  \"after:release\": \"echo Released, deploy next\"\n"
	cfg, err := LoadConfig(writeCfg(t, ".release-it-go.yaml", content))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	// viper's default string→slice hook split this into two commands
	if len(cfg.Hooks.AfterRelease) != 1 || cfg.Hooks.AfterRelease[0] != "echo Released, deploy next" {
		t.Errorf("after:release = %q, want the single npm-style string command", cfg.Hooks.AfterRelease)
	}
}

func TestLoadConfig_AllPipelineHookKeys_Load(t *testing.T) {
	content := `hooks:
  "before:prerequisites": ["a"]
  "after:prerequisites": ["b"]
  "before:commitlint": ["c"]
  "after:commitlint": ["d"]
  "before:version": ["e"]
  "after:version": ["f"]
  "before:changelog": ["g"]
  "after:changelog": ["h"]
  "before:notification": ["i"]
  "after:notification": ["j"]
`
	cfg, err := LoadConfig(writeCfg(t, ".release-it-go.yaml", content))
	if err != nil {
		t.Fatalf("the runner fires these events; the config must accept them: %v", err)
	}
	got := []string{
		cfg.Hooks.BeforePrerequisites[0], cfg.Hooks.AfterPrerequisites[0],
		cfg.Hooks.BeforeCommitlint[0], cfg.Hooks.AfterCommitlint[0],
		cfg.Hooks.BeforeVersion[0], cfg.Hooks.AfterVersion[0],
		cfg.Hooks.BeforeChangelog[0], cfg.Hooks.AfterChangelog[0],
		cfg.Hooks.BeforeNotification[0], cfg.Hooks.AfterNotification[0],
	}
	if strings.Join(got, "") != "abcdefghij" {
		t.Errorf("hook fields loaded out of order or empty: %v", got)
	}
}

func TestLoadConfig_TagNameExplicit_StillDetected(t *testing.T) {
	cfg, err := LoadConfig(writeCfg(t, ".release-it-go.yaml", "git:\n  tagName: \"${version}\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Git.TagNameExplicit {
		t.Error("TagNameExplicit must survive the loader refactor")
	}
}

// --- Validate(): reject values the pipeline would only choke on later ---

func TestValidate_Rules(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *Config)
		wantSub string // substring of the error; "" = must be valid
	}{
		{"defaults are valid", func(c *Config) {}, ""},
		{"tagName without ${version}", func(c *Config) { c.Git.TagName = "release" }, "git.tagName"},
		{"increment keyword ok", func(c *Config) { c.Increment = "minor" }, ""},
		{"increment explicit version ok", func(c *Config) { c.Increment = "1.5.0" }, ""},
		{"increment bogus", func(c *Config) { c.Increment = "bogus" }, "increment"},
		{"preReleaseId ok", func(c *Config) { c.PreReleaseID = "beta" }, ""},
		{"preReleaseId invalid chars", func(c *Config) { c.PreReleaseID = "beta!" }, "preReleaseId"},
		{"calver format ok", func(c *Config) { c.CalVer.Enabled = true; c.CalVer.Format = "yyyy.mm.dd" }, ""},
		{"calver format unsupported", func(c *Config) { c.CalVer.Enabled = true; c.CalVer.Format = "yy.mm.dd" }, "calver.format"},
		{"github.host with scheme", func(c *Config) { c.GitHub.Host = "https://ghe.corp" }, "github.host"},
		{"gitlab.origin without scheme", func(c *Config) { c.GitLab.Origin = "gitlab.corp" }, "gitlab.origin"},
		{"gitlab.origin ok", func(c *Config) { c.GitLab.Origin = "https://gitlab.corp" }, ""},
		{"github.timeout negative", func(c *Config) { c.GitHub.Timeout = -5 }, "github.timeout"},
		{"webhook type unknown", func(c *Config) {
			c.Notification.Webhooks = []WebhookConfig{{Type: "discord", URLRef: "X"}}
		}, "webhooks[0].type"},
		{"webhook timeout negative", func(c *Config) {
			c.Notification.Webhooks = []WebhookConfig{{Type: "slack", URLRef: "X", Timeout: -1}}
		}, "webhooks[0].timeout"},
		{"bumper out type unknown", func(c *Config) {
			c.Bumper.Out = []BumperFile{{File: "x.xml", Type: "xml"}}
		}, "bumper.out[0].type"},
		{"changelog preset unsupported", func(c *Config) { c.Changelog.Preset = "eslint" }, "changelog.preset"},
		{"changelog preset conventionalcommits ok", func(c *Config) { c.Changelog.Preset = "conventionalcommits" }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.mutate(cfg)
			err := cfg.Validate()
			if tt.wantSub == "" {
				if err != nil {
					t.Errorf("expected valid, got: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected an error mentioning %q", tt.wantSub)
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("error should name %q, got: %v", tt.wantSub, err)
			}
		})
	}
}

func TestLoadConfig_InvalidValue_FailsAtLoad(t *testing.T) {
	// The file path must be in the message so the user knows what to fix.
	path := writeCfg(t, ".release-it-go.yaml", "git:\n  tagName: release\n")
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("invalid config values must fail at load time, not deep in the pipeline")
	}
	if !strings.Contains(err.Error(), "tagName") {
		t.Errorf("error should name the field, got: %v", err)
	}
}

func TestLoadConfig_RemovedKeys_WarnAndLoad(t *testing.T) {
	content := `{
		"git": {"changelog": "git log", "commit": false},
		"github": {"web": true, "comments": {"submit": true}},
		"gitlab": {"preRelease": true},
		"changelog": {"addUnreleased": true},
		"calver": {"fallbackIncrement": "minor"},
		"bumper": {"out": [{"file": "x.json", "versionPrefix": "v"}]}
	}`
	cfg, err := LoadConfig(writeCfg(t, ".release-it.json", content))
	if err != nil {
		t.Fatalf("removed keys must keep old configs loadable, got: %v", err)
	}
	if cfg.Git.Commit {
		t.Error("real settings next to removed keys must still apply")
	}
	joined := strings.Join(cfg.Warnings, "\n")
	for _, key := range []string{"git.changelog", "github.web", "github.comments", "gitlab.preRelease", "changelog.addUnreleased", "calver.fallbackIncrement", "versionPrefix"} {
		if !strings.Contains(joined, key) {
			t.Errorf("expected a warning naming %s, warnings were:\n%s", key, joined)
		}
	}
}

func TestDefaultConfig_WiredFieldDefaults(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Changelog.AddVersionUrl {
		t.Error("changelog.addVersionUrl must default to true (compare links were always emitted before)")
	}
	if !cfg.GitLab.UseGenericPackageRepositoryForAssets {
		t.Error("gitlab.useGenericPackageRepositoryForAssets must default to true to keep existing upload behavior")
	}
}

// QA: the unknown-key error must read as our message only — no decoder
// header ("decoding failed due to the following error(s):") leaking through.
func TestLoadConfig_UnknownKey_ErrorHasNoDecoderNoise(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".release-it.yaml")
	if err := os.WriteFile(path, []byte("github:\n  relase: true\nhooks:\n  preCommit: [\"go vet\"]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected an unknown-key error")
	}
	msg := err.Error()
	for _, noise := range []string{"decoding failed", "error(s)", ":;"} {
		if strings.Contains(msg, noise) {
			t.Errorf("error leaks decoder text %q: %s", noise, msg)
		}
	}
	want := `unknown config key "github.relase" (did you mean "release"?); unknown config key "hooks.precommit" (did you mean "pre-commit"?)`
	if !strings.HasSuffix(msg, want) {
		t.Errorf("error = %q\nwant suffix %q", msg, want)
	}
}
