package runner

import (
	"errors"
	"fmt"
	"os"
	"regexp"
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
	return r.runPipeline(r.determineVersion)
}

// runPipeline executes the standard pipeline with the given version step.
// Every entry point goes through here so the prerequisite and token checks
// can never be bypassed: --only-version and --no-increment previously
// started at bump/changelog, so a dirty tree, wrong branch or missing token
// surfaced only after commit, tag and push had already happened.
func (r *Runner) runPipeline(versionStep func() error) error {
	start := time.Now()
	r.printBanner()

	steps := []pipelineStep{
		{"init", r.init},
		{"prerequisites", r.checkPrerequisites},
		{"commitlint", r.checkCommitLint},
		{"version", versionStep},
		{"bump", r.bumpFiles},
		{"changelog", r.generateChangelog},
		{"git:release", r.gitRelease},
		{"github:release", r.githubRelease},
		{"gitlab:release", r.gitlabRelease},
		{"notification", r.sendNotification},
	}

	if err := r.runSteps(steps); err != nil {
		return err
	}
	if r.ctx.noCommits {
		return nil
	}

	r.printSummary(time.Since(start))
	return nil
}

// runSteps executes pipeline steps in order, firing before:/after: hooks for
// each. The aggregate release-spanning events match npm release-it semantics:
// before:release fires ahead of the git:release step (before its own
// before:git:release hook) and after:release fires once every step has
// completed. On the graceful "no commits" abort, remaining hooks — including
// after:release — are intentionally skipped: the release did not happen.
func (r *Runner) runSteps(steps []pipelineStep) error {
	for _, step := range steps {
		if step.name == "git:release" {
			if err := r.ctx.HookRunner.RunHooks("before:release"); err != nil {
				return fmt.Errorf("before:release hook: %w", err)
			}
		}

		if err := r.ctx.HookRunner.RunHooks("before:" + step.name); err != nil {
			return fmt.Errorf("before:%s hook: %w", step.name, err)
		}

		if err := step.fn(); err != nil {
			return fmt.Errorf("%s: %w", step.name, err)
		}

		// Early exit if no commits to release.
		// After-hooks are intentionally skipped: the release is aborted,
		// so post-step hooks (e.g., notifications) should not execute.
		if r.ctx.noCommits {
			return nil
		}

		r.ctx.UpdateVars()

		if err := r.ctx.HookRunner.RunHooks("after:" + step.name); err != nil {
			return fmt.Errorf("after:%s hook: %w", step.name, err)
		}
	}

	if err := r.ctx.HookRunner.RunHooks("after:release"); err != nil {
		return fmt.Errorf("after:release hook: %w", err)
	}
	return nil
}

