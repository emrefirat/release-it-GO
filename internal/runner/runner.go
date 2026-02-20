package runner

import (
	"fmt"
	"os"
	"strings"
	"time"

	"release-it-go/internal/bumper"
	"release-it-go/internal/changelog"
	"release-it-go/internal/config"
	"release-it-go/internal/git"
	"release-it-go/internal/notification"
	"release-it-go/internal/release"
	"release-it-go/internal/ui"
	"release-it-go/internal/version"
)

// Runner orchestrates the release pipeline.
type Runner struct {
	ctx *ReleaseContext
}

// NewRunner creates a new pipeline runner from configuration.
func NewRunner(cfg *config.Config) *Runner {
	ctx := NewReleaseContext(cfg)
	return &Runner{ctx: ctx}
}

// pipelineStep defines a named step in the release pipeline.
type pipelineStep struct {
	name string
	fn   func() error
}

// printBanner prints the release-it-go banner at the start of the pipeline.
func (r *Runner) printBanner() {
	if r.ctx.IsDryRun {
		fmt.Fprintf(os.Stderr, "\n%s %s %s\n\n", ui.IconDryRun, ui.FormatBold("release-it-go"), ui.FormatDim("(dry-run)"))
	} else {
		fmt.Fprintf(os.Stderr, "\n%s %s\n\n", ui.IconRocket, ui.FormatBold("release-it-go"))
	}
}

// Run executes the full release pipeline.
func (r *Runner) Run() error {
	start := time.Now()
	r.printBanner()

	steps := []pipelineStep{
		{"init", r.init},
		{"prerequisites", r.checkPrerequisites},
		{"commitlint", r.checkCommitLint},
		{"version", r.determineVersion},
		{"bump", r.bumpFiles},
		{"changelog", r.generateChangelog},
		{"git:release", r.gitRelease},
		{"github:release", r.githubRelease},
		{"gitlab:release", r.gitlabRelease},
		{"notification", r.sendNotification},
	}

	for _, step := range steps {
		if err := r.ctx.HookRunner.RunHooks("before:" + step.name); err != nil {
			return fmt.Errorf("before:%s hook: %w", step.name, err)
		}

		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}

		// Early exit if no commits to release
		if r.ctx.noCommits {
			return nil
		}

		r.ctx.UpdateVars()

		if err := r.ctx.HookRunner.RunHooks("after:" + step.name); err != nil {
			return fmt.Errorf("after:%s hook: %w", step.name, err)
		}
	}

	r.printSummary(time.Since(start))
	return nil
}

// RunChangelogOnly generates and prints the changelog without performing a release.
func (r *Runner) RunChangelogOnly() error {
	if err := r.init(); err != nil {
		return err
	}
	if err := r.determineVersion(); err != nil {
		return err
	}

	latestTag := latestVersionToTag(r.ctx.LatestVersion)

	commits, err := r.ctx.Git.GetCommitsSinceTag(latestTag)
	if err != nil {
		return fmt.Errorf("getting commits: %w", err)
	}

	rawCommits := make([]changelog.RawCommit, len(commits))
	for i, msg := range commits {
		rawCommits[i] = changelog.RawCommit{Hash: "", Message: msg}
	}

	parsed := changelog.ParseCommits(rawCommits)
	opts := changelog.Options{
		KeepAChangelog: r.ctx.Config.Changelog.KeepAChangelog,
		RepoInfo:       r.ctx.RepoInfo,
	}
	changelogContent := changelog.GenerateChangelog(parsed, r.ctx.Version, r.ctx.LatestVersion, opts)
	fmt.Println(changelogContent)
	return nil
}

// RunReleaseVersionOnly determines and prints the next version.
func (r *Runner) RunReleaseVersionOnly() error {
	if err := r.init(); err != nil {
		return err
	}
	if err := r.determineVersion(); err != nil {
		return err
	}
	fmt.Println(r.ctx.Version)
	return nil
}

