package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fmt"
	"time"

	"release-it-go/internal/config"
	"release-it-go/internal/runner"
)

// newTestConfig creates a config suitable for integration tests.
// It disables push, upstream check, and CI-related features.
// Uses v-prefixed tags to match the convention used by changelog (v${version}).
func newTestConfig(dir string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.CI = true
	cfg.Git.Push = false
	cfg.Git.RequireUpstream = false
	cfg.Git.RequireCleanWorkingDir = false
	cfg.Git.AddUntrackedFiles = true
	cfg.Git.TagName = "v${version}"
	cfg.GitHub.Release = false
	cfg.GitLab.Release = false
	cfg.Changelog.Infile = filepath.Join(dir, "CHANGELOG.md")
	return cfg
}

func TestIntegration_FullReleasePipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: add user authentication",
		"fix: resolve login timeout",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify tag was created (v-prefixed)
	assertTagExists(t, dir, "v1.1.0")

	// Verify commit message
	msg := getLatestCommitMsg(t, dir)
	if !strings.Contains(msg, "release v1.1.0") {
		t.Errorf("expected commit message to contain 'release v1.1.0', got %q", msg)
	}

	// Verify CHANGELOG.md was created/updated
	assertChangelogContains(t, dir, "1.1.0")
	assertChangelogContains(t, dir, "user authentication")
}

func TestIntegration_PatchBump(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v2.3.1")
	createCommits(t, dir, []string{
		"fix: resolve crash on startup",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "patch"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	assertTagExists(t, dir, "v2.3.2")
}

func TestIntegration_MajorBump(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat!: redesign API endpoints",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "major"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	assertTagExists(t, dir, "v2.0.0")
}

func TestIntegration_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: add new feature",
	})

	cfg := newTestConfig(dir)
	cfg.DryRun = true
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() in dry-run failed: %v", err)
	}

	// In dry-run mode, no tag should be created
	assertTagNotExists(t, dir, "v1.1.0")
}

func TestIntegration_NoExistingTags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createCommits(t, dir, []string{
		"feat: initial feature",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"
	// Disable changelog and commit since there's no previous tag to compute commits from
	cfg.Changelog.Enabled = false
	cfg.Git.Commit = false

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() with no tags failed: %v", err)
	}

	// Should start from 0.0.0 and bump to 0.1.0
	assertTagExists(t, dir, "v0.1.0")
}

func TestIntegration_ChangelogOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: add search functionality",
		"fix: handle empty query",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.RunChangelogOnly()
	if err != nil {
		t.Fatalf("RunChangelogOnly() failed: %v", err)
	}

	// No tag should be created in changelog-only mode
	assertTagNotExists(t, dir, "v1.1.0")
}

func TestIntegration_ReleaseVersionOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v3.2.1")
	createCommits(t, dir, []string{
		"feat: add export",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.RunReleaseVersionOnly()
	if err != nil {
		t.Fatalf("RunReleaseVersionOnly() failed: %v", err)
	}

	// No tag should be created in release-version mode
	assertTagNotExists(t, dir, "v3.3.0")
}

func TestIntegration_DisableCommitAndTag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"fix: minor fix",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "patch"
	cfg.Git.Commit = false
	cfg.Git.Tag = false

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Tag should NOT be created when git.tag is false
	assertTagNotExists(t, dir, "v1.0.1")
}

func TestIntegration_ConventionalCommitAutoIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: add new dashboard",
		"feat: add analytics page",
	})

	cfg := newTestConfig(dir)
	// Don't set Increment - let conventional commits auto-detect minor

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Two feat commits should result in a minor bump: 1.0.0 -> 1.1.0
	assertTagExists(t, dir, "v1.1.0")
}

func TestIntegration_BreakingChangeAutoMajor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat!: remove legacy API",
	})

	cfg := newTestConfig(dir)
	// Don't set Increment - let conventional commits auto-detect major

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Breaking change should result in major bump: 1.0.0 -> 2.0.0
	assertTagExists(t, dir, "v2.0.0")
}

