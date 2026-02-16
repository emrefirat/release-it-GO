package runner

import (
	"testing"

	"release-it-go/internal/config"
	"release-it-go/internal/git"
)

func TestNewReleaseContext(t *testing.T) {
	cfg := &config.Config{
		DryRun:  true,
		Verbose: 1,
		CI:      true,
	}

	ctx := NewReleaseContext(cfg)
	if ctx == nil {
		t.Fatal("expected non-nil ReleaseContext")
	}
	if !ctx.IsDryRun {
		t.Error("expected IsDryRun to be true")
	}
	if !ctx.IsCI {
		t.Error("expected IsCI to be true")
	}
	if ctx.Config != cfg {
		t.Error("expected Config to match")
	}
	if ctx.Logger == nil {
		t.Error("expected non-nil Logger")
	}
	if ctx.Git == nil {
		t.Error("expected non-nil Git")
	}
	if ctx.Prompter == nil {
		t.Error("expected non-nil Prompter")
	}
	if ctx.HookRunner == nil {
		t.Error("expected non-nil HookRunner")
	}
	if ctx.Spinner == nil {
		t.Error("expected non-nil Spinner")
	}
	if ctx.Vars == nil {
		t.Error("expected non-nil Vars")
	}
}

func TestNewReleaseContext_NonCI(t *testing.T) {
	cfg := &config.Config{
		CI: false,
	}

	ctx := NewReleaseContext(cfg)
	// The Prompter type depends on ui.IsCI() too, but we can check it's not nil
	if ctx.Prompter == nil {
		t.Error("expected non-nil Prompter")
	}
}

func TestReleaseContext_UpdateVars(t *testing.T) {
	cfg := &config.Config{}
	ctx := NewReleaseContext(cfg)

	ctx.Version = "1.2.3"
	ctx.LatestVersion = "1.2.2"
	ctx.TagName = "v1.2.3"
	ctx.Changelog = "- feat: new feature"
	ctx.ReleaseURL = "https://github.com/emfi/release-it-go/releases/v1.2.3"
	ctx.BranchName = "main"
	ctx.RepoInfo = &git.RepoInfo{
		Remote:     "https://github.com/emfi/release-it-go.git",
		Protocol:   "https",
		Host:       "github.com",
		Owner:      "emfi",
		Repository: "release-it-go",
	}

	ctx.UpdateVars()

	tests := []struct {
		key      string
		expected string
	}{
		{"version", "1.2.3"},
		{"latestVersion", "1.2.2"},
		{"tagName", "v1.2.3"},
		{"changelog", "- feat: new feature"},
		{"releaseUrl", "https://github.com/emfi/release-it-go/releases/v1.2.3"},
		{"branchName", "main"},
		{"repo.remote", "https://github.com/emfi/release-it-go.git"},
		{"repo.protocol", "https"},
		{"repo.host", "github.com"},
		{"repo.owner", "emfi"},
		{"repo.repository", "release-it-go"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if ctx.Vars[tt.key] != tt.expected {
				t.Errorf("expected Vars[%q] = %q, got %q", tt.key, tt.expected, ctx.Vars[tt.key])
			}
		})
	}
}

func TestReleaseContext_UpdateVars_NilRepoInfo(t *testing.T) {
	cfg := &config.Config{}
	ctx := NewReleaseContext(cfg)

	ctx.Version = "1.0.0"
	ctx.RepoInfo = nil

	// Should not panic
	ctx.UpdateVars()

	if ctx.Vars["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", ctx.Vars["version"])
	}
	if _, ok := ctx.Vars["repo.owner"]; ok {
		t.Error("expected repo.owner to not be set when RepoInfo is nil")
	}
}