// RunOnlyVersion prompts for version selection, then runs the rest automatically.
func (r *Runner) RunOnlyVersion() error {
	r.printBanner()

	if err := r.init(); err != nil {
		return err
	}
	if err := r.determineVersion(); err != nil {
		return err
	}

	// After version is selected, run rest in CI mode (no prompts)
	r.ctx.IsCI = true

	steps := []pipelineStep{
		{"bump", r.bumpFiles},
		{"changelog", r.generateChangelog},
		{"git:release", r.gitRelease},
		{"github:release", r.githubRelease},
		{"gitlab:release", r.gitlabRelease},
		{"notification", r.sendNotification},
	}

	start := time.Now()
	for _, step := range steps {
		if err := r.ctx.HookRunner.RunHooks("before:" + step.name); err != nil {
			return fmt.Errorf("before:%s hook: %w", step.name, err)
		}
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
		r.ctx.UpdateVars()
		if err := r.ctx.HookRunner.RunHooks("after:" + step.name); err != nil {
			return fmt.Errorf("after:%s hook: %w", step.name, err)
		}
	}

	r.printSummary(time.Since(start))
	return nil
}

// RunNoIncrement runs the release pipeline without incrementing the version.
func (r *Runner) RunNoIncrement() error {
	r.printBanner()

	if err := r.init(); err != nil {
		return err
	}

	// Get latest version but don't increment
	latestTag, err := r.ctx.Git.GetLatestTag()
	if err != nil {
		return fmt.Errorf("getting latest tag: %w", err)
	}

	parsed, parseErr := version.ParseVersion(latestTag)
	if parseErr != nil {
		return fmt.Errorf("parsing version %q: %w", latestTag, parseErr)
	}

	r.ctx.LatestVersion = parsed.String()
	r.ctx.Version = parsed.String()
	r.ctx.TagName = renderTagName(r.ctx.Config.Git.TagName, r.ctx.Version)
	r.ctx.UpdateVars()

	// Run remaining steps
	start := time.Now()
	steps := []pipelineStep{
		{"changelog", r.generateChangelog},
		{"git:release", r.gitRelease},
		{"github:release", r.githubRelease},
		{"gitlab:release", r.gitlabRelease},
		{"notification", r.sendNotification},
	}

	for _, step := range steps {
		if err := r.ctx.HookRunner.RunHooks("before:" + step.name); err != nil {
			return fmt.Errorf("before:%s hook: %w", step.name, err)
		}
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}
		r.ctx.UpdateVars()
		if err := r.ctx.HookRunner.RunHooks("after:" + step.name); err != nil {
			return fmt.Errorf("after:%s hook: %w", step.name, err)
		}
	}

	r.printSummary(time.Since(start))
	return nil
}

// init initializes the release context with repo info and branch name.
func (r *Runner) init() error {
	r.ctx.Spinner.Start("Initialized")

	repoInfo, err := git.GetRepoInfo("")
	if err != nil {
		r.ctx.Logger.Verbose("Could not get repo info: %v", err)
		// Non-fatal: repo info is optional for local-only operations
	} else {
		r.ctx.RepoInfo = repoInfo
	}

	branchName, err := git.GetBranchName()
	if err != nil {
		r.ctx.Logger.Verbose("Could not get branch name: %v", err)
	} else {
		r.ctx.BranchName = branchName
	}

	r.ctx.UpdateVars()
	r.ctx.Spinner.Stop(true)
	return nil
}

// errNoCommits is a sentinel value to detect the "no commits" condition.
var errNoCommits = "no commits since latest tag"

// checkPrerequisites runs all prerequisite checks.
func (r *Runner) checkPrerequisites() error {
	r.ctx.Spinner.Start("Prerequisites checked")

	if err := r.ctx.Git.CheckPrerequisites(); err != nil {
		if strings.Contains(err.Error(), errNoCommits) {
			r.ctx.Spinner.Stop(true)
			r.ctx.Logger.Print("  %s No commits since latest tag. Nothing to release.", ui.IconWarning)
			r.ctx.noCommits = true
			return nil
		}
		r.ctx.Spinner.Stop(false)
		return err
	}

	if err := r.checkTokens(); err != nil {
		r.ctx.Spinner.Stop(false)
		return err
	}

	r.ctx.Spinner.Stop(true)
	return nil
}