func TestIntegration_BumperFileUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	// Create package.json with version
	pkgJSON := filepath.Join(dir, "package.json")
	writeFile(t, pkgJSON, `{"name": "test-app", "version": "1.0.0"}`)

	// Create VERSION file
	versionFile := filepath.Join(dir, "VERSION")
	writeFile(t, versionFile, "1.0.0\n")

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: add feature",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"
	cfg.Bumper.Enabled = true
	cfg.Bumper.Out = []config.BumperFile{
		{File: pkgJSON, Path: "version"},
		{File: versionFile, ConsumeWholeFile: true},
	}

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify package.json was updated
	data, _ := os.ReadFile(pkgJSON)
	if !strings.Contains(string(data), `"version": "1.1.0"`) {
		t.Errorf("package.json version not updated, got: %s", string(data))
	}

	// Verify VERSION file was updated
	data, _ = os.ReadFile(versionFile)
	if strings.TrimSpace(string(data)) != "1.1.0" {
		t.Errorf("VERSION file not updated, got: %q", string(data))
	}
}

func TestIntegration_KeepAChangelog(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: add user profiles",
		"fix: correct email validation",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"
	cfg.Changelog.KeepAChangelog = true

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Keep-a-changelog format uses "Added" and "Fixed" sections
	assertChangelogContains(t, dir, "## [1.1.0]")
	assertChangelogContains(t, dir, "Added")
}

func TestIntegration_HookExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	markerFile := filepath.Join(dir, "hook_executed.txt")

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: new feature",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "patch"
	cfg.Hooks.AfterInit = []string{"echo hook_ran > " + markerFile}

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Verify hook was executed
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("expected after:init hook to create marker file")
	}
}

func TestIntegration_HookFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{
		"feat: new feature",
	})

	cfg := newTestConfig(dir)
	cfg.Increment = "patch"
	cfg.Hooks.BeforeInit = []string{"exit 1"}

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err == nil {
		t.Error("expected Run() to fail when before:init hook fails")
	}
	if !strings.Contains(err.Error(), "hook") {
		t.Errorf("expected hook error, got: %v", err)
	}
}

func TestIntegration_ConfigFromJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	configContent := `{
		"git": {
			"commit": true,
			"tag": true,
			"push": false,
			"requireUpstream": false,
			"requireCleanWorkingDir": false,
			"addUntrackedFiles": true,
			"tagName": "v${version}",
			"commitMessage": "chore: release v${version}"
		},
		"ci": true,
		"increment": "patch"
	}`
	configPath := filepath.Join(dir, ".release-it.json")
	writeFile(t, configPath, configContent)

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"fix: bug fix"})

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	cfg.Changelog.Infile = filepath.Join(dir, "CHANGELOG.md")

	r := runner.NewRunner(cfg)
	err = r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	assertTagExists(t, dir, "v1.0.1")
}

func TestIntegration_ConfigFromYAML(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	configContent := `git:
  commit: true
  tag: true
  push: false
  requireUpstream: false
  requireCleanWorkingDir: false
  addUntrackedFiles: true
  tagName: "v${version}"
  commitMessage: "chore: release v${version}"
ci: true
increment: minor
`
	configPath := filepath.Join(dir, ".release-it.yaml")
	writeFile(t, configPath, configContent)

	initGitRepo(t, dir)
	createTag(t, dir, "v2.0.0")
	createCommits(t, dir, []string{"feat: new feature"})

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	cfg.Changelog.Infile = filepath.Join(dir, "CHANGELOG.md")

	r := runner.NewRunner(cfg)
	err = r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	assertTagExists(t, dir, "v2.1.0")
}

func TestIntegration_NoIncrement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.5.0")
	createCommits(t, dir, []string{
		"chore: update docs",
	})

	cfg := newTestConfig(dir)
	cfg.Git.Commit = false
	cfg.Git.Tag = false

	r := runner.NewRunner(cfg)
	err := r.RunNoIncrement()
	if err != nil {
		t.Fatalf("RunNoIncrement() failed: %v", err)
	}

	// Should still generate changelog but not create new tags beyond v1.5.0
	assertTagNotExists(t, dir, "v1.5.1")
}

func TestIntegration_PreRelease_BranchAware_ContinueSeries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	// Setup: main branch with v1.2.4
	initGitRepo(t, dir)
	createTag(t, dir, "v1.2.4")

	// Create deneme branch and start pre-release series
	runGit(t, dir, "checkout", "-b", "deneme")
	createCommits(t, dir, []string{"feat: deneme feature 1"})
	createTag(t, dir, "v1.2.5-deneme.0")

	// Add another commit - this should continue the series
	createCommits(t, dir, []string{"feat: deneme feature 2"})

	cfg := newTestConfig(dir)
	cfg.PreReleaseID = "deneme"
	cfg.Increment = "patch"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Should continue series: v1.2.5-deneme.0 → v1.2.5-deneme.1
	assertTagExists(t, dir, "v1.2.5-deneme.1")
}

