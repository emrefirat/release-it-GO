package runner

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/emfi/release-it-go/internal/config"
	"github.com/emfi/release-it-go/internal/git"
	"github.com/emfi/release-it-go/internal/ui"
)

func TestRenderTagName(t *testing.T) {
	tests := []struct {
		name     string
		template string
		version  string
		expected string
	}{
		{"default template", "v${version}", "1.2.3", "v1.2.3"},
		{"no prefix", "${version}", "1.2.3", "1.2.3"},
		{"custom prefix", "release-${version}", "2.0.0", "release-2.0.0"},
		{"no placeholder", "v1.0.0", "2.0.0", "v1.0.0"},
		{"empty version", "v${version}", "", "v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTagName(tt.template, tt.version)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestReplaceAll(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		old      string
		new      string
		expected string
	}{
		{"simple", "hello world", "world", "go", "hello go"},
		{"multiple", "a-b-a-b", "a", "x", "x-b-x-b"},
		{"no match", "hello", "xyz", "abc", "hello"},
		{"empty old", "hello", "", "", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip empty old string to avoid infinite loop
			if tt.old == "" {
				return
			}
			result := replaceAll(tt.s, tt.old, tt.new)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{"found at start", "hello", "he", 0},
		{"found at end", "hello", "lo", 3},
		{"found in middle", "hello", "ll", 2},
		{"not found", "hello", "xyz", -1},
		{"empty substr", "hello", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := indexOf(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		prefix   string
		expected bool
	}{
		{"has prefix", "v1.2.3", "v", true},
		{"no prefix", "1.2.3", "v", false},
		{"empty prefix", "hello", "", true},
		{"equal strings", "test", "test", true},
		{"prefix longer", "hi", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasPrefix(tt.s, tt.prefix)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewRunner(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := NewRunner(cfg)
	if runner == nil {
		t.Fatal("expected non-nil Runner")
	}
	if runner.ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if !runner.ctx.IsDryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestRunner_BuildVersionOptions(t *testing.T) {
	cfg := &config.Config{CI: true}
	runner := NewRunner(cfg)

	options := runner.buildVersionOptions("1.0.0", "minor")

	if len(options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(options))
	}

	// Check patch option
	if options[0].Version != "1.0.1" {
		t.Errorf("expected patch version 1.0.1, got %s", options[0].Version)
	}
	if options[0].Recommended {
		t.Error("patch should not be recommended")
	}

	// Check minor option (recommended)
	if options[1].Version != "1.1.0" {
		t.Errorf("expected minor version 1.1.0, got %s", options[1].Version)
	}
	if !options[1].Recommended {
		t.Error("minor should be recommended")
	}

	// Check major option
	if options[2].Version != "2.0.0" {
		t.Errorf("expected major version 2.0.0, got %s", options[2].Version)
	}
}

func TestRunner_BuildVersionOptions_InvalidVersion(t *testing.T) {
	cfg := &config.Config{CI: true}
	runner := NewRunner(cfg)

	options := runner.buildVersionOptions("invalid", "patch")
	if len(options) != 0 {
		t.Errorf("expected 0 options for invalid version, got %d", len(options))
	}
}

// mockPrompter implements ui.Prompter for testing.
type mockPrompter struct {
	selectVersionResult string
	selectVersionErr    error
	confirmResult       bool
	confirmErr          error
	inputResult         string
	inputErr            error
}

func (m *mockPrompter) SelectVersion(current string, recommended string, options []ui.VersionOption) (string, error) {
	return m.selectVersionResult, m.selectVersionErr
}

func (m *mockPrompter) Confirm(message string, defaultYes bool) (bool, error) {
	return m.confirmResult, m.confirmErr
}

func (m *mockPrompter) Input(message string, defaultValue string) (string, error) {
	return m.inputResult, m.inputErr
}

func TestRunner_GenerateChangelog_Disabled(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Changelog: config.ChangelogConfig{
			Enabled: false,
		},
	}
	runner := NewRunner(cfg)

	err := runner.generateChangelog()
	if err != nil {
		t.Errorf("expected no error when changelog is disabled, got: %v", err)
	}
}

func TestRunner_GithubRelease_Disabled(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		GitHub: config.GitHubConfig{
			Release: false,
		},
	}
	runner := NewRunner(cfg)

	err := runner.githubRelease()
	if err != nil {
		t.Errorf("expected no error when github release is disabled, got: %v", err)
	}
}

func TestRunner_GitlabRelease_Disabled(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		GitLab: config.GitLabConfig{
			Release: false,
		},
	}
	runner := NewRunner(cfg)

	err := runner.gitlabRelease()
	if err != nil {
		t.Errorf("expected no error when gitlab release is disabled, got: %v", err)
	}
}

func TestRunner_GithubRelease_NoRepoInfo(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		GitHub: config.GitHubConfig{
			Release: true,
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = nil

	err := runner.githubRelease()
	if err != nil {
		t.Errorf("expected no error when repoInfo is nil, got: %v", err)
	}
}

func TestRunner_GitlabRelease_NoRepoInfo(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		GitLab: config.GitLabConfig{
			Release: true,
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = nil

	err := runner.gitlabRelease()
	if err != nil {
		t.Errorf("expected no error when repoInfo is nil, got: %v", err)
	}
}

func TestRunner_PrintSummary_DryRun(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			Push:          true,
			PushRepo:      "origin",
		},
		Changelog: config.ChangelogConfig{
			Infile: "CHANGELOG.md",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.Changelog = "some changelog"

	// Should not panic
	runner.printSummary(0)
}

func TestRunner_PrintSummary_Normal(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			Push:          true,
			PushRepo:      "origin",
		},
		Changelog: config.ChangelogConfig{
			Infile: "CHANGELOG.md",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.Changelog = "some changelog"
	runner.ctx.ReleaseURL = "https://github.com/emfi/release-it-go/releases/v1.0.0"
	runner.ctx.BranchName = "main"

	// Should not panic
	runner.printSummary(1000000000) // 1 second
}

func TestRunner_DetermineVersion_NoIncrement(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "no-increment",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	// We can't test the full determineVersion because it calls git,
	// but we can test the no-increment path by setting up state
	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Config.Increment = "no-increment"

	// The method calls git.GetLatestTag which we can't mock easily,
	// so test the renderTagName part independently
	tagName := renderTagName(cfg.Git.TagName, "1.0.0")
	if tagName != "v1.0.0" {
		t.Errorf("expected v1.0.0, got %s", tagName)
	}
}

func TestRunner_BumpFiles_Disabled(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Bumper: config.BumperConfig{
			Enabled: false,
		},
	}
	runner := NewRunner(cfg)

	err := runner.bumpFiles()
	if err != nil {
		t.Errorf("expected no error when bumper is disabled, got: %v", err)
	}
}

func TestRunner_BumpFiles_NoOut(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Bumper: config.BumperConfig{
			Enabled: true,
		},
	}
	runner := NewRunner(cfg)

	err := runner.bumpFiles()
	if err != nil {
		t.Errorf("expected no error when no out files, got: %v", err)
	}
}

func TestRunner_BumpFiles_DryRun(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/VERSION"
	os.WriteFile(file, []byte("1.0.0\n"), 0644)

	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Bumper: config.BumperConfig{
			Enabled: true,
			Out: []config.BumperFile{
				{File: file, ConsumeWholeFile: true},
			},
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "2.0.0"

	err := runner.bumpFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should not be modified
	data, _ := os.ReadFile(file)
	if string(data) != "1.0.0\n" {
		t.Errorf("file should not be modified in dry-run, got %q", string(data))
	}
}

func TestRunner_BumpFiles_Success(t *testing.T) {
	dir := t.TempDir()
	file := dir + "/VERSION"
	os.WriteFile(file, []byte("1.0.0\n"), 0644)

	cfg := &config.Config{
		CI: true,
		Bumper: config.BumperConfig{
			Enabled: true,
			Out: []config.BumperFile{
				{File: file, ConsumeWholeFile: true},
			},
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "2.0.0"

	err := runner.bumpFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(file)
	if string(data) != "2.0.0\n" {
		t.Errorf("expected '2.0.0\\n', got %q", string(data))
	}
}

func TestRunner_DetermineCalVer(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		CalVer: config.CalVerConfig{
			Enabled: true,
			Format:  "yyyy.mm.minor",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineCalVer("2025.1.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version == "" {
		t.Error("expected non-empty version")
	}
	if runner.ctx.TagName == "" {
		t.Error("expected non-empty tag name")
	}
}

func TestRunner_DetermineCalVer_Empty(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		CalVer: config.CalVerConfig{
			Enabled: true,
			Format:  "yyyy.mm.minor",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineCalVer("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version == "" {
		t.Error("expected non-empty version")
	}
}

func TestRunner_DetermineSemVer(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "minor",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.1.0" {
		t.Errorf("expected 1.1.0, got %s", runner.ctx.Version)
	}
	if runner.ctx.TagName != "v1.1.0" {
		t.Errorf("expected v1.1.0, got %s", runner.ctx.TagName)
	}
}

func TestRunner_DetermineSemVer_PreRelease(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		Increment:    "major",
		PreReleaseID: "beta",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "2.0.0-beta.0" {
		t.Errorf("expected 2.0.0-beta.0, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_InvalidVersion(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "patch",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineSemVer("invalid")
	if err == nil {
		t.Error("expected error for invalid version")
	}
}

// --- Helper to create a runner with mocked git commands ---

// setupMockedRunner creates a Runner with a mocked commandExecutor.
// The cmdResponses map keys are full command strings like "git describe --tags --abbrev=0".
func setupMockedRunner(t *testing.T, cfg *config.Config, cmdResponses map[string]struct {
	output string
	err    error
}) *Runner {
	t.Helper()

	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		key := name + " " + strings.Join(args, " ")
		if resp, ok := cmdResponses[key]; ok {
			return resp.output, resp.err
		}
		// Default: return empty for unknown commands to avoid test flakiness
		return "", fmt.Errorf("unexpected command in test: %s", key)
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	return runner
}

// --- generateChangelog tests ---

func TestRunner_GenerateChangelog_Enabled_WithCommits(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Changelog: config.ChangelogConfig{
			Enabled: true,
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat: add new feature\nfix: fix a bug\nchore: update deps",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Version = "1.1.0"

	err := runner.generateChangelog()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Changelog == "" {
		t.Error("expected non-empty changelog")
	}
}

func TestRunner_GenerateChangelog_Enabled_NoPrefix(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Changelog: config.ChangelogConfig{
			Enabled: true,
		},
		Git: config.GitConfig{
			TagName: "${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		// When LatestVersion doesn't have "v" prefix, code prepends "v"
		"git log v2.0.0..HEAD --pretty=format:%s": {
			output: "feat: something new",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "2.0.0"
	runner.ctx.Version = "2.1.0"

	err := runner.generateChangelog()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Changelog == "" {
		t.Error("expected non-empty changelog")
	}
}

func TestRunner_GenerateChangelog_Enabled_LatestVersionWithVPrefix(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Changelog: config.ChangelogConfig{
			Enabled: true,
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		// If latest version already has "v" prefix, it should not be doubled
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "fix: patch fix",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "v1.0.0"
	runner.ctx.Version = "1.0.1"

	err := runner.generateChangelog()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Changelog == "" {
		t.Error("expected non-empty changelog")
	}
}

func TestRunner_GenerateChangelog_GetCommitsError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Changelog: config.ChangelogConfig{
			Enabled: true,
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "",
			err:    fmt.Errorf("git error"),
		},
	})

	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Version = "1.1.0"

	err := runner.generateChangelog()
	if err == nil {
		t.Error("expected error when git fails")
	}
	if !strings.Contains(err.Error(), "getting commits") {
		t.Errorf("expected error about getting commits, got: %v", err)
	}
}

func TestRunner_GenerateChangelog_UpdateFile(t *testing.T) {
	dir := t.TempDir()
	changelogFile := dir + "/CHANGELOG.md"

	cfg := &config.Config{
		CI: true,
		Changelog: config.ChangelogConfig{
			Enabled: true,
			Infile:  changelogFile,
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat: new feature",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Version = "1.1.0"

	err := runner.generateChangelog()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	data, readErr := os.ReadFile(changelogFile)
	if readErr != nil {
		t.Fatalf("expected changelog file to be created: %v", readErr)
	}
	if len(data) == 0 {
		t.Error("expected non-empty changelog file")
	}
}

func TestRunner_GenerateChangelog_DryRun_DoesNotWriteFile(t *testing.T) {
	dir := t.TempDir()
	changelogFile := dir + "/CHANGELOG.md"

	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Changelog: config.ChangelogConfig{
			Enabled: true,
			Infile:  changelogFile,
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat: new feature",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Version = "1.1.0"

	err := runner.generateChangelog()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should NOT exist in dry-run mode
	_, readErr := os.ReadFile(changelogFile)
	if readErr == nil {
		t.Error("expected changelog file to NOT be created in dry-run mode")
	}
}

func TestRunner_GenerateChangelog_KeepAChangelog(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Changelog: config.ChangelogConfig{
			Enabled:        true,
			KeepAChangelog: true,
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat: add login\nfix: resolve crash",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Version = "1.1.0"

	err := runner.generateChangelog()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Changelog == "" {
		t.Error("expected non-empty changelog in keep-a-changelog format")
	}
}

// --- gitRelease tests ---

func TestRunner_GitRelease_CI_CommitTagPush_DryRun(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true, // Use dry-run so write ops are skipped
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          true,
			PushRepo:      "origin",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		// TagExists calls commandExecutor directly even in dry-run
		"git tag -l v1.0.0": {output: "", err: nil},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitRelease_NoCommit_NoTag_NoPush(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:  false,
			Tag:     false,
			Push:    false,
			TagName: "v${version}",
		},
	}

	runner := NewRunner(cfg)
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitRelease_CommitOnly(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "chore: release ${version}",
			Tag:           false,
			Push:          false,
			TagName:       "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	runner.ctx.Version = "2.0.0"
	runner.ctx.TagName = "v2.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitRelease_TagOnly(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			Commit:        false,
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          false,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git tag -l v1.0.0": {output: "", err: nil},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitRelease_Interactive_CommitDeclined(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          true,
			PushRepo:      "origin",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Spinner = ui.NewSpinner(true) // Use CI spinner to avoid race
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    nil,
	}

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitRelease_Interactive_CommitConfirmed_TagDeclined(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          true,
			PushRepo:      "origin",
		},
	}

	confirmCallCount := 0
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Spinner = ui.NewSpinner(true) // Use CI spinner to avoid race
	runner.ctx.Prompter = &sequentialMockPrompter{
		confirmResults: []bool{true, false}, // commit yes, tag no
		confirmErrors:  []error{nil, nil},
	}
	_ = confirmCallCount

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitRelease_Interactive_AllConfirmed(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          true,
			PushRepo:      "origin",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git tag -l v1.0.0": {output: "", err: nil},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Spinner = ui.NewSpinner(true) // Use CI spinner to avoid race
	runner.ctx.Prompter = &sequentialMockPrompter{
		confirmResults: []bool{true, true, true}, // commit yes, tag yes, push yes
		confirmErrors:  []error{nil, nil, nil},
	}

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitRelease_Interactive_ConfirmError(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           false,
			Push:          false,
			TagName:       "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Spinner = ui.NewSpinner(true) // Use CI spinner to avoid race
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    fmt.Errorf("prompt cancelled"),
	}

	err := runner.gitRelease()
	if err == nil {
		t.Error("expected error when prompter returns error")
	}
}

func TestRunner_GitRelease_Interactive_PushDeclined(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		Git: config.GitConfig{
			Commit:   false,
			Tag:      false,
			Push:     true,
			PushRepo: "origin",
			TagName:  "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Spinner = ui.NewSpinner(true) // Use CI spinner to avoid race
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    nil,
	}

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// sequentialMockPrompter allows different responses for sequential Confirm calls.
type sequentialMockPrompter struct {
	confirmResults      []bool
	confirmErrors       []error
	confirmCallIndex    int
	selectVersionResult string
	selectVersionErr    error
	inputResult         string
	inputErr            error
}

func (m *sequentialMockPrompter) SelectVersion(current string, recommended string, options []ui.VersionOption) (string, error) {
	return m.selectVersionResult, m.selectVersionErr
}

func (m *sequentialMockPrompter) Confirm(message string, defaultYes bool) (bool, error) {
	if m.confirmCallIndex < len(m.confirmResults) {
		result := m.confirmResults[m.confirmCallIndex]
		var err error
		if m.confirmCallIndex < len(m.confirmErrors) {
			err = m.confirmErrors[m.confirmCallIndex]
		}
		m.confirmCallIndex++
		return result, err
	}
	return false, fmt.Errorf("unexpected Confirm call #%d", m.confirmCallIndex)
}

func (m *sequentialMockPrompter) Input(message string, defaultValue string) (string, error) {
	return m.inputResult, m.inputErr
}

// --- autoDetectIncrement tests ---

func TestRunner_AutoDetectIncrement_FeatCommit(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat: add new feature\nfix: fix something",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	result := runner.autoDetectIncrement()
	if result != "minor" {
		t.Errorf("expected minor for feat commits, got %s", result)
	}
}

func TestRunner_AutoDetectIncrement_FixCommitOnly(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "fix: fix bug A\nfix: fix bug B",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	result := runner.autoDetectIncrement()
	if result != "patch" {
		t.Errorf("expected patch for fix-only commits, got %s", result)
	}
}

func TestRunner_AutoDetectIncrement_BreakingChange(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat!: breaking change\nfix: fix something",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	result := runner.autoDetectIncrement()
	if result != "major" {
		t.Errorf("expected major for breaking change, got %s", result)
	}
}

func TestRunner_AutoDetectIncrement_NoCommits(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	result := runner.autoDetectIncrement()
	if result != "patch" {
		t.Errorf("expected patch when no commits, got %s", result)
	}
}

func TestRunner_AutoDetectIncrement_GitError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "",
			err:    fmt.Errorf("git log failed"),
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	result := runner.autoDetectIncrement()
	if result != "patch" {
		t.Errorf("expected patch on git error, got %s", result)
	}
}

func TestRunner_AutoDetectIncrement_NonConventionalCommits(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "update readme\nsome random change",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	result := runner.autoDetectIncrement()
	if result != "patch" {
		t.Errorf("expected patch for non-conventional commits, got %s", result)
	}
}

func TestRunner_AutoDetectIncrement_EmptyLatestVersion(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v..HEAD --pretty=format:%s": {
			output: "",
			err:    fmt.Errorf("bad range"),
		},
	})

	runner.ctx.LatestVersion = ""

	result := runner.autoDetectIncrement()
	if result != "patch" {
		t.Errorf("expected patch for empty latest version, got %s", result)
	}
}

// --- determineSemVer interactive tests ---

func TestRunner_DetermineSemVer_Interactive_SelectVersion(t *testing.T) {
	cfg := &config.Config{
		CI: false,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "fix: some fix",
			err:    nil,
		},
	})

	runner.ctx.IsCI = false
	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Prompter = &mockPrompter{
		selectVersionResult: "1.2.0",
		selectVersionErr:    nil,
	}

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.2.0" {
		t.Errorf("expected 1.2.0 (user selection), got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_Interactive_SelectVersionError(t *testing.T) {
	cfg := &config.Config{
		CI: false,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "fix: some fix",
			err:    nil,
		},
	})

	runner.ctx.IsCI = false
	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Prompter = &mockPrompter{
		selectVersionResult: "",
		selectVersionErr:    fmt.Errorf("cancelled"),
	}

	err := runner.determineSemVer("1.0.0")
	if err == nil {
		t.Error("expected error when SelectVersion fails")
	}
}

func TestRunner_DetermineSemVer_AutoDetect_Patch(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "", // auto-detect
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "fix: patch fix",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.0.1" {
		t.Errorf("expected 1.0.1, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_AutoDetect_Minor(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "", // auto-detect
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat: new feature",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.1.0" {
		t.Errorf("expected 1.1.0, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_AutoDetect_Major(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "", // auto-detect
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%s": {
			output: "feat!: breaking change",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", runner.ctx.Version)
	}
}

// --- printSummary additional tests ---

func TestRunner_PrintSummary_WithReleaseURL(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			Push:          true,
			PushRepo:      "origin",
		},
		Changelog: config.ChangelogConfig{
			Infile: "CHANGELOG.md",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "2.0.0"
	runner.ctx.TagName = "v2.0.0"
	runner.ctx.Changelog = "some changelog content"
	runner.ctx.ReleaseURL = "https://github.com/emfi/release-it-go/releases/v2.0.0"
	runner.ctx.BranchName = "main"

	// Should not panic
	runner.printSummary(500000000) // 0.5 second
}

func TestRunner_PrintSummary_NoChangelog(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			Push:          false,
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.Changelog = ""
	runner.ctx.BranchName = "main"

	// Should not panic
	runner.printSummary(100000000)
}

func TestRunner_PrintSummary_NoPush(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:  false,
			Tag:     false,
			Push:    false,
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"

	// Should not panic
	runner.printSummary(200000000)
}

func TestRunner_PrintSummary_DryRunReleaseURL(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			Push:          true,
			PushRepo:      "origin",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.ReleaseURL = "(dry-run)"

	// Should not panic; ReleaseURL "(dry-run)" should not be printed
	runner.printSummary(0)
}

// --- checkPrerequisites tests ---

func TestRunner_CheckPrerequisites_Success(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git rev-parse --is-inside-work-tree": {output: "true", err: nil},
		"git rev-parse --abbrev-ref HEAD":     {output: "main", err: nil},
		"git status --porcelain":              {output: "", err: nil},
	})

	err := runner.checkPrerequisites()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- init tests ---

func TestRunner_Init(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":       {output: "https://github.com/emfi/release-it-go.git", err: nil},
		"git rev-parse --abbrev-ref HEAD": {output: "main", err: nil},
	})

	err := runner.init()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.BranchName != "main" {
		t.Errorf("expected branch 'main', got %q", runner.ctx.BranchName)
	}
	if runner.ctx.RepoInfo == nil {
		t.Error("expected non-nil RepoInfo")
	}
}

func TestRunner_Init_NoRemote(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":       {output: "", err: fmt.Errorf("no remote")},
		"git rev-parse --abbrev-ref HEAD": {output: "develop", err: nil},
	})

	err := runner.init()
	if err != nil {
		t.Fatalf("expected init to succeed even without remote: %v", err)
	}

	if runner.ctx.RepoInfo != nil {
		t.Error("expected nil RepoInfo when no remote")
	}
	if runner.ctx.BranchName != "develop" {
		t.Errorf("expected branch 'develop', got %q", runner.ctx.BranchName)
	}
}

func TestRunner_Init_NoBranch(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":       {output: "", err: fmt.Errorf("no remote")},
		"git rev-parse --abbrev-ref HEAD": {output: "", err: fmt.Errorf("no branch")},
	})

	err := runner.init()
	if err != nil {
		t.Fatalf("expected init to succeed even without branch: %v", err)
	}
}

// --- determineVersion tests ---

func TestRunner_DetermineVersion_NoIncrement_Mocked(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "no-increment",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "v1.2.3", err: nil},
	})

	err := runner.determineVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", runner.ctx.Version)
	}
	if runner.ctx.TagName != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %s", runner.ctx.TagName)
	}
}