// checkTokens verifies that required API tokens are set and valid when
// GitHub/GitLab releases are enabled. This catches missing or invalid tokens
// early in the pipeline instead of failing late during the release step.
func (r *Runner) checkTokens() error {
	cfg := r.ctx.Config

	if cfg.GitHub.Release && !cfg.GitHub.SkipChecks {
		tokenRef := cfg.GitHub.TokenRef
		if tokenRef == "" {
			tokenRef = "GITHUB_TOKEN"
		}
		if os.Getenv(tokenRef) == "" {
			return fmt.Errorf("GitHub release is enabled but %s is not set", tokenRef)
		}
		if r.ctx.RepoInfo != nil {
			client, err := release.NewGitHubClient(&cfg.GitHub, r.ctx.RepoInfo, r.ctx.Logger, r.ctx.IsDryRun)
			if err != nil {
				return fmt.Errorf("GitHub client: %w", err)
			}
			if err := client.ValidateToken(); err != nil {
				return err
			}
		}
	}

	if cfg.GitLab.Release && !cfg.GitLab.SkipChecks {
		tokenRef := cfg.GitLab.TokenRef
		if tokenRef == "" {
			tokenRef = "GITLAB_TOKEN"
		}
		if os.Getenv(tokenRef) == "" {
			return fmt.Errorf("GitLab release is enabled but %s is not set", tokenRef)
		}
		if r.ctx.RepoInfo != nil {
			client, err := release.NewGitLabClient(&cfg.GitLab, r.ctx.RepoInfo, r.ctx.Logger, r.ctx.IsDryRun)
			if err != nil {
				return fmt.Errorf("GitLab client: %w", err)
			}
			if err := client.ValidateToken(); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkCommitLint validates that commits since last tag follow conventional commit format.
func (r *Runner) checkCommitLint() error {
	if !r.ctx.Config.Git.RequireConventionalCommits {
		return nil
	}

	r.ctx.Spinner.Start("Commit conventions checked")

	latestTag := latestVersionToTag(r.ctx.LatestVersion)
	if r.ctx.LatestVersion == "" {
		// Try to get latest tag for lint check before version is determined
		tag, err := r.ctx.Git.GetLatestTag()
		if err == nil && tag != "" {
			latestTag = tag
		}
	}

	commitInfos, err := r.ctx.Git.GetCommitsWithHashSinceTag(latestTag)
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("getting commits for lint: %w", err)
	}

	if len(commitInfos) == 0 {
		r.ctx.Spinner.Stop(true)
		return nil
	}

	lintInputs := make([]changelog.LintInput, len(commitInfos))
	for i, ci := range commitInfos {
		lintInputs[i] = changelog.LintInput{Hash: ci.Hash, Subject: ci.Subject}
	}

	_, failed := changelog.LintCommits(lintInputs)
	if len(failed) > 0 {
		r.ctx.Spinner.Stop(false)
		return formatLintError(failed, len(commitInfos))
	}

	r.ctx.Spinner.Stop(true)
	return nil
}

// RunCheckCommits runs commit lint as a standalone operation and prints results.
func (r *Runner) RunCheckCommits() error {
	if err := r.init(); err != nil {
		return err
	}

	latestTag, err := r.ctx.Git.GetLatestTag()
	if err != nil {
		r.ctx.Logger.Verbose("No previous tags found")
		latestTag = ""
	}

	commitInfos, err := r.ctx.Git.GetCommitsWithHashSinceTag(latestTag)
	if err != nil {
		return fmt.Errorf("getting commits: %w", err)
	}

	if len(commitInfos) == 0 {
		fmt.Println("No commits found to lint.")
		return nil
	}

	lintInputs := make([]changelog.LintInput, len(commitInfos))
	for i, ci := range commitInfos {
		lintInputs[i] = changelog.LintInput{Hash: ci.Hash, Subject: ci.Subject}
	}

	passed, failed := changelog.LintCommits(lintInputs)

	// Verbose: show all checked commits with their status
	if r.ctx.Logger.GetVerbose() >= 1 {
		for _, p := range passed {
			r.ctx.Logger.Print("  %s %s %s", ui.FormatSuccess(ui.IconSuccess), p.Hash[:7], p.Subject)
		}
		for _, f := range failed {
			r.ctx.Logger.Print("  %s %s %s ← %s", ui.FormatError(ui.IconFail), f.Hash[:7], f.Subject, f.Reason)
		}
		fmt.Fprintln(os.Stderr)
	}

	if len(failed) == 0 {
		fmt.Printf("All %d commits are conventional. %s\n", len(passed), ui.IconSuccess)
		return nil
	}

	return formatLintError(failed, len(commitInfos))
}

// formatLintError builds a formatted error message for failed commit lints.
func formatLintError(failed []changelog.LintResult, total int) error {
	var b strings.Builder
	b.WriteString("Commit lint failed:\n")
	for _, f := range failed {
		fmt.Fprintf(&b, "  %-10s %-40s ← %s\n", f.Hash, f.Subject, f.Reason)
	}
	fmt.Fprintf(&b, "\n  %d of %d commits are not conventional.\n", len(failed), total)
	b.WriteString("  Use --ignore-commit-lint to bypass.\n")
	return fmt.Errorf("%s", b.String())
}

// determineVersion determines the next version based on config and commits.
func (r *Runner) determineVersion() error {
	// Try reading version from bumper input file first
	var bumperVersion string
	if r.ctx.Config.Bumper.Enabled && r.ctx.Config.Bumper.In != nil {
		b := bumper.NewBumper(&r.ctx.Config.Bumper, r.ctx.Logger, r.ctx.IsDryRun)
		v, err := b.ReadVersion()
		if err != nil {
			r.ctx.Logger.Verbose("Could not read version from bumper: %v", err)
		} else if v != "" {
			bumperVersion = v
			r.ctx.Logger.Verbose("Read version from bumper: %s", v)
		}
	}

	// Get latest version from git tags
	latestTag, err := r.ctx.Git.GetLatestTag()
	if err != nil {
		r.ctx.Logger.Verbose("No previous tags found, starting from 0.0.0")
		latestTag = "0.0.0"
	}

	// Use bumper version if available and no git tag
	if bumperVersion != "" && latestTag == "0.0.0" {
		latestTag = bumperVersion
	}

	// Branch-aware pre-release: resolve base tag from merged tags only
	preReleaseID := r.ctx.Config.PreReleaseID
	if preReleaseID != "" {
		resolved, resolveErr := r.resolvePreReleaseBaseTag(preReleaseID)
		if resolveErr != nil {
			r.ctx.Logger.Verbose("Could not resolve branch-aware pre-release tag: %v", resolveErr)
		} else if resolved != "" {
			latestTag = resolved
			r.ctx.Logger.Verbose("Branch-aware pre-release base tag: %s", resolved)
		}
	}

	latestVersion := latestTag
	parsed, parseErr := version.ParseVersion(latestTag)
	if parseErr == nil {
		latestVersion = parsed.String()
	}
	r.ctx.LatestVersion = latestVersion

	// If explicit version is set in config, use it
	if r.ctx.Config.Increment == "no-increment" {
		r.ctx.Version = latestVersion
		r.ctx.TagName = renderTagName(r.ctx.Config.Git.TagName, r.ctx.Version)
		return nil
	}

	// CalVer mode
	if r.ctx.Config.CalVer.Enabled {
		return r.determineCalVer(latestVersion)
	}

	// SemVer mode
	return r.determineSemVer(latestVersion)
}

// determineCalVer calculates the next calendar version.
func (r *Runner) determineCalVer(latestVersion string) error {
	cv := version.NewCalVer(
		r.ctx.Config.CalVer.Format,
		r.ctx.Config.CalVer.Increment,
		r.ctx.Config.CalVer.FallbackIncrement,
	)

	newVersion, err := cv.NextVersion(latestVersion)
	if err != nil {
		return fmt.Errorf("calculating CalVer: %w", err)
	}

	r.ctx.Version = newVersion
	r.ctx.TagName = renderTagName(r.ctx.Config.Git.TagName, newVersion)
	r.ctx.Logger.Print("  %s Version (CalVer): %s → %s", ui.IconVersion, latestVersion, newVersion)
	return nil
}

// determineSemVer calculates the next semantic version.
func (r *Runner) determineSemVer(latestVersion string) error {
	// Determine increment type
	increment := r.ctx.Config.Increment
	if increment == "" {
		increment = r.autoDetectIncrement()
	}

	if increment == "" {
		increment = "patch"
	}

	parsedCurrent, parseErr := version.ParseVersion(latestVersion)
	if parseErr != nil {
		return fmt.Errorf("parsing current version %q: %w", latestVersion, parseErr)
	}

	preReleaseID := r.ctx.Config.PreReleaseID
	incrementType := increment
	if preReleaseID != "" {
		// If current version is already a pre-release with the same ID,
		// use "prerelease" to increment the number (e.g. beta.0 → beta.1).
		// Otherwise use "pre+increment" to start a new pre-release series.
		if parsedCurrent.Prerelease() != "" && strings.HasPrefix(parsedCurrent.Prerelease(), preReleaseID+".") {
			incrementType = "prerelease"
		} else {
			incrementType = "pre" + increment
		}
	}

	newSemver, err := version.IncrementVersion(parsedCurrent, incrementType, preReleaseID)
	if err != nil {
		return fmt.Errorf("incrementing version: %w", err)
	}
	newVersionStr := newSemver.String()

	// Interactive mode: let user choose
	if !r.ctx.IsCI && increment == r.autoDetectIncrement() {
		options := r.buildVersionOptions(latestVersion, increment)
		if len(options) > 0 {
			selected, err := r.ctx.Prompter.SelectVersion(latestVersion, newVersionStr, options)
			if err != nil {
				return err
			}
			newVersionStr = selected
		}
	}

	r.ctx.Version = newVersionStr
	r.ctx.TagName = renderTagName(r.ctx.Config.Git.TagName, newVersionStr)
	r.ctx.Logger.Print("  %s Version: %s → %s", ui.IconVersion, latestVersion, newVersionStr)
	return nil
}

// resolvePreReleaseBaseTag determines the correct base tag for pre-release versioning
// by looking only at tags merged into the current HEAD. This prevents cross-branch
// tag pollution (e.g., beta tags from another branch affecting the "deneme" series).
//
// Algorithm:
//  1. Find the latest pre-release tag merged into HEAD with matching preReleaseID
//  2. Find the latest stable (non-pre-release) tag merged into HEAD
//  3. If pre-release tag exists and its base version >= stable → continue series
//  4. Otherwise → return stable tag (or "") to start a new series
func (r *Runner) resolvePreReleaseBaseTag(preReleaseID string) (string, error) {
	preTag, err := r.ctx.Git.GetLatestPreReleaseTagMerged(preReleaseID)
	if err != nil {
		return "", fmt.Errorf("getting merged pre-release tag: %w", err)
	}

	stableTag, err := r.ctx.Git.GetLatestStableTagMerged()
	if err != nil {
		return "", fmt.Errorf("getting merged stable tag: %w", err)
	}

	// No pre-release tag found for this ID → new series from stable or default
	if preTag == "" {
		return stableTag, nil
	}

	// Pre-release tag found, check if it's still valid
	// (its base version should be >= the latest stable version)
	if stableTag == "" {
		// No stable tag, pre-release tag is the base
		return preTag, nil
	}

	parsedPre, err := version.ParseVersion(preTag)
	if err != nil {
		return stableTag, nil
	}

	parsedStable, err := version.ParseVersion(stableTag)
	if err != nil {
		return preTag, nil
	}

	// Compare base versions: strip pre-release from pre-release tag
	preBase := fmt.Sprintf("%d.%d.%d", parsedPre.Major(), parsedPre.Minor(), parsedPre.Patch())
	preBaseParsed, err := version.ParseVersion(preBase)
	if err != nil {
		return stableTag, nil
	}

	// If pre-release base >= stable → continue series
	if preBaseParsed.Compare(parsedStable) >= 0 {
		return preTag, nil
	}

	// Pre-release base < stable → new series
	return stableTag, nil
}

// bumpFiles writes the new version to configured bumper output files.
func (r *Runner) bumpFiles() error {
	if !r.ctx.Config.Bumper.Enabled || len(r.ctx.Config.Bumper.Out) == 0 {
		return nil
	}

	r.ctx.Spinner.Start("Version files updated")

	b := bumper.NewBumper(&r.ctx.Config.Bumper, r.ctx.Logger, r.ctx.IsDryRun)
	if err := b.WriteVersion(r.ctx.Version); err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("bumping version files: %w", err)
	}

	r.ctx.Spinner.Stop(true)
	return nil
}

// autoDetectIncrement uses conventional commits to determine the bump type.
func (r *Runner) autoDetectIncrement() string {
	latestTag := latestVersionToTag(r.ctx.LatestVersion)

	commits, err := r.ctx.Git.GetCommitsSinceTag(latestTag)
	if err != nil || len(commits) == 0 {
		return "patch"
	}

	rawCommits := make([]changelog.RawCommit, len(commits))
	for i, msg := range commits {
		rawCommits[i] = changelog.RawCommit{Hash: "", Message: msg}
	}

	parsed := changelog.ParseCommits(rawCommits)
	bump := changelog.AnalyzeBump(parsed)

	if bump == changelog.BumpNone {
		return "patch"
	}
	return bump.String()
}

// buildVersionOptions creates version options for the interactive prompt.
func (r *Runner) buildVersionOptions(current string, recommended string) []ui.VersionOption {
	options := make([]ui.VersionOption, 0, 3)

	parsedCurrent, err := version.ParseVersion(current)
	if err != nil {
		return options
	}

	for _, inc := range []string{"patch", "minor", "major"} {
		ver, err := version.IncrementVersion(parsedCurrent, inc, "")
		if err != nil {
			continue
		}
		verStr := ver.String()
		options = append(options, ui.VersionOption{
			Label:       fmt.Sprintf("%s (%s)", inc, verStr),
			Version:     verStr,
			Recommended: inc == recommended,
		})
	}

	return options
}

// generateChangelog creates changelog content.
func (r *Runner) generateChangelog() error {
	if !r.ctx.Config.Changelog.Enabled {
		return nil
	}

	r.ctx.Spinner.Start("Changelog generated")

	latestTag := latestVersionToTag(r.ctx.LatestVersion)

	commits, err := r.ctx.Git.GetCommitsSinceTag(latestTag)
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("getting commits: %w", err)
	}

	rawCommits := make([]changelog.RawCommit, len(commits))
	for i, msg := range commits {
		rawCommits[i] = changelog.RawCommit{Hash: "", Message: msg}
	}

	parsed := changelog.ParseCommits(rawCommits)
	opts := changelog.Options{
		KeepAChangelog: r.ctx.Config.Changelog.KeepAChangelog,
		RepoInfo:       r.ctx.RepoInfo,
	}

	changelogContent := changelog.GenerateChangelog(parsed, r.ctx.Version, r.ctx.LatestVersion, opts)
	r.ctx.Changelog = changelogContent

	// Update CHANGELOG.md file if configured
	if r.ctx.Config.Changelog.Infile != "" && !r.ctx.IsDryRun {
		header := r.ctx.Config.Changelog.Header
		if err := changelog.UpdateChangelogFile(r.ctx.Config.Changelog.Infile, changelogContent, header); err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("updating changelog file: %w", err)
		}
		// Explicitly stage the changelog file so it's included in the release commit
		if err := r.ctx.Git.StageFile(r.ctx.Config.Changelog.Infile); err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("staging changelog file: %w", err)
		}
	} else if r.ctx.Config.Changelog.Infile != "" && r.ctx.IsDryRun {
		r.ctx.Logger.DryRun("Would update %s", r.ctx.Config.Changelog.Infile)
	}

	r.ctx.Spinner.Stop(true)
	return nil
}

