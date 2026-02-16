package runner

import (
	"fmt"
	"time"

	"release-it-go/internal/bumper"
	"release-it-go/internal/changelog"
	"release-it-go/internal/config"
	"release-it-go/internal/git"
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

// Run executes the full release pipeline.
func (r *Runner) Run() error {
	start := time.Now()

	steps := []pipelineStep{
		{"init", r.init},
		{"prerequisites", r.checkPrerequisites},
		{"version", r.determineVersion},
		{"bump", r.bumpFiles},
		{"changelog", r.generateChangelog},
		{"git:release", r.gitRelease},
		{"github:release", r.githubRelease},
		{"gitlab:release", r.gitlabRelease},
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

// RunChangelogOnly generates and prints the changelog without performing a release.
func (r *Runner) RunChangelogOnly() error {
	if err := r.init(); err != nil {
		return err
	}
	if err := r.determineVersion(); err != nil {
		return err
	}

	latestTag := r.ctx.LatestVersion
	if latestTag != "" && !hasPrefix(latestTag, "v") {
		latestTag = "v" + latestTag
	}

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
	r.ctx.Spinner.Start("Initializing")

	repoInfo, err := git.GetRepoInfo("")
	if err != nil {
		r.ctx.Spinner.Stop(false)
		r.ctx.Logger.Verbose("Could not get repo info: %v", err)
		// Non-fatal: repo info is optional for local-only operations
	} else {
		r.ctx.RepoInfo = repoInfo
	}

	branchName, err := git.GetBranchName()
	if err != nil {
		r.ctx.Spinner.Stop(false)
		r.ctx.Logger.Verbose("Could not get branch name: %v", err)
	} else {
		r.ctx.BranchName = branchName
	}

	r.ctx.UpdateVars()
	r.ctx.Spinner.Stop(true)
	return nil
}

// checkPrerequisites runs all prerequisite checks.
func (r *Runner) checkPrerequisites() error {
	r.ctx.Spinner.Start("Checking prerequisites")

	if err := r.ctx.Git.CheckPrerequisites(); err != nil {
		r.ctx.Spinner.Stop(false)
		return err
	}

	r.ctx.Spinner.Stop(true)
	return nil
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
	r.ctx.Logger.Info("Version (CalVer): %s → %s", latestVersion, newVersion)
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
		incrementType = "pre" + increment
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
	r.ctx.Logger.Info("Version: %s → %s", latestVersion, newVersionStr)
	return nil
}

// bumpFiles writes the new version to configured bumper output files.
func (r *Runner) bumpFiles() error {
	if !r.ctx.Config.Bumper.Enabled || len(r.ctx.Config.Bumper.Out) == 0 {
		return nil
	}

	r.ctx.Spinner.Start("Updating version files")

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
	latestTag := r.ctx.LatestVersion
	if latestTag != "" && !hasPrefix(latestTag, "v") {
		latestTag = "v" + latestTag
	}

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

	r.ctx.Spinner.Start("Generating changelog")

	latestTag := r.ctx.LatestVersion
	if latestTag != "" && !hasPrefix(latestTag, "v") {
		latestTag = "v" + latestTag
	}

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
		r.ctx.Spinner.Start("Staging files")
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
				r.ctx.Logger.Info("Skipped commit")
				return nil
			}
		}

		r.ctx.Spinner.Start("Creating commit")
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
				r.ctx.Logger.Info("Skipped tag")
				return nil
			}
		}

		r.ctx.Spinner.Start(fmt.Sprintf("Creating tag %s", r.ctx.TagName))
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
				r.ctx.Logger.Info("Skipped push")
				return nil
			}
		}

		r.ctx.Spinner.Start("Pushing to remote")
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
			r.ctx.Logger.Info("Skipped GitHub release")
			return nil
		}
	}

	r.ctx.Spinner.Start("Creating GitHub release")

	client, err := release.NewGitHubClient(&r.ctx.Config.GitHub, r.ctx.RepoInfo, r.ctx.Logger, r.ctx.IsDryRun)
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("GitHub client: %w", err)
	}

	if err := client.ValidateToken(); err != nil {
		r.ctx.Spinner.Stop(false)
		return err
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
			r.ctx.Logger.Info("Skipped GitLab release")
			return nil
		}
	}

	r.ctx.Spinner.Start("Creating GitLab release")

	client, err := release.NewGitLabClient(&r.ctx.Config.GitLab, r.ctx.RepoInfo, r.ctx.Logger, r.ctx.IsDryRun)
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("GitLab client: %w", err)
	}

	if err := client.ValidateToken(); err != nil {
		r.ctx.Spinner.Stop(false)
		return err
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

// printSummary prints the release summary after successful completion.
func (r *Runner) printSummary(duration time.Duration) {
	fmt.Println()

	if r.ctx.IsDryRun {
		fmt.Printf("release-it-go %s (dry-run)\n\n", r.ctx.Version)
	} else {
		fmt.Printf("release-it-go %s\n\n", r.ctx.Version)
	}

	cfg := &r.ctx.Config.Git
	prefix := ""
	if r.ctx.IsDryRun {
		prefix = "[dry-run] "
	}

	if r.ctx.Changelog != "" {
		if r.ctx.IsDryRun {
			fmt.Printf("  %sChangelog: %s would be updated\n", prefix, r.ctx.Config.Changelog.Infile)
		} else {
			fmt.Printf("  Changelog: %s updated\n", r.ctx.Config.Changelog.Infile)
		}
	}

	if cfg.Commit {
		commitMsg := renderTagName(cfg.CommitMessage, r.ctx.Version)
		if r.ctx.IsDryRun {
			fmt.Printf("  %sWould commit: %s\n", prefix, commitMsg)
		} else {
			fmt.Printf("  Committed: %s\n", commitMsg)
		}
	}

	if cfg.Tag {
		if r.ctx.IsDryRun {
			fmt.Printf("  %sWould tag: %s\n", prefix, r.ctx.TagName)
		} else {
			fmt.Printf("  Tagged: %s\n", r.ctx.TagName)
		}
	}

	if cfg.Push {
		if r.ctx.IsDryRun {
			fmt.Printf("  %sWould push to: %s\n", prefix, cfg.PushRepo)
		} else {
			fmt.Printf("  Pushed: %s (%s)\n", cfg.PushRepo, r.ctx.BranchName)
		}
	}

	if r.ctx.ReleaseURL != "" && r.ctx.ReleaseURL != "(dry-run)" {
		fmt.Printf("  Release: %s\n", r.ctx.ReleaseURL)
	}

	fmt.Println()
	if r.ctx.IsDryRun {
		fmt.Println(ui.FormatDim("Done! (dry-run, no changes made)"))
	} else {
		fmt.Printf("Done! (in %.1fs)\n", duration.Seconds())
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