func TestIntegration_PreRelease_BranchAware_NewSeries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	// Setup: main branch with v1.2.4, then advance to v2.0.0
	initGitRepo(t, dir)
	createTag(t, dir, "v1.2.4")
	createCommits(t, dir, []string{"feat!: breaking change"})
	createTag(t, dir, "v2.0.0")

	// Create new deneme branch from v2.0.0
	// (old deneme tags like v1.2.5-deneme.0 are NOT reachable from this branch)
	runGit(t, dir, "checkout", "-b", "deneme")
	createCommits(t, dir, []string{"feat: new deneme feature"})

	cfg := newTestConfig(dir)
	cfg.PreReleaseID = "deneme"
	cfg.Increment = "patch"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Should start new series from v2.0.0: v2.0.1-deneme.0
	assertTagExists(t, dir, "v2.0.1-deneme.0")
}

func TestIntegration_PreRelease_BranchAware_MasterAdvanced(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	// Setup: main branch with v1.2.4
	initGitRepo(t, dir)
	defaultBranch := strings.TrimSpace(runGit(t, dir, "rev-parse", "--abbrev-ref", "HEAD"))
	createTag(t, dir, "v1.2.4")

	// Create deneme branch and start pre-release series
	runGit(t, dir, "checkout", "-b", "deneme")
	createCommits(t, dir, []string{"feat: deneme feature"})
	createTag(t, dir, "v1.2.5-deneme.0")

	// Go back to default branch and advance it (deneme branch doesn't see these)
	runGit(t, dir, "checkout", defaultBranch)
	createCommits(t, dir, []string{"feat: main feature"})
	createTag(t, dir, "v1.3.0")
	createCommits(t, dir, []string{"feat!: breaking"})
	createTag(t, dir, "v2.0.0")

	// Back to deneme - default branch advanced but deneme.0 is still reachable
	runGit(t, dir, "checkout", "deneme")
	createCommits(t, dir, []string{"feat: another deneme feature"})

	cfg := newTestConfig(dir)
	cfg.PreReleaseID = "deneme"
	cfg.Increment = "patch"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// deneme.0 is reachable, base(1.2.5) >= stable(1.2.4) → continue series
	assertTagExists(t, dir, "v1.2.5-deneme.1")
}

func TestIntegration_PreRelease_NoFlag_BehaviorUnchanged(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"feat: new feature"})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"
	// No PreReleaseID set - standard behavior

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Standard behavior: 1.0.0 → 1.1.0
	assertTagExists(t, dir, "v1.1.0")
}

func TestIntegration_TagFormatChange_VPrefixRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)

	// Phase 1: Tags with v prefix (old format)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"feat: feature one"})
	createTag(t, dir, "v1.1.0")

	// Phase 2: Developer changes tagName to "${version}" (no v prefix)
	createCommits(t, dir, []string{"feat: feature two"})
	cfg := newTestConfig(dir)
	cfg.Git.TagName = "${version}" // No more v prefix
	cfg.Git.Commit = false
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	if err := r.Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Should create "1.2.0" not "v1.2.0"
	assertTagExists(t, dir, "1.2.0")

	// Phase 3: Another release with new format — should find "1.2.0" as latest
	createCommits(t, dir, []string{"fix: bugfix"})
	cfg2 := newTestConfig(dir)
	cfg2.Git.TagName = "${version}"
	cfg2.Git.Commit = false
	cfg2.Increment = "patch"

	r2 := runner.NewRunner(cfg2)
	if err := r2.Run(); err != nil {
		t.Fatalf("Second Run() failed: %v", err)
	}

	// Should find "1.2.0" as latest (not "v1.1.0") and produce "1.2.1"
	assertTagExists(t, dir, "1.2.1")
	assertTagNotExists(t, dir, "v1.2.1")
}

// --- Phase 19: New integration test scenarios ---