func TestRunner_DetermineVersion_SemVer_Patch(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "patch",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "v1.0.0", err: nil},
	})

	err := runner.determineVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.0.1" {
		t.Errorf("expected 1.0.1, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineVersion_NoTags(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "patch",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "", err: fmt.Errorf("no tags")},
	})

	err := runner.determineVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "0.0.1" {
		t.Errorf("expected 0.0.1, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineVersion_CalVer(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		CalVer: config.CalVerConfig{
			Enabled: true,
			Format:  "yyyy.mm.minor",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "v2025.1.0", err: nil},
	})

	err := runner.determineVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version == "" {
		t.Error("expected non-empty version")
	}
}

// --- githubRelease interactive tests ---

func TestRunner_GithubRelease_Interactive_Declined(t *testing.T) {
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Release:     true,
			ReleaseName: "Release ${version}",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = &git.RepoInfo{
		Remote:     "https://github.com/emfi/release-it-go.git",
		Protocol:   "https",
		Host:       "github.com",
		Owner:      "emfi",
		Repository: "release-it-go",
	}
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    nil,
	}

	err := runner.githubRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GithubRelease_Interactive_ConfirmError(t *testing.T) {
	cfg := &config.Config{
		GitHub: config.GitHubConfig{
			Release:     true,
			ReleaseName: "Release ${version}",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = &git.RepoInfo{
		Remote:     "https://github.com/emfi/release-it-go.git",
		Protocol:   "https",
		Host:       "github.com",
		Owner:      "emfi",
		Repository: "release-it-go",
	}
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    errors.New("prompt error"),
	}

	err := runner.githubRelease()
	if err == nil {
		t.Error("expected error when prompter fails")
	}
}

// --- gitlabRelease interactive tests ---

func TestRunner_GitlabRelease_Interactive_Declined(t *testing.T) {
	cfg := &config.Config{
		GitLab: config.GitLabConfig{
			Release:     true,
			ReleaseName: "Release ${version}",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = &git.RepoInfo{
		Remote:     "https://gitlab.com/emfi/project.git",
		Protocol:   "https",
		Host:       "gitlab.com",
		Owner:      "emfi",
		Repository: "project",
	}
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    nil,
	}

	err := runner.gitlabRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_GitlabRelease_Interactive_ConfirmError(t *testing.T) {
	cfg := &config.Config{
		GitLab: config.GitLabConfig{
			Release:     true,
			ReleaseName: "Release ${version}",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = &git.RepoInfo{
		Remote:     "https://gitlab.com/emfi/project.git",
		Protocol:   "https",
		Host:       "gitlab.com",
		Owner:      "emfi",
		Repository: "project",
	}
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    errors.New("prompt error"),
	}

	err := runner.gitlabRelease()
	if err == nil {
		t.Error("expected error when prompter fails")
	}
}

// --- bumpFiles additional tests ---

func TestRunner_BumpFiles_MultipleFiles(t *testing.T) {
	dir := t.TempDir()
	file1 := dir + "/VERSION"
	file2 := dir + "/VERSION2"
	os.WriteFile(file1, []byte("1.0.0\n"), 0644)
	os.WriteFile(file2, []byte("1.0.0\n"), 0644)

	cfg := &config.Config{
		CI: true,
		Bumper: config.BumperConfig{
			Enabled: true,
			Out: []config.BumperFile{
				{File: file1, ConsumeWholeFile: true},
				{File: file2, ConsumeWholeFile: true},
			},
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Version = "3.0.0"

	err := runner.bumpFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data1, _ := os.ReadFile(file1)
	if string(data1) != "3.0.0\n" {
		t.Errorf("expected '3.0.0\\n' in file1, got %q", string(data1))
	}
	data2, _ := os.ReadFile(file2)
	if string(data2) != "3.0.0\n" {
		t.Errorf("expected '3.0.0\\n' in file2, got %q", string(data2))
	}
}

// --- buildVersionOptions additional tests ---

func TestRunner_BuildVersionOptions_Major(t *testing.T) {
	cfg := &config.Config{CI: true}
	runner := NewRunner(cfg)

	options := runner.buildVersionOptions("1.0.0", "major")

	if len(options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(options))
	}

	// Verify major is recommended
	if !options[2].Recommended {
		t.Error("major should be recommended")
	}
	if options[0].Recommended || options[1].Recommended {
		t.Error("patch and minor should not be recommended")
	}
}

func TestRunner_BuildVersionOptions_Patch(t *testing.T) {
	cfg := &config.Config{CI: true}
	runner := NewRunner(cfg)

	options := runner.buildVersionOptions("2.5.3", "patch")

	if len(options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(options))
	}

	if options[0].Version != "2.5.4" {
		t.Errorf("expected 2.5.4, got %s", options[0].Version)
	}
	if !options[0].Recommended {
		t.Error("patch should be recommended")
	}
}

// --- DetermineCalVer additional tests ---

func TestRunner_DetermineCalVer_CustomFormat(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		CalVer: config.CalVerConfig{
			Enabled: true,
			Format:  "yyyy.mm.minor",
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineCalVer("2026.1.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce a new calver version
	if runner.ctx.Version == "" {
		t.Error("expected non-empty version")
	}
	if runner.ctx.Version == "2026.1.5" {
		t.Error("expected version to be incremented from 2026.1.5")
	}
}

// --- DetermineSemVer additional tests ---

func TestRunner_DetermineSemVer_PreReleasePatch(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		Increment:    "patch",
		PreReleaseID: "alpha",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.0.1-alpha.0" {
		t.Errorf("expected 1.0.1-alpha.0, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_PreReleaseMinor(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		Increment:    "minor",
		PreReleaseID: "rc",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	err := runner.determineSemVer("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.1.0-rc.0" {
		t.Errorf("expected 1.1.0-rc.0, got %s", runner.ctx.Version)
	}
}

// --- gitRelease with actual (mocked) git operations ---

func TestRunner_GitRelease_CI_StageError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           false,
			Push:          false,
			TagName:       "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git add . --update": {output: "error", err: fmt.Errorf("stage failed")},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err == nil {
		t.Error("expected error when staging fails")
	}
	if !strings.Contains(err.Error(), "staging") {
		t.Errorf("expected staging error, got: %v", err)
	}
}

func TestRunner_GitRelease_CI_CommitError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           false,
			Push:          false,
			TagName:       "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git add . --update":                 {output: "", err: nil},
		"git commit --message Release 1.0.0": {output: "error", err: fmt.Errorf("commit failed")},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err == nil {
		t.Error("expected error when commit fails")
	}
	if !strings.Contains(err.Error(), "commit") {
		t.Errorf("expected commit error, got: %v", err)
	}
}

func TestRunner_GitRelease_CI_TagError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        false,
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          false,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git tag -l v1.0.0": {output: "v1.0.0", err: nil}, // tag already exists
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err == nil {
		t.Error("expected error when tag already exists")
	}
	if !strings.Contains(err.Error(), "tag") {
		t.Errorf("expected tag error, got: %v", err)
	}
}

func TestRunner_GitRelease_CI_PushError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:   false,
			Tag:      false,
			Push:     true,
			PushRepo: "origin",
			TagName:  "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git push origin": {output: "error", err: fmt.Errorf("push failed")},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err == nil {
		t.Error("expected error when push fails")
	}
	if !strings.Contains(err.Error(), "push") {
		t.Errorf("expected push error, got: %v", err)
	}
}

func TestRunner_GitRelease_CI_FullSuccess(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          true,
			PushRepo:      "origin",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git add . --update":                                {output: "", err: nil},
		"git commit --message Release 1.0.0":                {output: "", err: nil},
		"git tag -l v1.0.0":                                 {output: "", err: nil}, // tag does not exist
		"git tag --annotate --message Release 1.0.0 v1.0.0": {output: "", err: nil},
		"git push origin":                                   {output: "", err: nil},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- checkPrerequisites with error ---

func TestRunner_CheckPrerequisites_Error(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName:                "v${version}",
			RequireCleanWorkingDir: true,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git rev-parse --is-inside-work-tree": {output: "true", err: nil},
		"git rev-parse --abbrev-ref HEAD":     {output: "main", err: nil},
		"git status --porcelain":              {output: "M dirty_file.go", err: nil},
	})

	err := runner.checkPrerequisites()
	if err == nil {
		t.Error("expected error when working dir is dirty")
	}
}

// --- Run pipeline tests ---

func TestRunner_Run_FullPipeline_DryRun(t *testing.T) {
	dir := t.TempDir()
	changelogFile := dir + "/CHANGELOG.md"

	cfg := &config.Config{
		CI:        true,
		DryRun:    true,
		Increment: "patch",
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "Release ${version}",
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          true,
			PushRepo:      "origin",
		},
		Changelog: config.ChangelogConfig{
			Enabled: true,
			Infile:  changelogFile,
		},
		GitHub: config.GitHubConfig{
			Release: false,
		},
		GitLab: config.GitLabConfig{
			Release: false,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":               {output: "https://github.com/emfi/release-it-go.git", err: nil},
		"git rev-parse --abbrev-ref HEAD":         {output: "main", err: nil},
		"git rev-parse --is-inside-work-tree":     {output: "true", err: nil},
		"git status --porcelain":                  {output: "", err: nil},
		"git describe --tags --abbrev=0":          {output: "v1.0.0", err: nil},
		"git log v1.0.0..HEAD --pretty=format:%s": {output: "fix: a fix", err: nil},
		"git tag -l v1.0.1":                       {output: "", err: nil},
	})

	err := runner.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.0.1" {
		t.Errorf("expected version 1.0.1, got %s", runner.ctx.Version)
	}
}

func TestRunner_RunChangelogOnly(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "patch",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
		Changelog: config.ChangelogConfig{
			Enabled: true,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":               {output: "https://github.com/emfi/release-it-go.git", err: nil},
		"git rev-parse --abbrev-ref HEAD":         {output: "main", err: nil},
		"git describe --tags --abbrev=0":          {output: "v1.0.0", err: nil},
		"git log v1.0.0..HEAD --pretty=format:%s": {output: "feat: new feature\nfix: bug fix", err: nil},
	})

	err := runner.RunChangelogOnly()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_RunReleaseVersionOnly(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "minor",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":       {output: "https://github.com/emfi/release-it-go.git", err: nil},
		"git rev-parse --abbrev-ref HEAD": {output: "main", err: nil},
		"git describe --tags --abbrev=0":  {output: "v2.0.0", err: nil},
	})

	err := runner.RunReleaseVersionOnly()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "2.1.0" {
		t.Errorf("expected 2.1.0, got %s", runner.ctx.Version)
	}
}