// gitRelease performs git stage, commit, tag, and push.
func (r *Runner) gitRelease() error {
	cfg := &r.ctx.Config.Git

	// Stage
	if cfg.Commit {
		r.ctx.Spinner.Start("Files staged")
		if err := r.ctx.Git.Stage(); err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("staging: %w", err)
		}
		r.ctx.Spinner.Stop(true)

		// Commit
		commitMsg := renderTagName(cfg.CommitMessage, r.ctx.Version)
		if !r.ctx.IsCI {
			confirmed, err := r.ctx.Prompter.Confirm(
				fmt.Sprintf("Commit (%s)?", commitMsg), true)
			if err != nil {
				return err
			}
			if !confirmed {
				r.ctx.Logger.Print("  %s Skipped commit", ui.IconSkip)
				return nil
			}
		}

		r.ctx.Spinner.Start("Committed")
		if err := r.ctx.Git.Commit(commitMsg); err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("commit: %w", err)
		}
		r.ctx.Spinner.Stop(true)
	}

	// Tag
	if cfg.Tag {
		annotation := renderTagName(cfg.TagAnnotation, r.ctx.Version)

		if !r.ctx.IsCI {
			confirmed, err := r.ctx.Prompter.Confirm(
				fmt.Sprintf("Tag (%s)?", r.ctx.TagName), true)
			if err != nil {
				return err
			}
			if !confirmed {
				r.ctx.Logger.Print("  %s Skipped tag", ui.IconSkip)
				return nil
			}
		}

		r.ctx.Spinner.Start(fmt.Sprintf("Tagged %s", r.ctx.TagName))
		if err := r.ctx.Git.CreateTag(r.ctx.TagName, annotation); err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("tag: %w", err)
		}
		r.ctx.Spinner.Stop(true)
	}

	// Push
	if cfg.Push {
		if !r.ctx.IsCI {
			confirmed, err := r.ctx.Prompter.Confirm("Push?", true)
			if err != nil {
				return err
			}
			if !confirmed {
				r.ctx.Logger.Print("  %s Skipped push", ui.IconSkip)
				return nil
			}
		}

		r.ctx.Spinner.Start("Pushed to remote")
		if err := r.ctx.Git.Push(); err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("push: %w", err)
		}
		r.ctx.Spinner.Stop(true)
	}

	return nil
}