func TestIntegration_CalVer_YearMonthMinor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createCommits(t, dir, []string{"feat: initial feature"})

	cfg := newTestConfig(dir)
	cfg.CalVer = config.CalVerConfig{
		Enabled:           true,
		Format:            "yyyy.mm.minor",
		FallbackIncrement: "minor",
	}
	cfg.Increment = ""

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	now := time.Now()
	expectedTag := fmt.Sprintf("v%d.%d.0", now.Year(), int(now.Month()))
	assertTagExists(t, dir, expectedTag)
}

func TestIntegration_TagFormat_NoPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "1.0.0") // No v prefix
	createCommits(t, dir, []string{"feat: new feature"})

	cfg := newTestConfig(dir)
	cfg.Git.TagName = "${version}" // No v prefix
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	assertTagExists(t, dir, "1.1.0")
	assertTagNotExists(t, dir, "v1.1.0") // Should NOT create v-prefixed tag
}

func TestIntegration_TagFormat_SequentialNoPrefix(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "1.0.0") // No prefix
	createCommits(t, dir, []string{"fix: first fix"})

	cfg := newTestConfig(dir)
	cfg.Git.TagName = "${version}"
	cfg.Increment = "patch"

	r := runner.NewRunner(cfg)
	if err := r.Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}
	assertTagExists(t, dir, "1.0.1")

	// Second release from 1.0.1 → 1.1.0
	createCommits(t, dir, []string{"feat: new feature"})
	cfg2 := newTestConfig(dir)
	cfg2.Git.TagName = "${version}"
	cfg2.Increment = "minor"
	r2 := runner.NewRunner(cfg2)
	if err := r2.Run(); err != nil {
		t.Fatalf("Second Run() failed: %v", err)
	}
	assertTagExists(t, dir, "1.1.0")
}

func TestIntegration_WorkingDirCleanAfterRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"feat: new feature"})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	if err := r.Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	assertTagExists(t, dir, "v1.1.0")
	assertWorkingDirClean(t, dir)
	assertTagCount(t, dir, 2) // v1.0.0 + v1.1.0
	assertFileContains(t, filepath.Join(dir, "CHANGELOG.md"), "1.1.0")
}

func TestIntegration_Error_DirtyWorkingDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"feat: feature"})

	// Create uncommitted file to make working dir dirty
	writeFile(t, filepath.Join(dir, "dirty.txt"), "uncommitted content")

	cfg := newTestConfig(dir)
	cfg.Git.RequireCleanWorkingDir = true
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err == nil {
		t.Fatal("expected error for dirty working directory")
	}
	if !strings.Contains(err.Error(), "not clean") {
		t.Errorf("expected 'not clean' in error, got: %v", err)
	}
}

func TestIntegration_Error_NoCommitsSinceTag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	// No commits after tag

	cfg := newTestConfig(dir)
	cfg.Git.RequireCommits = true
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.Run()
	// Should not error but should exit gracefully (no new tag)
	if err != nil {
		t.Fatalf("expected graceful exit, got error: %v", err)
	}
	assertTagNotExists(t, dir, "v1.1.0")
}

func TestIntegration_Changelog_Disabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"feat: new feature"})

	cfg := newTestConfig(dir)
	cfg.Changelog.Enabled = false
	cfg.Git.Commit = false // No changelog file → nothing to commit
	cfg.Increment = "minor"

	r := runner.NewRunner(cfg)
	err := r.Run()
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	// Tag should be created
	assertTagExists(t, dir, "v1.1.0")
	// CHANGELOG.md should NOT be created
	assertFileNotExists(t, filepath.Join(dir, "CHANGELOG.md"))
}

func TestIntegration_SequentialReleases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(origDir) }()

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")

	// First release
	createCommits(t, dir, []string{"fix: first fix"})
	cfg1 := newTestConfig(dir)
	cfg1.Increment = "patch"
	r1 := runner.NewRunner(cfg1)
	if err := r1.Run(); err != nil {
		t.Fatalf("First Run() failed: %v", err)
	}
	assertTagExists(t, dir, "v1.0.1")

	// Second release
	createCommits(t, dir, []string{"feat: new feature"})
	cfg2 := newTestConfig(dir)
	cfg2.Increment = "minor"
	r2 := runner.NewRunner(cfg2)
	if err := r2.Run(); err != nil {
		t.Fatalf("Second Run() failed: %v", err)
	}
	assertTagExists(t, dir, "v1.1.0")
}