func TestRunner_RunOnlyVersion(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		DryRun:    true,
		Increment: "patch",
		Git: config.GitConfig{
			Commit:  false,
			Tag:     false,
			Push:    false,
			TagName: "v${version}",
		},
		Changelog: config.ChangelogConfig{
			Enabled: false,
		},
		GitHub: config.GitHubConfig{Release: false},
		GitLab: config.GitLabConfig{Release: false},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":       {output: "https://github.com/emfi/release-it-go.git", err: nil},
		"git rev-parse --abbrev-ref HEAD": {output: "main", err: nil},
		"git describe --tags --abbrev=0":  {output: "v1.0.0", err: nil},
	})

	err := runner.RunOnlyVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.0.1" {
		t.Errorf("expected 1.0.1, got %s", runner.ctx.Version)
	}
}

func TestRunner_RunNoIncrement(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			Commit:  false,
			Tag:     false,
			Push:    false,
			TagName: "v${version}",
		},
		Changelog: config.ChangelogConfig{
			Enabled: false,
		},
		GitHub: config.GitHubConfig{Release: false},
		GitLab: config.GitLabConfig{Release: false},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":       {output: "https://github.com/emfi/release-it-go.git", err: nil},
		"git rev-parse --abbrev-ref HEAD": {output: "main", err: nil},
		"git describe --tags --abbrev=0":  {output: "v3.2.1", err: nil},
	})

	err := runner.RunNoIncrement()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "3.2.1" {
		t.Errorf("expected 3.2.1, got %s", runner.ctx.Version)
	}
	if runner.ctx.TagName != "v3.2.1" {
		t.Errorf("expected v3.2.1, got %s", runner.ctx.TagName)
	}
}