// githubRelease creates a GitHub release.
func (r *Runner) githubRelease() error {
	if !r.ctx.Config.GitHub.Release || r.ctx.RepoInfo == nil {
		return nil
	}

	if !r.ctx.IsCI {
		releaseName := renderTagName(r.ctx.Config.GitHub.ReleaseName, r.ctx.Version)
		confirmed, err := r.ctx.Prompter.Confirm(
			fmt.Sprintf("Create a release on GitHub (%s)?", releaseName), true)
		if err != nil {
			return err
		}
		if !confirmed {
			r.ctx.Logger.Print("  %s Skipped GitHub release", ui.IconSkip)
			return nil
		}
	}

	r.ctx.Spinner.Start("GitHub release created")

	client, err := release.NewGitHubClient(&r.ctx.Config.GitHub, r.ctx.RepoInfo, r.ctx.Logger, r.ctx.IsDryRun)
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("GitHub client: %w", err)
	}

	releaseName := renderTagName(r.ctx.Config.GitHub.ReleaseName, r.ctx.Version)
	releaseNotes := r.ctx.Changelog

	result, err := client.CreateRelease(release.ReleaseOptions{
		TagName:            r.ctx.TagName,
		ReleaseName:        releaseName,
		ReleaseNotes:       releaseNotes,
		Draft:              r.ctx.Config.GitHub.Draft,
		PreRelease:         r.ctx.Config.GitHub.PreRelease,
		MakeLatest:         r.ctx.Config.GitHub.MakeLatest,
		AutoGenerate:       r.ctx.Config.GitHub.AutoGenerate,
		DiscussionCategory: r.ctx.Config.GitHub.DiscussionCategoryName,
	})
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return err
	}

	r.ctx.ReleaseURL = result.URL

	// Upload assets
	if len(r.ctx.Config.GitHub.Assets) > 0 {
		assets, err := release.ResolveAssets(r.ctx.Config.GitHub.Assets)
		if err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("resolving assets: %w", err)
		}
		if len(assets) > 0 {
			if err := client.UploadAssets(result.ID, assets); err != nil {
				r.ctx.Spinner.Stop(false)
				return err
			}
		}
	}

	r.ctx.Spinner.Stop(true)
	return nil
}

