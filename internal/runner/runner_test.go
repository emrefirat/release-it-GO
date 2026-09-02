package runner

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"release-it-go/internal/changelog"
	"release-it-go/internal/config"
	"release-it-go/internal/git"
	applog "release-it-go/internal/log"
	"release-it-go/internal/ui"
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

func TestLatestVersionToTag(t *testing.T) {
	tests := []struct {
		name            string
		latestVersion   string
		tagNameTemplate string
		expected        string
	}{
		{"empty version", "", "v${version}", ""},
		{"zero version", "0.0.0", "v${version}", ""},
		{"normal with v template", "1.0.0", "v${version}", "v1.0.0"},
		{"v-prefixed version with v template", "v1.0.0", "v${version}", "v1.0.0"},
		{"normal with bare template", "1.0.0", "${version}", "1.0.0"},
		{"empty template", "1.0.0", "", "1.0.0"},
		{"custom prefix template", "2.0.0", "release-${version}", "release-2.0.0"},
		{"v-prefixed version with custom template", "v2.0.0", "release-${version}", "release-2.0.0"},
		{"pre-release version", "1.0.0-beta.1", "v${version}", "v1.0.0-beta.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := latestVersionToTag(tt.latestVersion, tt.tagNameTemplate)
			if result != tt.expected {
				t.Errorf("latestVersionToTag(%q, %q) = %q, want %q",
					tt.latestVersion, tt.tagNameTemplate, result, tt.expected)
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

func (m *mockPrompter) Select(question string, options []string, defaultIndex int) (int, error) {
	return defaultIndex, nil
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

// testGitHubRepo returns a RepoInfo pointing at github.com.
func testGitHubRepo() *git.RepoInfo {
	return &git.RepoInfo{
		Host:       "github.com",
		Owner:      "testowner",
		Repository: "testrepo",
		Protocol:   "https",
	}
}

// testGitLabRepo returns a RepoInfo pointing at gitlab.com.
func testGitLabRepo() *git.RepoInfo {
	return &git.RepoInfo{
		Host:       "gitlab.com",
		Owner:      "testowner",
		Repository: "testrepo",
		Protocol:   "https",
	}
}

func TestRunner_GithubRelease_DryRun_SetsReleaseURL(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "fake-token")
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		GitHub: config.GitHubConfig{
			Release:     true,
			TokenRef:    "GITHUB_TOKEN",
			ReleaseName: "Release ${version}",
			MakeLatest:  true,
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitHubRepo()
	runner.ctx.Version = "1.2.3"
	runner.ctx.TagName = "v1.2.3"
	runner.ctx.Changelog = "- some change\n"

	if err := runner.githubRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.ReleaseURL != "(dry-run)" {
		t.Errorf("expected ReleaseURL=(dry-run), got %q", runner.ctx.ReleaseURL)
	}
}

func TestRunner_GithubRelease_NonCI_PromptDecline_Skips(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "fake-token")
	cfg := &config.Config{
		DryRun: true,
		GitHub: config.GitHubConfig{
			Release:     true,
			TokenRef:    "GITHUB_TOKEN",
			ReleaseName: "Release ${version}",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitHubRepo()
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	// NewRunner sets IsCI via ui.IsCI() which returns true when stdin isn't a
	// TTY (as in `go test`). Force interactive path for this test.
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{confirmResult: false}

	if err := runner.githubRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.ReleaseURL != "" {
		t.Errorf("expected empty ReleaseURL on decline, got %q", runner.ctx.ReleaseURL)
	}
}

func TestRunner_GithubRelease_NonCI_PromptAccept_ProceedsInDryRun(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "fake-token")
	cfg := &config.Config{
		DryRun: true,
		GitHub: config.GitHubConfig{
			Release:     true,
			TokenRef:    "GITHUB_TOKEN",
			ReleaseName: "Release ${version}",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitHubRepo()
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{confirmResult: true}

	if err := runner.githubRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.ReleaseURL != "(dry-run)" {
		t.Errorf("expected ReleaseURL=(dry-run) after accept, got %q", runner.ctx.ReleaseURL)
	}
}

func TestRunner_GithubRelease_MissingToken_ReturnsError(t *testing.T) {
	// Ensure the env var is empty for this test.
	t.Setenv("GITHUB_TOKEN", "")
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		GitHub: config.GitHubConfig{
			Release:  true,
			TokenRef: "GITHUB_TOKEN",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitHubRepo()

	err := runner.githubRelease()
	if err == nil {
		t.Fatal("expected error when GITHUB_TOKEN is not set")
	}
	if !strings.Contains(err.Error(), "GitHub client") {
		t.Errorf("expected wrap with 'GitHub client', got %v", err)
	}
}

func TestRunner_GitlabRelease_DryRun_SetsReleaseURL(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "fake-token")
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		GitLab: config.GitLabConfig{
			Release:     true,
			TokenRef:    "GITLAB_TOKEN",
			TokenHeader: "Private-Token",
			ReleaseName: "Release ${version}",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitLabRepo()
	runner.ctx.Version = "1.2.3"
	runner.ctx.TagName = "v1.2.3"
	runner.ctx.Changelog = "- fix x\n"

	if err := runner.gitlabRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.ReleaseURL != "(dry-run)" {
		t.Errorf("expected ReleaseURL=(dry-run), got %q", runner.ctx.ReleaseURL)
	}
}

func TestRunner_GitlabRelease_NonCI_PromptDecline_Skips(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "fake-token")
	cfg := &config.Config{
		DryRun: true,
		GitLab: config.GitLabConfig{
			Release:     true,
			TokenRef:    "GITLAB_TOKEN",
			TokenHeader: "Private-Token",
			ReleaseName: "Release ${version}",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitLabRepo()
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{confirmResult: false}

	if err := runner.gitlabRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.ReleaseURL != "" {
		t.Errorf("expected empty ReleaseURL on decline, got %q", runner.ctx.ReleaseURL)
	}
}

func TestRunner_GitlabRelease_NonCI_PromptAccept_ProceedsInDryRun(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "fake-token")
	cfg := &config.Config{
		DryRun: true,
		GitLab: config.GitLabConfig{
			Release:     true,
			TokenRef:    "GITLAB_TOKEN",
			TokenHeader: "Private-Token",
			ReleaseName: "Release ${version}",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitLabRepo()
	runner.ctx.Version = "1.0.0"
	runner.ctx.TagName = "v1.0.0"
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{confirmResult: true}

	if err := runner.gitlabRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.ReleaseURL != "(dry-run)" {
		t.Errorf("expected ReleaseURL=(dry-run) after accept, got %q", runner.ctx.ReleaseURL)
	}
}

func TestRunner_GitlabRelease_MissingToken_ReturnsError(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "")
	cfg := &config.Config{
		CI:     true,
		DryRun: true,
		GitLab: config.GitLabConfig{
			Release:     true,
			TokenRef:    "GITLAB_TOKEN",
			TokenHeader: "Private-Token",
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.RepoInfo = testGitLabRepo()

	err := runner.gitlabRelease()
	if err == nil {
		t.Fatal("expected error when GITLAB_TOKEN is not set")
	}
	if !strings.Contains(err.Error(), "GitLab client") {
		t.Errorf("expected wrap with 'GitLab client', got %v", err)
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
	_ = os.WriteFile(file, []byte("1.0.0\n"), 0644)

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
	_ = os.WriteFile(file, []byte("1.0.0\n"), 0644)

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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0001\x1ffeat: add new feature\x1e\nabc0002\x1ffix: fix a bug\x1e\nabc0003\x1fchore: update deps\x1e",
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
		// TagName="${version}" so tag has no prefix: 2.0.0
		"git log 2.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0004\x1ffeat: something new\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0005\x1ffix: patch fix\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0006\x1ffeat: new feature\x1e",
			err:    nil,
		},
		"git add " + changelogFile: {
			output: "",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0007\x1ffeat: new feature\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0008\x1ffeat: add login\x1e\nabc0009\x1ffix: resolve crash\x1e",
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
		// commit yes, tag no, push no — declining the tag no longer cancels
		// the push prompt (each operation is confirmed independently)
		confirmResults: []bool{true, false, false},
		confirmErrors:  []error{nil, nil, nil},
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

func (m *sequentialMockPrompter) Select(question string, options []string, defaultIndex int) (int, error) {
	return defaultIndex, nil
}

// --- autoDetectIncrement tests ---

func TestRunner_AutoDetectIncrement_BreakingChangeFooter(t *testing.T) {
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: change API\n\nBREAKING CHANGE: the old endpoints were removed\x1e",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	result := runner.autoDetectIncrement()
	if result != "major" {
		t.Errorf("expected major for BREAKING CHANGE footer (spec-canonical form), got %s", result)
	}
}

func TestRunner_AutoDetectIncrement_TagFormatTransition_FallsBackToRawTag(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "${version}", // new format — but repo's latest tag is v-prefixed
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log 1.1.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "",
			err:    fmt.Errorf("fatal: ambiguous argument '1.1.0..HEAD': unknown revision"),
		},
		"git describe --tags --abbrev=0": {
			output: "v1.1.0",
			err:    nil,
		},
		"git log v1.1.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: new capability\x1e",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.1.0"

	result := runner.autoDetectIncrement()
	if result != "minor" {
		t.Errorf("expected minor via raw-tag fallback (was silently 'patch' before), got %s", result)
	}
}

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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc000e\x1ffeat: add new feature\x1e\nabc000f\x1ffix: fix something\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0010\x1ffix: fix bug A\x1e\nabc0011\x1ffix: fix bug B\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0012\x1ffeat!: breaking change\x1e\nabc0013\x1ffix: fix something\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0014\x1fupdate readme\x1e\nabc0015\x1fsome random change\x1e",
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
		"git log v..HEAD --pretty=format:%h%x1f%B%x1e": {
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0016\x1ffix: some fix\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0017\x1ffix: some fix\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0018\x1ffix: patch fix\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc0019\x1ffeat: new feature\x1e",
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
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc001a\x1ffeat!: breaking change\x1e",
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
		"git remote get-url origin":       {output: "https://github.com/testowner/testrepo.git", err: nil},
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
		Remote:     "https://github.com/testowner/testrepo.git",
		Protocol:   "https",
		Host:       "github.com",
		Owner:      "testowner",
		Repository: "testrepo",
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
		Remote:     "https://github.com/testowner/testrepo.git",
		Protocol:   "https",
		Host:       "github.com",
		Owner:      "testowner",
		Repository: "testrepo",
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
	_ = os.WriteFile(file1, []byte("1.0.0\n"), 0644)
	_ = os.WriteFile(file2, []byte("1.0.0\n"), 0644)

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

func TestRunner_DetermineSemVer_PreReleaseIncrement(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		Increment:    "patch",
		PreReleaseID: "deneme2",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	// Current version is already pre-release with same ID → should increment number
	err := runner.determineSemVer("1.6.0-deneme2.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.6.0-deneme2.1" {
		t.Errorf("expected 1.6.0-deneme2.1, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_PreReleaseDifferentID(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		Increment:    "patch",
		PreReleaseID: "rc",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	// Current version has different pre-release ID → should start new series
	err := runner.determineSemVer("1.6.0-beta.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "1.6.0-rc.0" {
		t.Errorf("expected 1.6.0-rc.0, got %s", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_PreReleaseSecondIncrement(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		Increment:    "minor",
		PreReleaseID: "beta",
		Git: config.GitConfig{
			TagName: "v${version}",
		},
	}
	runner := NewRunner(cfg)

	// Already beta.5 → should become beta.6
	err := runner.determineSemVer("2.0.0-beta.5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if runner.ctx.Version != "2.0.0-beta.6" {
		t.Errorf("expected 2.0.0-beta.6, got %s", runner.ctx.Version)
	}
}

// --- gitRelease with actual (mocked) git operations ---

func TestRunner_GitRelease_CI_StageError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:            true,
			CommitMessage:     "Release ${version}",
			Tag:               false,
			Push:              false,
			TagName:           "v${version}",
			AddUntrackedFiles: true, // whole-tree staging path
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git add .": {output: "error", err: fmt.Errorf("stage failed")},
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
		"git remote get-url origin":                         {output: "https://github.com/testowner/testrepo.git", err: nil},
		"git rev-parse --abbrev-ref HEAD":                   {output: "main", err: nil},
		"git rev-parse --is-inside-work-tree":               {output: "true", err: nil},
		"git status --porcelain":                            {output: "", err: nil},
		"git config user.name":                              {output: "Test User", err: nil},
		"git config user.email":                             {output: "test@example.com", err: nil},
		"git describe --tags --abbrev=0":                    {output: "v1.0.0", err: nil},
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {output: "abc001b\x1ffix: a fix\x1e", err: nil},
		"git tag -l v1.0.1":                                 {output: "", err: nil},
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
		"git remote get-url origin":                         {output: "https://github.com/testowner/testrepo.git", err: nil},
		"git rev-parse --abbrev-ref HEAD":                   {output: "main", err: nil},
		"git describe --tags --abbrev=0":                    {output: "v1.0.0", err: nil},
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {output: "abc001c\x1ffeat: new feature\x1e\nabc001d\x1ffix: bug fix\x1e", err: nil},
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
		"git remote get-url origin":       {output: "https://github.com/testowner/testrepo.git", err: nil},
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
		"git rev-parse --is-inside-work-tree": {output: "true", err: nil},
		"git remote get-url origin":           {output: "https://github.com/testowner/testrepo.git", err: nil},
		"git rev-parse --abbrev-ref HEAD":     {output: "main", err: nil},
		"git describe --tags --abbrev=0":      {output: "v1.0.0", err: nil},
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
		"git rev-parse --is-inside-work-tree": {output: "true", err: nil},
		"git remote get-url origin":           {output: "https://github.com/testowner/testrepo.git", err: nil},
		"git rev-parse --abbrev-ref HEAD":     {output: "main", err: nil},
		"git describe --tags --abbrev=0":      {output: "v3.2.1", err: nil},
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
		"git rev-parse --is-inside-work-tree": {output: "true", err: nil},
		"git remote get-url origin":           {output: "", err: fmt.Errorf("no remote")},
		"git rev-parse --abbrev-ref HEAD":     {output: "main", err: nil},
		"git describe --tags --abbrev=0":      {output: "", err: fmt.Errorf("no tags")},
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
	_ = os.WriteFile(versionFile, []byte("5.0.0\n"), 0644)

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

// --- checkTokens tests ---

func TestRunner_CheckTokens_NoRelease(t *testing.T) {
	cfg := &config.Config{
		CI:     true,
		GitHub: config.GitHubConfig{Release: false},
		GitLab: config.GitLabConfig{Release: false},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err != nil {
		t.Errorf("expected no error when releases are disabled, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitHubMissingToken(t *testing.T) {
	// Ensure GITHUB_TOKEN is unset for this test
	t.Setenv("GITHUB_TOKEN", "")

	cfg := &config.Config{
		CI: true,
		GitHub: config.GitHubConfig{
			Release: true,
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err == nil {
		t.Error("expected error when GITHUB_TOKEN is missing")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("expected error about GITHUB_TOKEN, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitHubTokenSet(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "ghp_test123")

	cfg := &config.Config{
		CI: true,
		GitHub: config.GitHubConfig{
			Release: true,
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err != nil {
		t.Errorf("expected no error when GITHUB_TOKEN is set, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitHubCustomTokenRef(t *testing.T) {
	t.Setenv("MY_GH_TOKEN", "")

	cfg := &config.Config{
		CI: true,
		GitHub: config.GitHubConfig{
			Release:  true,
			TokenRef: "MY_GH_TOKEN",
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err == nil {
		t.Error("expected error when custom token ref is missing")
	}
	if !strings.Contains(err.Error(), "MY_GH_TOKEN") {
		t.Errorf("expected error about MY_GH_TOKEN, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitHubSkipChecks(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	cfg := &config.Config{
		CI: true,
		GitHub: config.GitHubConfig{
			Release:    true,
			SkipChecks: true,
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err != nil {
		t.Errorf("expected no error when skipChecks is true, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitLabMissingToken(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "")

	cfg := &config.Config{
		CI: true,
		GitLab: config.GitLabConfig{
			Release: true,
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err == nil {
		t.Error("expected error when GITLAB_TOKEN is missing")
	}
	if !strings.Contains(err.Error(), "GITLAB_TOKEN") {
		t.Errorf("expected error about GITLAB_TOKEN, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitLabTokenSet(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "glpat-test123")

	cfg := &config.Config{
		CI: true,
		GitLab: config.GitLabConfig{
			Release: true,
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err != nil {
		t.Errorf("expected no error when GITLAB_TOKEN is set, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitLabCustomTokenRef(t *testing.T) {
	t.Setenv("MY_GL_TOKEN", "")

	cfg := &config.Config{
		CI: true,
		GitLab: config.GitLabConfig{
			Release:  true,
			TokenRef: "MY_GL_TOKEN",
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err == nil {
		t.Error("expected error when custom GitLab token ref is missing")
	}
	if !strings.Contains(err.Error(), "MY_GL_TOKEN") {
		t.Errorf("expected error about MY_GL_TOKEN, got: %v", err)
	}
}

func TestRunner_CheckTokens_GitLabSkipChecks(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "")

	cfg := &config.Config{
		CI: true,
		GitLab: config.GitLabConfig{
			Release:    true,
			SkipChecks: true,
		},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err != nil {
		t.Errorf("expected no error when skipChecks is true, got: %v", err)
	}
}

func TestRunner_CheckTokens_BothEnabled_BothSet(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "ghp_test")
	t.Setenv("GITLAB_TOKEN", "glpat-test")

	cfg := &config.Config{
		CI:     true,
		GitHub: config.GitHubConfig{Release: true},
		GitLab: config.GitLabConfig{Release: true},
	}
	runner := NewRunner(cfg)

	err := runner.checkTokens()
	if err != nil {
		t.Errorf("expected no error when both tokens are set, got: %v", err)
	}
}

// --- sendNotification tests ---

func TestRunner_SendNotification_Disabled(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Notification: config.NotificationConfig{
			Enabled: false,
		},
	}
	runner := NewRunner(cfg)

	err := runner.sendNotification()
	if err != nil {
		t.Errorf("expected no error when notification is disabled, got: %v", err)
	}
}

func TestRunner_SendNotification_EmptyWebhooks(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Notification: config.NotificationConfig{
			Enabled:  true,
			Webhooks: []config.WebhookConfig{},
		},
	}
	runner := NewRunner(cfg)

	err := runner.sendNotification()
	if err != nil {
		t.Errorf("expected no error with empty webhooks, got: %v", err)
	}
}

func TestRunner_SendNotification_NonFatal(t *testing.T) {
	// Missing env var should cause notification to fail, but runner should not return error
	cfg := &config.Config{
		CI: true,
		Notification: config.NotificationConfig{
			Enabled: true,
			Webhooks: []config.WebhookConfig{
				{Type: "slack", URLRef: "NONEXISTENT_WEBHOOK_URL_XYZ"},
			},
		},
	}
	runner := NewRunner(cfg)
	runner.ctx.Vars = map[string]string{"version": "1.0.0"}

	err := runner.sendNotification()
	if err != nil {
		t.Errorf("notification errors should be non-fatal, got: %v", err)
	}
}

// --- checkPrerequisites: no commits path ---

func TestRunner_CheckPrerequisites_NoCommits(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName:        "v${version}",
			RequireCommits: true,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git rev-parse --is-inside-work-tree": {output: "true", err: nil},
		"git rev-parse --abbrev-ref HEAD":     {output: "main", err: nil},
		"git status --porcelain":              {output: "", err: nil},
		"git config user.name":                {output: "Test", err: nil},
		"git config user.email":               {output: "test@example.com", err: nil},
		"git describe --tags --abbrev=0":      {output: "v1.0.0", err: nil},
		// empty log output = no commits; checkCommits returns git.ErrNoCommits
		"git log v1.0.0..HEAD --oneline": {output: "", err: nil},
	})

	err := runner.checkPrerequisites()
	if err != nil {
		t.Fatalf("expected no error for no-commits (soft), got: %v", err)
	}
	if !runner.ctx.noCommits {
		t.Error("expected noCommits to be set to true")
	}
}

func TestRunner_CheckPrerequisites_TokenCheckFails(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName: "v${version}",
		},
		GitHub: config.GitHubConfig{
			Release: true,
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

	// Ensure GITHUB_TOKEN is not set
	_ = os.Unsetenv("GITHUB_TOKEN")

	err := runner.checkPrerequisites()
	if err == nil {
		t.Fatal("expected error when GitHub release enabled but no token")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("expected GITHUB_TOKEN error, got: %v", err)
	}
}

// --- formatLintError tests ---

func TestFormatLintError_SingleFailure(t *testing.T) {
	failed := []changelog.LintResult{
		{Hash: "abc1234", Subject: "bad commit message", Reason: "missing type prefix"},
	}

	err := formatLintError(failed, 5, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "Commit lint failed") {
		t.Errorf("expected 'Commit lint failed' in error, got: %s", msg)
	}
	if !strings.Contains(msg, "abc1234") {
		t.Errorf("expected hash in error, got: %s", msg)
	}
	if !strings.Contains(msg, "bad commit message") {
		t.Errorf("expected subject in error, got: %s", msg)
	}
	if !strings.Contains(msg, "missing type prefix") {
		t.Errorf("expected reason in error, got: %s", msg)
	}
	if !strings.Contains(msg, "1 of 5") {
		t.Errorf("expected '1 of 5' in error, got: %s", msg)
	}
	if !strings.Contains(msg, "--ignore-commit-lint") {
		t.Errorf("expected bypass hint in error, got: %s", msg)
	}
}

func TestFormatLintError_MultipleFailures(t *testing.T) {
	failed := []changelog.LintResult{
		{Hash: "abc1234", Subject: "bad one", Reason: "missing type"},
		{Hash: "def5678", Subject: "bad two", Reason: "empty subject"},
	}

	err := formatLintError(failed, 10, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "2 of 10") {
		t.Errorf("expected '2 of 10' in error, got: %s", msg)
	}
	if !strings.Contains(msg, "abc1234") || !strings.Contains(msg, "def5678") {
		t.Errorf("expected both hashes in error, got: %s", msg)
	}
}

// --- checkCommitLint tests ---

func TestRunner_CheckCommitLint_Disabled(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			RequireConventionalCommits: false,
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	err := runner.checkCommitLint()
	if err != nil {
		t.Fatalf("expected no error when disabled, got: %v", err)
	}
}

func TestRunner_CheckCommitLint_NoCommits(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			RequireConventionalCommits: true,
			TagName:                    "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {
			output: "v1.0.0",
			err:    nil,
		},
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {
			output: "",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	err := runner.checkCommitLint()
	if err != nil {
		t.Fatalf("expected no error for no commits, got: %v", err)
	}
}

func TestRunner_CheckCommitLint_AllConventional(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			RequireConventionalCommits: true,
			TagName:                    "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {
			output: "abc1234||feat: add new feature\ndef5678||fix: fix a bug",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	err := runner.checkCommitLint()
	if err != nil {
		t.Fatalf("expected no error for conventional commits, got: %v", err)
	}
}

func TestRunner_CheckCommitLint_FailsOnNonConventional(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			RequireConventionalCommits: true,
			TagName:                    "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {
			output: "abc1234||feat: add feature\ndef5678||bad commit message",
			err:    nil,
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	err := runner.checkCommitLint()
	if err == nil {
		t.Fatal("expected error for non-conventional commit")
	}
	if !strings.Contains(err.Error(), "Commit lint failed") {
		t.Errorf("expected lint error message, got: %v", err)
	}
}

func TestRunner_CheckCommitLint_GitError(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			RequireConventionalCommits: true,
			TagName:                    "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {
			output: "",
			err:    fmt.Errorf("git error"),
		},
	})

	runner.ctx.LatestVersion = "1.0.0"

	err := runner.checkCommitLint()
	if err == nil {
		t.Fatal("expected error on git failure")
	}
	if !strings.Contains(err.Error(), "getting commits for lint") {
		t.Errorf("expected wrapped git error, got: %v", err)
	}
}

func TestRunner_CheckCommitLint_NoLatestVersion_UsesGetLatestTag(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			RequireConventionalCommits: true,
			TagName:                    "v${version}",
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {
			output: "v0.5.0",
			err:    nil,
		},
		"git log v0.5.0..HEAD --pretty=format:%h||%s": {
			output: "abc1234||feat: something",
			err:    nil,
		},
	})

	// LatestVersion is empty, so it should call GetLatestTag
	runner.ctx.LatestVersion = ""

	err := runner.checkCommitLint()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// --- RunCheckCommits tests ---

func TestRunner_RunCheckCommits_NoCommits(t *testing.T) {
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
		"git describe --tags --abbrev=0": {
			output: "v1.0.0",
			err:    nil,
		},
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {
			output: "",
			err:    nil,
		},
		"git remote get-url origin": {
			output: "https://github.com/user/repo.git",
			err:    nil,
		},
		"git rev-parse --abbrev-ref HEAD": {
			output: "main",
			err:    nil,
		},
	})

	err := runner.RunCheckCommits()
	if err != nil {
		t.Fatalf("expected no error for no commits, got: %v", err)
	}
}

func TestRunner_RunCheckCommits_AllPass(t *testing.T) {
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
		"git describe --tags --abbrev=0": {
			output: "v1.0.0",
			err:    nil,
		},
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {
			output: "abc1234||feat: new feature\ndef5678||fix: bug fix",
			err:    nil,
		},
		"git remote get-url origin": {
			output: "https://github.com/user/repo.git",
			err:    nil,
		},
		"git rev-parse --abbrev-ref HEAD": {
			output: "main",
			err:    nil,
		},
	})

	err := runner.RunCheckCommits()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunner_RunCheckCommits_WithFailures(t *testing.T) {
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
		"git describe --tags --abbrev=0": {
			output: "v1.0.0",
			err:    nil,
		},
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {
			output: "abc1234||feat: good commit\ndef5678||this is bad",
			err:    nil,
		},
		"git remote get-url origin": {
			output: "https://github.com/user/repo.git",
			err:    nil,
		},
		"git rev-parse --abbrev-ref HEAD": {
			output: "main",
			err:    nil,
		},
	})

	err := runner.RunCheckCommits()
	if err == nil {
		t.Fatal("expected error for non-conventional commits")
	}
	if !strings.Contains(err.Error(), "Commit lint failed") {
		t.Errorf("expected lint error, got: %v", err)
	}
}

func TestRunner_RunCheckCommits_NoTags(t *testing.T) {
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
		"git describe --tags --abbrev=0": {
			output: "",
			err:    fmt.Errorf("no tags"),
		},
		"git log --pretty=format:%h||%s": {
			output: "abc1234||feat: initial commit",
			err:    nil,
		},
		"git remote get-url origin": {
			output: "https://github.com/user/repo.git",
			err:    nil,
		},
		"git rev-parse --abbrev-ref HEAD": {
			output: "main",
			err:    nil,
		},
	})

	err := runner.RunCheckCommits()
	if err != nil {
		t.Fatalf("expected no error when no tags, got: %v", err)
	}
}

// --- resolvePreReleaseBaseTag tests ---

func TestRunner_ResolvePreReleaseBaseTag_ContinueSeries(t *testing.T) {
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
		"git tag -l --merged HEAD --sort=-v:refname": {
			output: "v1.2.3-beta.1\nv1.2.3-beta.0\nv1.2.0\nv1.1.0",
			err:    nil,
		},
	})

	tag, err := runner.resolvePreReleaseBaseTag("beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.2.3-beta.1" {
		t.Errorf("expected v1.2.3-beta.1 (continue series), got %q", tag)
	}
}

func TestRunner_ResolvePreReleaseBaseTag_NewSeries(t *testing.T) {
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
		// Pre-release base (1.0.0) < stable (2.0.0) → new series
		"git tag -l --merged HEAD --sort=-v:refname": {
			output: "v2.0.0\nv1.0.1-beta.0\nv1.0.0",
			err:    nil,
		},
	})

	tag, err := runner.resolvePreReleaseBaseTag("beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v2.0.0" {
		t.Errorf("expected v2.0.0 (new series), got %q", tag)
	}
}

func TestRunner_ResolvePreReleaseBaseTag_NoPreReleaseTag(t *testing.T) {
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
		"git tag -l --merged HEAD --sort=-v:refname": {
			output: "v1.0.0\nv0.9.0",
			err:    nil,
		},
	})

	tag, err := runner.resolvePreReleaseBaseTag("alpha")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.0.0" {
		t.Errorf("expected v1.0.0 (stable, no alpha tags), got %q", tag)
	}
}

func TestRunner_ResolvePreReleaseBaseTag_NoStableTag(t *testing.T) {
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
		"git tag -l --merged HEAD --sort=-v:refname": {
			output: "v0.1.0-rc.2\nv0.1.0-rc.1",
			err:    nil,
		},
	})

	tag, err := runner.resolvePreReleaseBaseTag("rc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v0.1.0-rc.2" {
		t.Errorf("expected v0.1.0-rc.2 (no stable, continue pre-release), got %q", tag)
	}
}

func TestRunner_ResolvePreReleaseBaseTag_NoTags(t *testing.T) {
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
		"git tag -l --merged HEAD --sort=-v:refname": {
			output: "",
			err:    nil,
		},
	})

	tag, err := runner.resolvePreReleaseBaseTag("beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "" {
		t.Errorf("expected empty (no tags at all), got %q", tag)
	}
}

func TestRunner_ResolvePreReleaseBaseTag_GitError(t *testing.T) {
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
		"git tag -l --merged HEAD --sort=-v:refname": {
			output: "",
			err:    fmt.Errorf("git error"),
		},
	})

	_, err := runner.resolvePreReleaseBaseTag("beta")
	if err == nil {
		t.Error("expected error when git fails")
	}
}

// --- filterContributors tests ---

func TestFilterContributors_NoIgnored(t *testing.T) {
	contributors := []string{"Alice", "Bob", "Charlie"}
	result := filterContributors(contributors, nil)
	if len(result) != 3 {
		t.Errorf("expected 3, got %d", len(result))
	}
}

func TestFilterContributors_SomeIgnored(t *testing.T) {
	contributors := []string{"Alice", "Jenkins", "Bob", "GitLab Bot", "Charlie"}
	ignored := []string{"Jenkins", "GitLab Bot"}
	result := filterContributors(contributors, ignored)
	if len(result) != 3 {
		t.Errorf("expected 3, got %d: %v", len(result), result)
	}
	for _, name := range result {
		if name == "Jenkins" || name == "GitLab Bot" {
			t.Errorf("expected %q to be filtered out", name)
		}
	}
}

func TestFilterContributors_AllIgnored(t *testing.T) {
	contributors := []string{"Jenkins", "Bot"}
	ignored := []string{"Jenkins", "Bot"}
	result := filterContributors(contributors, ignored)
	if len(result) != 0 {
		t.Errorf("expected 0, got %d: %v", len(result), result)
	}
}

func TestFilterContributors_EmptyContributors(t *testing.T) {
	result := filterContributors(nil, []string{"Jenkins"})
	if len(result) != 0 {
		t.Errorf("expected 0, got %d", len(result))
	}
}

// Ensure imports are used
var _ = errors.New

func TestRunner_DetermineSemVer_ExplicitVersion(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "2.5.0", // -i 1.5.0 / positional "2.5.0": explicit target version
		Git:       config.GitConfig{TagName: "v${version}"},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	if err := runner.determineSemVer("1.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.Version != "2.5.0" {
		t.Errorf("Version = %q, want 2.5.0", runner.ctx.Version)
	}
	if runner.ctx.TagName != "v2.5.0" {
		t.Errorf("TagName = %q, want v2.5.0", runner.ctx.TagName)
	}
}

func TestRunner_DetermineSemVer_ExplicitVersionWithVPrefix(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "v3.0.0",
		Git:       config.GitConfig{TagName: "v${version}"},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	if err := runner.determineSemVer("1.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.Version != "3.0.0" {
		t.Errorf("Version = %q, want 3.0.0 (v prefix stripped)", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_ExplicitIncrement_DoesNotPrompt(t *testing.T) {
	cfg := &config.Config{
		Increment: "minor", // explicitly chosen — npm never prompts here
		Git:       config.GitConfig{TagName: "v${version}"},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		// auto-detect would also say "minor" — the old prompt condition
		// (increment == autoDetect) fired exactly in this coincidence
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: something\x1e",
			err:    nil,
		},
	})
	runner.ctx.IsCI = false // force the interactive branch
	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Prompter = &mockPrompter{
		selectVersionErr: fmt.Errorf("prompt must not be shown for an explicit increment"),
	}

	if err := runner.determineSemVer("1.0.0"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.Version != "1.1.0" {
		t.Errorf("Version = %q, want 1.1.0", runner.ctx.Version)
	}
}

func TestRunner_DetermineVersion_InfersVPrefixFromLatestTag(t *testing.T) {
	cfg := &config.Config{
		CI:  true,
		Git: config.GitConfig{TagName: "${version}"}, // shipped default, not user-set
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "v1.2.2", err: nil},
		// Inference must make the pipeline query the REAL v-prefixed tag —
		// previously "1.2.2..HEAD" failed and auto-increment fell to patch.
		"git log v1.2.2..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: new capability\x1e",
			err:    nil,
		},
	})

	if err := runner.determineVersion(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.Version != "1.3.0" {
		t.Errorf("Version = %q, want 1.3.0 (feat → minor)", runner.ctx.Version)
	}
	if runner.ctx.TagName != "v1.3.0" {
		t.Errorf("TagName = %q, want v1.3.0 (v prefix inferred from history)", runner.ctx.TagName)
	}
}

func TestRunner_DetermineVersion_ExplicitTagName_NoInference(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			TagName:         "${version}",
			TagNameExplicit: true, // user wrote tagName in the config file
		},
	}

	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "v1.2.2", err: nil},
		"git log 1.2.2..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "",
			err:    fmt.Errorf("unknown revision"),
		},
		// raw-tag fallback path
		"git log v1.2.2..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: new capability\x1e",
			err:    nil,
		},
	})

	if err := runner.determineVersion(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.TagName != "1.3.0" {
		t.Errorf("TagName = %q, want 1.3.0 (explicit template must be respected)", runner.ctx.TagName)
	}
}

func TestRunner_GitRelease_DeclineCommit_StillTagsAndPushes(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitConfig{
			Commit: true, Tag: true, Push: true,
			TagName: "v${version}", TagAnnotation: "Release ${version}",
			CommitMessage: "chore: release v${version}",
		},
	}

	var calls []string
	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		if strings.HasPrefix(call, "git diff --cached") {
			return "staged.txt", nil // there ARE staged changes
		}
		return "", nil
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	runner.ctx.IsCI = false
	runner.ctx.Version = "1.1.0"
	runner.ctx.TagName = "v1.1.0"
	// npm asks commit/tag/push independently: declining the commit must not
	// silently cancel the tag and push prompts.
	runner.ctx.Prompter = &sequentialMockPrompter{confirmResults: []bool{false, true, true}}

	if err := runner.gitRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined := strings.Join(calls, "\n")
	if strings.Contains(joined, "git commit") {
		t.Error("declined commit must not run git commit")
	}
	if !strings.Contains(joined, "--annotate") {
		t.Errorf("tag must still be created after a declined commit; calls:\n%s", joined)
	}
	if !strings.Contains(joined, "git push") {
		t.Errorf("push must still run after a declined commit; calls:\n%s", joined)
	}
}

func TestRunner_GitRelease_DeclineTag_StillPushes(t *testing.T) {
	cfg := &config.Config{
		Git: config.GitConfig{
			Commit: false, Tag: true, Push: true,
			TagName: "v${version}", TagAnnotation: "Release ${version}",
		},
	}

	var calls []string
	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		calls = append(calls, name+" "+strings.Join(args, " "))
		return "", nil
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	runner.ctx.IsCI = false
	runner.ctx.Version = "1.1.0"
	runner.ctx.TagName = "v1.1.0"
	runner.ctx.Prompter = &sequentialMockPrompter{confirmResults: []bool{false, true}} // tag: no, push: yes

	if err := runner.gitRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined := strings.Join(calls, "\n")
	if strings.Contains(joined, "--annotate") {
		t.Error("declined tag must not be created")
	}
	if !strings.Contains(joined, "git push") {
		t.Errorf("push must still run after a declined tag; calls:\n%s", joined)
	}
}

func TestRunner_BuildVersionOptions_PreRelease(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		PreReleaseID: "beta",
		Git:          config.GitConfig{TagName: "v${version}"},
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	options := runner.buildVersionOptions("1.2.3", "preminor")

	if len(options) != 3 {
		t.Fatalf("expected 3 pre-release options, got %d: %+v", len(options), options)
	}
	want := []string{"1.2.4-beta.0", "1.3.0-beta.0", "2.0.0-beta.0"}
	for i, w := range want {
		if options[i].Version != w {
			t.Errorf("options[%d].Version = %q, want %q", i, options[i].Version, w)
		}
	}
	for _, o := range options {
		if !strings.Contains(o.Version, "-beta.") {
			t.Errorf("pre-release menu must never offer a stable version, got %q", o.Version)
		}
	}
}

func TestRunner_BuildVersionOptions_PreRelease_ContinueSeries(t *testing.T) {
	cfg := &config.Config{
		CI:           true,
		PreReleaseID: "beta",
		Git:          config.GitConfig{TagName: "v${version}"},
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})

	options := runner.buildVersionOptions("1.3.0-beta.1", "prerelease")

	if len(options) == 0 {
		t.Fatal("expected options")
	}
	if options[0].Version != "1.3.0-beta.2" {
		t.Errorf("first option = %q, want 1.3.0-beta.2 (continue current series)", options[0].Version)
	}
	if !options[0].Recommended {
		t.Error("continue-series option should carry the recommendation")
	}
}

// optionCapturingPrompter records the options offered to the user.
type optionCapturingPrompter struct {
	capturedOptions []ui.VersionOption
	returnVersion   string
}

func (m *optionCapturingPrompter) SelectVersion(current string, recommended string, options []ui.VersionOption) (string, error) {
	m.capturedOptions = options
	if m.returnVersion != "" {
		return m.returnVersion, nil
	}
	return recommended, nil
}
func (m *optionCapturingPrompter) Confirm(message string, defaultYes bool) (bool, error) {
	return true, nil
}
func (m *optionCapturingPrompter) Input(message string, defaultValue string) (string, error) {
	return "", nil
}
func (m *optionCapturingPrompter) Select(question string, options []string, defaultIndex int) (int, error) {
	return defaultIndex, nil
}

func TestRunner_DetermineSemVer_PreRelease_PromptKeepsIdentifier(t *testing.T) {
	cfg := &config.Config{
		PreReleaseID: "beta",
		Git:          config.GitConfig{TagName: "v${version}"},
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.2.3..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: something\x1e",
			err:    nil,
		},
	})
	runner.ctx.IsCI = false
	runner.ctx.LatestVersion = "1.2.3"
	prompter := &optionCapturingPrompter{returnVersion: "1.3.0-beta.0"}
	runner.ctx.Prompter = prompter

	if err := runner.determineSemVer("1.2.3"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prompter.capturedOptions) == 0 {
		t.Fatal("expected the prompt to be shown with options")
	}
	for _, o := range prompter.capturedOptions {
		if !strings.Contains(o.Version, "-beta.") {
			t.Errorf("prompt offered a stable version %q — the pre-release identifier was dropped", o.Version)
		}
	}
	if runner.ctx.Version != "1.3.0-beta.0" {
		t.Errorf("Version = %q, want 1.3.0-beta.0", runner.ctx.Version)
	}
}

func TestRunner_RenderReleaseTemplate_AllVars(t *testing.T) {
	cfg := &config.Config{CI: true, Git: config.GitConfig{TagName: "v${version}"}}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{})
	runner.ctx.Version = "1.2.0"
	runner.ctx.Vars = map[string]string{
		"latestVersion":   "1.1.0",
		"branchName":      "main",
		"repo.repository": "my-repo",
	}

	got := runner.renderReleaseTemplate("release ${version} (was ${latestVersion}) on ${branchName} of ${repo.repository}, keep ${unknown}")
	want := "release 1.2.0 (was 1.1.0) on main of my-repo, keep ${unknown}"
	if got != want {
		t.Errorf("renderReleaseTemplate:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestRunner_GitRelease_CommitMessageTemplateVars(t *testing.T) {
	cfg := &config.Config{
		CI: true,
		Git: config.GitConfig{
			Commit:        true,
			CommitMessage: "chore: release ${version} on ${branchName}",
		},
	}

	var commitMsg string
	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		if strings.HasPrefix(call, "git diff --cached") {
			return "staged.txt", nil
		}
		if len(args) > 2 && args[0] == "commit" && args[1] == "--message" {
			commitMsg = args[2]
		}
		return "", nil
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	runner.ctx.Version = "1.1.0"
	runner.ctx.BranchName = "main"
	runner.ctx.UpdateVars()

	if err := runner.gitRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if commitMsg != "chore: release 1.1.0 on main" {
		t.Errorf("commit message = %q, want branchName rendered (npm template var parity)", commitMsg)
	}
}

func TestRunner_GitRelease_StagesTrackedChanges_ByDefault(t *testing.T) {
	// npm release-it stages tracked modifications (git add . --update) so
	// files changed by hooks (dist/, lockfiles) land in the release commit.
	// Untracked files are only swept in with addUntrackedFiles.
	cfg := &config.Config{
		CI:  true,
		Git: config.GitConfig{Commit: true, CommitMessage: "chore: release ${version}"},
	}

	var calls []string
	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		if strings.HasPrefix(call, "git diff --cached") {
			return "dist/app.js", nil
		}
		return "", nil
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	runner.ctx.Version = "1.1.0"

	if err := runner.gitRelease(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined := strings.Join(calls, "\n")
	if !strings.Contains(joined, "git add . --update") {
		t.Errorf("tracked modifications must be staged (git add . --update); calls:\n%s", joined)
	}
	if strings.Contains(joined+"\n", "git add .\n") {
		t.Errorf("untracked files must not be swept in without addUntrackedFiles; calls:\n%s", joined)
	}
}
func TestRunner_GitRelease_PushError_SuggestsRecovery(t *testing.T) {
	cfg := &config.Config{
		CI:  true,
		Git: config.GitConfig{Push: true},
	}

	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		if len(args) > 0 && args[0] == "push" {
			return "! [rejected] main -> main (fetch first)", fmt.Errorf("exit status 1")
		}
		return "", nil
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	runner.ctx.Version = "1.1.0"
	runner.ctx.TagName = "v1.1.0"

	err := runner.gitRelease()
	if err == nil {
		t.Fatal("expected push error")
	}
	if !strings.Contains(err.Error(), "--no-increment") {
		t.Errorf("push failure must point at the --no-increment recovery flow, got: %v", err)
	}
}

func TestRunner_RunCheckCommits_Verbose_ListsEachFailureOnce(t *testing.T) {
	cfg := &config.Config{
		CI:      true,
		Verbose: 1,
		Git:     config.GitConfig{TagName: "v${version}"},
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0":              {output: "v1.0.0", err: nil},
		"git log v1.0.0..HEAD --pretty=format:%h||%s": {output: "abc1234||feat: good\ndef5678||bad message here", err: nil},
	})
	var buf strings.Builder
	runner.ctx.Logger = applog.NewLoggerWithWriter(1, false, &buf)

	err := runner.RunCheckCommits()
	if err == nil {
		t.Fatal("expected lint failure")
	}

	combined := buf.String() + err.Error()
	if n := strings.Count(combined, "bad message here"); n != 1 {
		t.Errorf("failure listed %d times, want exactly once (verbose ✓/✗ list and error block duplicated it):\n%s", n, combined)
	}
	if !strings.Contains(err.Error(), "1 of 2 commits") {
		t.Errorf("summary should still report the count, got: %v", err)
	}
}

func TestRunner_CommitsSinceLatestRelease_NoTags_BumperVersion_UsesAllCommits(t *testing.T) {
	// bumper.in supplied "5.0.0" but the repo has no tags: the rendered tag
	// v5.0.0 does not exist, GetLatestTag fails too, and the changelog step
	// used to abort with "ambiguous argument". Fall back to all commits.
	cfg := &config.Config{
		CI:  true,
		Git: config.GitConfig{TagName: "v${version}"},
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v5.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "", err: fmt.Errorf("fatal: ambiguous argument 'v5.0.0..HEAD'"),
		},
		"git describe --tags --abbrev=0": {output: "", err: fmt.Errorf("fatal: No names found")},
		"git log --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: first\x1e", err: nil,
		},
	})
	runner.ctx.LatestVersion = "5.0.0"

	commits, err := runner.commitsSinceLatestRelease()
	if err != nil {
		t.Fatalf("expected fallback to all commits, got error: %v", err)
	}
	if len(commits) != 1 || commits[0].Message != "feat: first" {
		t.Errorf("commits = %+v, want the single feat commit", commits)
	}
}

// --- entry points must enforce the same safety checks as Run() ---

func TestRunner_RunOnlyVersion_DirtyTree_FailsBeforeAnyGitWrite(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Increment: "patch",
		Git: config.GitConfig{
			Commit: true, Tag: true,
			TagName:                "v${version}",
			RequireCleanWorkingDir: true,
		},
	}
	var calls []string
	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		switch {
		case strings.HasPrefix(call, "git rev-parse --is-inside-work-tree"):
			return "true", nil
		case strings.HasPrefix(call, "git status --porcelain"):
			return " M dirty.txt", nil
		}
		return "", nil
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	err := runner.RunOnlyVersion()
	if err == nil {
		t.Fatal("expected the dirty working tree to abort --only-version")
	}
	if !strings.Contains(err.Error(), "not clean") {
		t.Errorf("error should name the failed prerequisite, got: %v", err)
	}
	for _, c := range calls {
		if strings.HasPrefix(c, "git commit") || strings.HasPrefix(c, "git tag --annotate") {
			t.Errorf("no git write may happen after a failed prerequisite, saw: %s", c)
		}
	}
}

func TestRunner_RunNoIncrement_MissingToken_FailsBeforeAnyGitWrite(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	cfg := &config.Config{
		CI:     true,
		Git:    config.GitConfig{Commit: true, Tag: true, TagName: "v${version}"},
		GitHub: config.GitHubConfig{Release: true, TokenRef: "GITHUB_TOKEN"},
	}
	var calls []string
	restore := git.SetCommandExecutorForTest(func(name string, args ...string) (string, error) {
		call := name + " " + strings.Join(args, " ")
		calls = append(calls, call)
		switch {
		case strings.HasPrefix(call, "git rev-parse --is-inside-work-tree"):
			return "true", nil
		case strings.HasPrefix(call, "git describe"):
			return "v1.0.0", nil
		case strings.HasPrefix(call, "git config user."):
			return "Test User", nil // identity check passes; the token check is the one under test
		}
		return "", nil
	})
	t.Cleanup(restore)

	runner := NewRunner(cfg)
	err := runner.RunNoIncrement()
	if err == nil {
		t.Fatal("expected the missing token to abort --no-increment before any git write")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("error should name the missing token, got: %v", err)
	}
	for _, c := range calls {
		if strings.HasPrefix(c, "git commit") || strings.HasPrefix(c, "git tag --annotate") {
			t.Errorf("no git write may happen after a failed prerequisite, saw: %s", c)
		}
	}
}

func TestRunner_RunChangelogOnly_NeverPrompts(t *testing.T) {
	// Scripting mode: VERSION=$(release-it-go --changelog) must not block on
	// an interactive version menu even in a TTY.
	cfg := &config.Config{Git: config.GitConfig{TagName: "v${version}"}}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git describe --tags --abbrev=0": {output: "v1.0.0", err: nil},
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {
			output: "abc1234\x1ffeat: thing\x1e", err: nil,
		},
	})
	runner.ctx.IsCI = false
	runner.ctx.Prompter = &mockPrompter{selectVersionErr: fmt.Errorf("prompt must not be shown in --changelog mode")}

	if err := runner.RunChangelogOnly(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.Version != "1.1.0" {
		t.Errorf("Version = %q, want auto-detected 1.1.0", runner.ctx.Version)
	}
}

func TestRunner_DetermineSemVer_PreIncrementWithPreRelease(t *testing.T) {
	tests := []struct {
		name, increment, current, want string
	}{
		{"preminor keyword", "preminor", "1.0.0", "1.1.0-beta.0"},
		{"prepatch keyword", "prepatch", "1.0.0", "1.0.1-beta.0"},
		{"premajor keyword", "premajor", "1.0.0", "2.0.0-beta.0"},
		{"prerelease continues series", "prerelease", "1.1.0-beta.0", "1.1.0-beta.1"},
		{"plain minor gets pre prefix", "minor", "1.0.0", "1.1.0-beta.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				CI:           true,
				Increment:    tt.increment,
				PreReleaseID: "beta",
				Git:          config.GitConfig{TagName: "v${version}"},
			}
			runner := setupMockedRunner(t, cfg, map[string]struct {
				output string
				err    error
			}{})
			if err := runner.determineSemVer(tt.current); err != nil {
				t.Fatalf("unexpected error (these are the keywords the menu itself offers): %v", err)
			}
			if runner.ctx.Version != tt.want {
				t.Errorf("Version = %q, want %q", runner.ctx.Version, tt.want)
			}
		})
	}
}

func TestRunner_DetermineSemVer_ExplicitVersionNotGreater_Errors(t *testing.T) {
	for _, explicit := range []string{"0.5.0", "1.0.0"} {
		cfg := &config.Config{
			CI:        true,
			Increment: explicit,
			Git:       config.GitConfig{TagName: "v${version}"},
		}
		runner := setupMockedRunner(t, cfg, map[string]struct {
			output string
			err    error
		}{})
		err := runner.determineSemVer("1.0.0")
		if err == nil {
			t.Errorf("explicit %s on top of 1.0.0 must be rejected (npm requires a greater version)", explicit)
			continue
		}
		if !strings.Contains(err.Error(), "greater") {
			t.Errorf("error should explain the ordering rule, got: %v", err)
		}
	}
}

func TestRunner_CheckPrerequisites_ReleaseEnabledWithoutRepoInfo_Errors(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "tok")
	cfg := &config.Config{
		CI:     true,
		Git:    config.GitConfig{TagName: "v${version}"},
		GitHub: config.GitHubConfig{Release: true, TokenRef: "GITHUB_TOKEN"},
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git rev-parse --is-inside-work-tree": {output: "true", err: nil},
	})
	runner.ctx.RepoInfo = nil // origin missing or unparseable

	err := runner.checkPrerequisites()
	if err == nil {
		t.Fatal("a GitHub release cannot be created without repository info; this used to be skipped silently with a green 'Done'")
	}
	if !strings.Contains(err.Error(), "remote") {
		t.Errorf("error should point at the remote, got: %v", err)
	}
}

func TestRunner_Init_UsesPushRepoForRepoInfo(t *testing.T) {
	cfg := &config.Config{
		CI:  true,
		Git: config.GitConfig{TagName: "v${version}", PushRepo: "upstream"},
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git remote get-url upstream":     {output: "https://github.com/o/r.git", err: nil},
		"git rev-parse --abbrev-ref HEAD": {output: "main", err: nil},
	})

	if err := runner.init(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.ctx.RepoInfo == nil || runner.ctx.RepoInfo.Owner != "o" {
		t.Errorf("RepoInfo should come from git.pushRepo (upstream), got %+v", runner.ctx.RepoInfo)
	}
}

func TestRunner_AutoDetectIncrement_GitError_IsLogged(t *testing.T) {
	cfg := &config.Config{CI: true, Git: config.GitConfig{TagName: "v${version}"}}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {output: "", err: fmt.Errorf("fatal: bad revision")},
		"git describe --tags --abbrev=0":                    {output: "", err: fmt.Errorf("no tags")},
		"git log --pretty=format:%h%x1f%B%x1e":              {output: "", err: fmt.Errorf("fatal: bad revision")},
	})
	var buf strings.Builder
	runner.ctx.Logger = applog.NewLoggerWithWriter(0, false, &buf)
	runner.ctx.LatestVersion = "1.0.0"

	if got := runner.autoDetectIncrement(); got != "patch" {
		t.Errorf("fallback increment = %q, want patch", got)
	}
	if !strings.Contains(strings.ToLower(buf.String()), "bad revision") {
		t.Errorf("a git failure behind the 'patch' fallback must be visible, log was:\n%s", buf.String())
	}
}

func TestRunner_GenerateChangelog_CompareURLUsesTagNames(t *testing.T) {
	cfg := &config.Config{
		CI:        true,
		Git:       config.GitConfig{TagName: "v${version}"},
		Changelog: config.ChangelogConfig{Enabled: true, AddVersionUrl: true}, // no Infile → no file write
	}
	runner := setupMockedRunner(t, cfg, map[string]struct {
		output string
		err    error
	}{
		"git log v1.0.0..HEAD --pretty=format:%h%x1f%B%x1e": {output: "abc1234\x1ffeat: thing\x1e", err: nil},
	})
	runner.ctx.RepoInfo = &git.RepoInfo{Host: "github.com", Owner: "o", Repository: "r"}
	runner.ctx.LatestVersion = "1.0.0"
	runner.ctx.Version = "1.1.0"
	runner.ctx.TagName = "v1.1.0"

	if err := runner.generateChangelog(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(runner.ctx.Changelog, "compare/v1.0.0...v1.1.0") {
		t.Errorf("compare link must use tag names on a v-tagged repo, got:\n%s", runner.ctx.Changelog)
	}
}