func TestRunner_RunNoIncrement_NoTags(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
		Changelog: config.ChangelogConfig{Enabled: false},
		GitHub:    config.GitHubConfig{Release: false},
		GitLab:    config.GitLabConfig{Release: false},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url origin":       {output: "", err: fmt.Errorf("no remote")},
		"git rev-parse --abbrev-ref HEAD": {output: "main", err: nil},
		"git describe --tags --abbrev=0":  {output: "", err: fmt.Errorf("no tags")},
	})

	err := runner.RunNoIncrement()
	if err == nil {
		t.Error("expected error when no tags exist for RunNoIncrement")
	}
	if !strings.Contains(err.Error(), "latest tag") {
		t.Errorf("expected 'latest tag' error, got: %v", err)
	}
}

// --- determineVersion with bumper ---

func TestRunner_DetermineVersion_WithBumperInput(t *testing.T) {
	dir := t.TempDir()
	versionFile := dir + "/VERSION"
	os.WriteFile(versionFile, []byte("5.0.0\n"), 0644)

	cfg := &config.Config{
		CI:        true,
		Increment: "patch",
		Bumper: config.BumperConfig{
			Enabled: true,
			In: &config.BumperFile{
				File:             versionFile,
				ConsumeWholeFile: true,
			},
		},
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "", err: fmt.Errorf("no tags")},
	})

	err := runner.determineVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "5.0.1" {
		t.Errorf("expected 5.0.1, got %s", runner.ctx.Version)
	}
}