// gitlabRelease creates a GitLab release.
func (r *Runner) gitlabRelease() error {
	if !r.ctx.Config.GitLab.Release || r.ctx.RepoInfo == nil {
		return nil
	}

	if !r.ctx.IsCI {
		releaseName := renderTagName(r.ctx.Config.GitLab.ReleaseName, r.ctx.Version)
		confirmed, err := r.ctx.Prompter.Confirm(
			fmt.Sprintf("Create a release on GitLab (%s)?", releaseName), true)
		if err != nil {
			return err
		}
		if !confirmed {
			r.ctx.Logger.Print("  %s Skipped GitLab release", ui.IconSkip)
			return nil
		}
	}

	r.ctx.Spinner.Start("GitLab release created")

	client, err := release.NewGitLabClient(&r.ctx.Config.GitLab, r.ctx.RepoInfo, r.ctx.Logger, r.ctx.IsDryRun)
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("GitLab client: %w", err)
	}

	releaseName := renderTagName(r.ctx.Config.GitLab.ReleaseName, r.ctx.Version)
	releaseNotes := r.ctx.Changelog

	result, err := client.CreateRelease(release.ReleaseOptions{
		TagName:      r.ctx.TagName,
		ReleaseName:  releaseName,
		ReleaseNotes: releaseNotes,
	})
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return err
	}

	r.ctx.ReleaseURL = result.URL

	// Upload assets
	if len(r.ctx.Config.GitLab.Assets) > 0 {
		assets, err := release.ResolveAssets(r.ctx.Config.GitLab.Assets)
		if err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("resolving assets: %w", err)
		}
		if len(assets) > 0 {
			if err := client.UploadAssets(result.ID, assets); err != nil {
				r.ctx.Spinner.Stop(false)
				return err
			}
		}
	}

	r.ctx.Spinner.Stop(true)
	return nil
}