// RunChangelogOnly generates and prints the changelog without performing a release.
// Scripting mode: never prompts, even in a TTY.
func (r *Runner) RunChangelogOnly() error {
	r.ctx.IsCI = true
	if err := r.init(); err != nil {
		return err
	}
	if err := r.determineVersion(); err != nil {
		return err
	}

	rawCommits, err := r.commitsSinceLatestRelease()
	if err != nil {
		return fmt.Errorf("getting commits: %w", err)
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
// Scripting mode: never prompts, even in a TTY.
func (r *Runner) RunReleaseVersionOnly() error {
	r.ctx.IsCI = true
	if err := r.init(); err != nil {
		return err
	}
	if err := r.determineVersion(); err != nil {
		return err
	}
	fmt.Println(r.ctx.Version)
	return nil
}

// RunOnlyVersion prompts for the version interactively, then completes the
// release without further prompts. Same safety checks as Run().
func (r *Runner) RunOnlyVersion() error {
	return r.runPipeline(func() error {
		if err := r.determineVersion(); err != nil {
			return err
		}
		r.ctx.IsCI = true // the rest of the pipeline runs without prompts
		return nil
	})
}

// RunNoIncrement re-runs the release steps for the current version — npm
// release-it's documented recovery flow after a failed push. There are no
// new commits by definition, so requireCommits is disabled; every other
// prerequisite (clean tree, branch, upstream, identity, tokens) still applies.
func (r *Runner) RunNoIncrement() error {
	r.ctx.Config.Git.RequireCommits = false
	return r.runPipeline(r.determineNoIncrementVersion)
}

// determineNoIncrementVersion keeps the latest released version as the
// release version.
func (r *Runner) determineNoIncrementVersion() error {
	latestTag, err := r.ctx.Git.GetLatestTag()
	if err != nil {
		return fmt.Errorf("getting latest tag: %w", err)
	}
	r.inferTagNameFormat(latestTag)

	parsed, parseErr := version.ParseVersion(git.VersionFromTag(latestTag, r.ctx.Config.Git.TagName))
	if parseErr != nil {
		return fmt.Errorf("parsing version %q: %w", latestTag, parseErr)
	}

	r.ctx.LatestVersion = parsed.String()
	r.ctx.Version = parsed.String()
	r.ctx.TagName = renderTagName(r.ctx.Config.Git.TagName, r.ctx.Version)
	r.ctx.Logger.Print("  %s Version: %s (no increment)", ui.IconVersion, r.ctx.Version)
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

// checkPrerequisites runs all prerequisite checks.
func (r *Runner) checkPrerequisites() error {
	r.ctx.Spinner.Start("Prerequisites checked")

	if err := r.ctx.Git.CheckPrerequisites(); err != nil {
		if errors.Is(err, git.ErrNoCommits) {
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

	latestTag := latestVersionToTag(r.ctx.LatestVersion, r.ctx.Config.Git.TagName)
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
		return formatLintError(failed, len(commitInfos), true)
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

	// Verbose: show all checked commits with their status. The error below
	// then carries only the summary — repeating every failure was noise.
	verbose := r.ctx.Logger.GetVerbose() >= 1
	if verbose {
		for _, p := range passed {
			r.ctx.Logger.Print("  %s %s %s", ui.FormatSuccess(ui.IconSuccess), shortHash(p.Hash), p.Subject)
		}
		for _, f := range failed {
			r.ctx.Logger.Print("  %s %s %s ← %s", ui.FormatError(ui.IconFail), shortHash(f.Hash), f.Subject, lintReason(f))
		}
		fmt.Fprintln(os.Stderr)
	}

	if len(failed) == 0 {
		fmt.Printf("All %d commits are conventional. %s\n", len(passed), ui.IconSuccess)
		return nil
	}

	return formatLintError(failed, len(commitInfos), !verbose)
}

// formatLintError builds the error for failed commit lints. listDetails
// controls whether each failing commit is repeated in the error (false when
// the caller already printed the per-commit list).
func formatLintError(failed []changelog.LintResult, total int, listDetails bool) error {
	var b strings.Builder
	b.WriteString("Commit lint failed:\n")
	if listDetails {
		for _, f := range failed {
			fmt.Fprintf(&b, "  %-10s %-40s ← %s\n", f.Hash, f.Subject, lintReason(f))
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "  %d of %d commits are not conventional.\n", len(failed), total)
	b.WriteString("  Use --ignore-commit-lint to bypass.\n")
	return fmt.Errorf("%s", b.String())
}

// lintReason renders a failure reason with its type suggestion, if any.
func lintReason(f changelog.LintResult) string {
	if f.Suggestion != "" {
		return fmt.Sprintf("%s (did you mean %s?)", f.Reason, f.Suggestion)
	}
	return f.Reason
}

// shortHash abbreviates a commit hash for display.
func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

// inferTagNameFormat replicates npm release-it's tag-prefix inference: with
// the shipped default tagName template, a repo whose latest tag is v-prefixed
// keeps the v prefix for new tags. Without this, v-repos got unprefixed new
// tags AND conventional-commit auto-increment silently degraded to patch
// (the rendered unprefixed tag doesn't exist, so commit queries failed).
// A tagName written explicitly in the config file is always respected.
func (r *Runner) inferTagNameFormat(rawLatestTag string) {
	if r.ctx.Config.Git.TagNameExplicit || r.ctx.Config.Git.TagName != "${version}" {
		return
	}
	if !strings.HasPrefix(rawLatestTag, "v") {
		return
	}
	if _, err := version.ParseVersion(rawLatestTag); err != nil {
		return
	}
	r.ctx.Config.Git.TagName = "v${version}"
	r.ctx.Logger.Verbose("Inferred v-prefixed tagName from latest tag %s", rawLatestTag)
}

// releaseTemplateVarPattern matches ${var} placeholders (dotted keys allowed).
var releaseTemplateVarPattern = regexp.MustCompile(`\$\{([a-zA-Z][a-zA-Z0-9.]*)\}`)

// renderReleaseTemplate renders ${...} placeholders in user-facing templates
// (commit message, tag annotation, release names) using the same variable set
// available to lifecycle hooks (${branchName}, ${latestVersion}, ${repo.*},
// ...) plus the freshly computed ${version}. Unknown placeholders are kept
// literal. Single-pass, so substituted values are never re-substituted.
func (r *Runner) renderReleaseTemplate(tmpl string) string {
	return releaseTemplateVarPattern.ReplaceAllStringFunc(tmpl, func(m string) string {
		key := m[2 : len(m)-1]
		if key == "version" {
			return r.ctx.Version
		}
		if v, ok := r.ctx.Vars[key]; ok {
			return v
		}
		return m
	})
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
	} else {
		r.inferTagNameFormat(latestTag)
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

	// Strip the tagName template's literal prefix/suffix (release-${version})
	// before parsing — otherwise any template beyond ${version}/v${version}
	// failed with "invalid version" on the second release.
	latestVersion := git.VersionFromTag(latestTag, r.ctx.Config.Git.TagName)
	if parsed, parseErr := version.ParseVersion(latestVersion); parseErr == nil {
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
	// Determine increment type. An explicit increment (from the positional
	// argument, -i, or the config file) is used as-is and never prompts.
	increment := r.ctx.Config.Increment
	explicitIncrement := increment != ""

	parsedCurrent, parseErr := version.ParseVersion(latestVersion)
	if parseErr != nil {
		return fmt.Errorf("parsing current version %q: %w", latestVersion, parseErr)
	}

	// Explicit target version (release-it-go 1.5.0 / -i 1.5.0): use it
	// verbatim, matching npm release-it — which also requires it to be
	// greater than the latest version.
	if explicitIncrement && !version.IsIncrementType(increment) {
		if target, err := version.ParseVersion(increment); err == nil {
			if !target.GreaterThan(parsedCurrent) {
				return fmt.Errorf("explicit version %s must be greater than the latest version %s", target, parsedCurrent)
			}
			newVersionStr := target.String()
			r.ctx.Version = newVersionStr
			r.ctx.TagName = renderTagName(r.ctx.Config.Git.TagName, newVersionStr)
			r.ctx.Logger.Print("  %s Version: %s → %s", ui.IconVersion, latestVersion, newVersionStr)
			return nil
		}
	}

	if increment == "" {
		increment = r.autoDetectIncrement()
	}

	if increment == "" {
		increment = "patch"
	}

	preReleaseID := r.ctx.Config.PreReleaseID
	incrementType := increment
	if preReleaseID != "" {
		switch {
		case strings.HasPrefix(increment, "pre"):
			// Explicit pre* keyword — the same words the interactive menu
			// offers; prefixing again produced "prepreminor".
			incrementType = increment
		case parsedCurrent.Prerelease() != "" && strings.HasPrefix(parsedCurrent.Prerelease(), preReleaseID+"."):
			// Same series: bump the pre-release number (beta.0 → beta.1)
			incrementType = "prerelease"
		default:
			// Start a new pre-release series
			incrementType = "pre" + increment
		}
	}

	newSemver, err := version.IncrementVersion(parsedCurrent, incrementType, preReleaseID)
	if err != nil {
		return fmt.Errorf("incrementing version: %w", err)
	}
	newVersionStr := newSemver.String()

	// Interactive mode: let user choose — but only when the increment was
	// auto-detected. An explicit increment (positional/-i/config) never
	// prompts; the old value-equality check re-ran auto-detection and fired
	// whenever an explicit choice happened to coincide with it.
	if !r.ctx.IsCI && !explicitIncrement {
		options := r.buildVersionOptions(latestVersion, incrementType)
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
	updated, err := b.WriteVersionFiles(r.ctx.LatestVersion, r.ctx.Version)
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("bumping version files: %w", err)
	}
	r.ctx.BumpedFiles = updated

	r.ctx.Spinner.Stop(true)
	return nil
}

// commitsSinceLatestRelease returns raw commits (hash + full message) since
// the latest release tag, ready for the conventional-commit parser. Full
// messages matter: BREAKING CHANGE footers live in the body, which
// subject-only fetching silently dropped. Falls back to the raw git tag when
// the rendered tag name doesn't exist (tag format transitions, e.g. old
// "v1.1.0" tags with a new "${version}" template).
func (r *Runner) commitsSinceLatestRelease() ([]changelog.RawCommit, error) {
	latestTag := latestVersionToTag(r.ctx.LatestVersion, r.ctx.Config.Git.TagName)

	commits, err := r.ctx.Git.GetFullCommitsSinceTag(latestTag)
	if err != nil && latestTag != "" {
		rawTag, rawErr := r.ctx.Git.GetLatestTag()
		switch {
		case rawErr != nil:
			// No tags at all (the version came from bumper.in): every commit
			// belongs to this release.
			r.ctx.Logger.Debug("tag %q not found and no tags exist, using all commits", latestTag)
			commits, err = r.ctx.Git.GetFullCommitsSinceTag("")
		case rawTag != latestTag:
			r.ctx.Logger.Debug("tag %q not found, falling back to %q", latestTag, rawTag)
			commits, err = r.ctx.Git.GetFullCommitsSinceTag(rawTag)
		}
	}
	if err != nil {
		return nil, err
	}

	rawCommits := make([]changelog.RawCommit, len(commits))
	for i, c := range commits {
		rawCommits[i] = changelog.RawCommit{Hash: c.Hash, Message: c.Message}
	}
	return rawCommits, nil
}

// autoDetectIncrement uses conventional commits to determine the bump type.
func (r *Runner) autoDetectIncrement() string {
	rawCommits, err := r.commitsSinceLatestRelease()
	if err != nil || len(rawCommits) == 0 {
		return "patch"
	}

	parsed := changelog.ParseCommits(rawCommits)
	bump := changelog.AnalyzeBump(parsed)

	if bump == changelog.BumpNone {
		return "patch"
	}
	return bump.String()
}

// buildVersionOptions creates version options for the interactive prompt.
// With --preRelease, the menu offers pre-release variants (prepatch/preminor/
// premajor, plus continuing the current series) so the identifier is never
// silently dropped by picking a plain patch/minor/major.
func (r *Runner) buildVersionOptions(current string, recommended string) []ui.VersionOption {
	options := make([]ui.VersionOption, 0, 4)

	parsedCurrent, err := version.ParseVersion(current)
	if err != nil {
		return options
	}

	preReleaseID := r.ctx.Config.PreReleaseID
	increments := []string{"patch", "minor", "major"}
	if preReleaseID != "" {
		increments = []string{"prepatch", "preminor", "premajor"}
		if parsedCurrent.Prerelease() != "" && strings.HasPrefix(parsedCurrent.Prerelease(), preReleaseID+".") {
			// Continue the current pre-release series (beta.1 → beta.2) first
			increments = append([]string{"prerelease"}, increments...)
		}
	}

	for _, inc := range increments {
		ver, err := version.IncrementVersion(parsedCurrent, inc, preReleaseID)
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

	rawCommits, err := r.commitsSinceLatestRelease()
	if err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("getting commits: %w", err)
	}

	if len(rawCommits) == 0 {
		// Nothing new since the last release (e.g. a --no-increment recovery
		// run). Writing an empty section would corrupt CHANGELOG.md and the
		// resulting commit would move HEAD away from the existing tag.
		r.ctx.Changelog = ""
		r.ctx.Spinner.Update("Changelog unchanged (no new commits)")
		r.ctx.Spinner.Stop(true)
		return nil
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
		if err := r.stageReleaseFiles(); err != nil {
			r.ctx.Spinner.Stop(false)
			return fmt.Errorf("staging: %w", err)
		}
		r.ctx.Spinner.Stop(true)

		// Commit — skip if nothing staged (e.g., changelog disabled, no bumper)
		if !r.ctx.Git.HasStagedChanges() {
			r.ctx.Logger.Verbose("    ↳ No staged changes, skipping commit")
		} else {
			commitMsg := r.renderReleaseTemplate(cfg.CommitMessage)
			// npm asks commit/tag/push independently: declining one
			// operation must not cancel the ones after it.
			doCommit, err := r.confirmStep(fmt.Sprintf("Commit (%s)?", commitMsg), "commit")
			if err != nil {
				return err
			}
			if doCommit {
				r.ctx.Spinner.Start("Committed")
				if err := r.ctx.Git.Commit(commitMsg); err != nil {
					r.ctx.Spinner.Stop(false)
					return fmt.Errorf("commit: %w", err)
				}
				r.ctx.Spinner.Stop(true)
			}
		}
	}

	// Tag
	if cfg.Tag {
		doTag, err := r.confirmStep(fmt.Sprintf("Tag (%s)?", r.ctx.TagName), "tag")
		if err != nil {
			return err
		}
		if doTag {
			if err := r.createReleaseTag(); err != nil {
				return err
			}
		}
	}

	// Push
	if cfg.Push {
		doPush, err := r.confirmStep("Push?", "push")
		if err != nil {
			return err
		}
		if doPush {
			r.ctx.Spinner.Start("Pushed to remote")
			if err := r.ctx.Git.Push(); err != nil {
				r.ctx.Spinner.Stop(false)
				return fmt.Errorf("push: %w\n"+
					"  The release commit/tag may already exist locally. After fixing the push\n"+
					"  issue (e.g. git pull --rebase), re-run the remaining release steps with:\n"+
					"  release-it-go --no-increment", err)
			}
			r.ctx.Spinner.Stop(true)
		}
	}

	return nil
}

// stageReleaseFiles stages what this release actually changed: the bumper
// outputs recorded in BumpedFiles (the changelog stages itself right after it
// is written). addUntrackedFiles keeps the legacy whole-tree sweep for users
// who want untracked files in the release commit — otherwise unrelated local
// edits were silently swept into the "chore: release" commit whenever the
// clean-working-dir check was disabled.
func (r *Runner) stageReleaseFiles() error {
	if r.ctx.Config.Git.AddUntrackedFiles {
		return r.ctx.Git.Stage()
	}
	for _, f := range r.ctx.BumpedFiles {
		if err := r.ctx.Git.StageFile(f); err != nil {
			return fmt.Errorf("staging %s: %w", f, err)
		}
	}
	return nil
}

// confirmStep asks for interactive confirmation of one git operation.
// In CI mode it always confirms. A declined prompt skips only this operation
// (logged), never the operations after it; a prompter error (e.g. Ctrl+C)
// aborts the release.
func (r *Runner) confirmStep(question string, name string) (bool, error) {
	if r.ctx.IsCI {
		return true, nil
	}
	confirmed, err := r.ctx.Prompter.Confirm(question, true)
	if err != nil {
		return false, err
	}
	if !confirmed {
		r.ctx.Logger.Print("  %s Skipped %s", ui.IconSkip, name)
		return false, nil
	}
	return true, nil
}

// createReleaseTag creates the release tag, tolerating a tag that already
// exists at HEAD — npm release-it's documented recovery flow (--no-increment
// after a failed push) re-runs the release steps for the current version, and
// the tag from the previous attempt must not be fatal.
func (r *Runner) createReleaseTag() error {
	exists, existsErr := r.ctx.Git.TagExists(r.ctx.TagName)
	if existsErr == nil && exists {
		atHead, headErr := r.ctx.Git.TagPointsAtHead(r.ctx.TagName)
		if headErr == nil && atHead {
			r.ctx.Logger.Print("  %s Tag %s already exists at HEAD — skipping tag creation", ui.IconSkip, r.ctx.TagName)
			return nil
		}
		return fmt.Errorf("tag: tag %s already exists on a different commit (delete it or choose another version)", r.ctx.TagName)
	}

	annotation := r.renderReleaseTemplate(r.ctx.Config.Git.TagAnnotation)
	r.ctx.Spinner.Start(fmt.Sprintf("Tagged %s", r.ctx.TagName))
	if err := r.ctx.Git.CreateTag(r.ctx.TagName, annotation); err != nil {
		r.ctx.Spinner.Stop(false)
		return fmt.Errorf("tag: %w", err)
	}
	r.ctx.Spinner.Stop(true)
	return nil
}

// githubRelease creates a GitHub release.
func (r *Runner) githubRelease() error {
	if !r.ctx.Config.GitHub.Release || r.ctx.RepoInfo == nil {
		return nil
	}

	if !r.ctx.IsCI {
		releaseName := r.renderReleaseTemplate(r.ctx.Config.GitHub.ReleaseName)
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

	releaseName := r.renderReleaseTemplate(r.ctx.Config.GitHub.ReleaseName)
	releaseNotes := r.ctx.Changelog
	if releaseNotes == "" {
		r.ctx.Logger.Verbose("    ↳ Release notes are empty (changelog disabled or no commits parsed)")
	}

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
		releaseName := r.renderReleaseTemplate(r.ctx.Config.GitLab.ReleaseName)
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

	releaseName := r.renderReleaseTemplate(r.ctx.Config.GitLab.ReleaseName)
	releaseNotes := r.ctx.Changelog
	if releaseNotes == "" {
		r.ctx.Logger.Verbose("    ↳ Release notes are empty (changelog disabled or no commits parsed)")
	}

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

	// Build rich context for structured notifications (Teams MessageCard)
	richCtx := &notification.RichNotificationContext{
		Version:       r.ctx.Version,
		LatestVersion: r.ctx.LatestVersion,
		TagName:       r.ctx.TagName,
		Changelog:     r.ctx.Changelog,
		ReleaseURL:    r.ctx.ReleaseURL,
		BranchName:    r.ctx.BranchName,
	}
	if r.ctx.RepoInfo != nil {
		richCtx.RepoHost = r.ctx.RepoInfo.Host
		richCtx.RepoOwner = r.ctx.RepoInfo.Owner
		richCtx.RepoName = r.ctx.RepoInfo.Repository
	}

	// Apply webhook-level config (themeColor, imageUrl, ignoredContributors)
	var ignoredContributors []string
	for _, wh := range cfg.Webhooks {
		if wh.ThemeColor != "" && richCtx.ThemeColor == "" {
			richCtx.ThemeColor = wh.ThemeColor
		}
		if wh.ImageURL != "" && richCtx.ImageURL == "" {
			richCtx.ImageURL = wh.ImageURL
		}
		if len(wh.IgnoredContributors) > 0 {
			ignoredContributors = append(ignoredContributors, wh.IgnoredContributors...)
		}
	}

	// Gather commit count and contributors
	latestTag := latestVersionToTag(r.ctx.LatestVersion, r.ctx.Config.Git.TagName)
	if count, err := r.ctx.Git.GetCommitCountSinceTag(latestTag); err == nil {
		richCtx.CommitCount = count
	}
	if contributors, err := r.ctx.Git.GetContributorsSinceTag(latestTag); err == nil {
		richCtx.Contributors = filterContributors(contributors, ignoredContributors)
	}

	client.SetRichContext(richCtx)

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

// filterContributors removes ignored contributors from the list.
func filterContributors(contributors []string, ignored []string) []string {
	if len(ignored) == 0 {
		return contributors
	}
	ignoredSet := make(map[string]bool, len(ignored))
	for _, name := range ignored {
		ignoredSet[name] = true
	}
	filtered := make([]string, 0, len(contributors))
	for _, name := range contributors {
		if !ignoredSet[name] {
			filtered = append(filtered, name)
		}
	}
	return filtered
}

// renderTagName replaces ${version} in a template string.
func renderTagName(template string, version string) string {
	return strings.ReplaceAll(template, "${version}", version)
}

// latestVersionToTag converts LatestVersion to a git tag for commit range queries.
// Uses the tagName template from config to build the correct tag name.
// Returns empty string for initial release (0.0.0) since no tag exists yet,
// which causes GetFullCommitsSinceTag to return all commits.
func latestVersionToTag(latestVersion string, tagNameTemplate string) string {
	if latestVersion == "" || latestVersion == "0.0.0" {
		return ""
	}
	// Strip "v" prefix from version before applying template to avoid "vv" duplication
	cleanVersion := strings.TrimPrefix(latestVersion, "v")
	if tagNameTemplate != "" {
		return renderTagName(tagNameTemplate, cleanVersion)
	}
	return cleanVersion
}