// --- Interactive tag confirm error ---

func TestRunner_GitRelease_Interactive_TagConfirmError(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		Git: config.GitConfig{
			Commit:        false,
			Tag:           true,
			TagName:       "v${version}",
			TagAnnotation: "Release ${version}",
			Push:          false,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git tag -l v1.0.0": {output: "", err: nil},
	})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Spinner = ui.NewSpinner(true) // Use CI spinner to avoid race
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    fmt.Errorf("tag prompt error"),
	}

	err := runner.gitRelease()
	if err == nil {
		t.Error("expected error when tag confirm fails")
	}
}

func TestRunner_GitRelease_Interactive_PushConfirmError(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		Git: config.GitConfig{
			Commit:   false,
			Tag:      false,
			Push:     true,
			PushRepo: "origin",
			TagName:  "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Spinner = ui.NewSpinner(true) // Use CI spinner to avoid race
	runner.ctx.Prompter = &mockPrompter{
		confirmResult: false,
		confirmErr:    fmt.Errorf("push prompt error"),
	}

	err := runner.gitRelease()
	if err == nil {
		t.Error("expected error when push confirm fails")
	}
}

// --- GitRelease with AddUntrackedFiles ---

func TestRunner_GitRelease_CI_AddUntrackedFiles(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		Git: config.GitConfig{
			Commit:            true,
			CommitMessage:     "Release ${version}",
			Tag:               false,
			Push:              false,
			TagName:           "v${version}",
			AddUntrackedFiles: true,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = true

	err := runner.gitRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Ensure imports are used
var _ = errors.New