// sendNotification sends webhook notifications to configured endpoints.
// This step is non-fatal: if notifications fail, a warning is logged but the pipeline continues.
func (r *Runner) sendNotification() error {
	cfg := r.ctx.Config.Notification
	if !cfg.Enabled || len(cfg.Webhooks) == 0 {
		return nil
	}

	r.ctx.Spinner.Start("Notifications sent")

	client := notification.NewClient(cfg.Webhooks, r.ctx.Vars, r.ctx.Logger, r.ctx.IsDryRun)
	if err := client.SendAll(); err != nil {
		r.ctx.Spinner.Stop(false)
		r.ctx.Logger.Warn("Notification failed: %v", err)
		return nil // Non-fatal
	}

	r.ctx.Spinner.Stop(true)
	return nil
}

// printSummary prints a brief completion message.
func (r *Runner) printSummary(duration time.Duration) {
	fmt.Fprintln(os.Stderr)
	if r.ctx.IsDryRun {
		fmt.Fprintf(os.Stderr, "%s Done %s\n", ui.FormatSuccess(ui.IconSuccess), ui.FormatDim("(dry-run, no changes made)"))
	} else {
		fmt.Fprintf(os.Stderr, "%s Done in %.1fs\n", ui.FormatSuccess(ui.IconSuccess), duration.Seconds())
	}
}

// renderTagName replaces ${version} in a template string.
func renderTagName(template string, version string) string {
	result := template
	result = replaceAll(result, "${version}", version)
	return result
}

// replaceAll is a simple string replacement helper.
func replaceAll(s, old, new string) string {
	for {
		i := indexOf(s, old)
		if i < 0 {
			return s
		}
		s = s[:i] + new + s[i+len(old):]
	}
}

// indexOf finds the first occurrence of substr in s.
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// hasPrefix checks if s starts with prefix.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// latestVersionToTag converts LatestVersion to a git tag for commit range queries.
// Returns empty string for initial release (0.0.0) since no tag exists yet,
// which causes GetCommitsSinceTag to return all commits.
func latestVersionToTag(latestVersion string) string {
	if latestVersion == "" || latestVersion == "0.0.0" {
		return ""
	}
	if !hasPrefix(latestVersion, "v") {
		return "v" + latestVersion
	}
	return latestVersion
}
