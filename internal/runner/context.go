// Package runner provides the main release pipeline orchestrator.
// It coordinates all release steps: init, prerequisites, versioning,
// changelog, git operations, and platform releases.
package runner

import (
	"release-it-go/internal/config"
	"release-it-go/internal/git"
	"release-it-go/internal/hook"
	applog "release-it-go/internal/log"
	"release-it-go/internal/ui"
)

// ReleaseContext holds shared state throughout the release pipeline.
type ReleaseContext struct {
	Config     *config.Config
	Logger     *applog.Logger
	Git        *git.Git
	Prompter   ui.Prompter
	HookRunner *hook.HookRunner
	Spinner    *ui.Spinner

	// State populated during pipeline execution
	LatestVersion string
	Version       string
	TagName       string
	Changelog     string
	ReleaseURL    string
	RepoInfo      *git.RepoInfo
	BranchName    string
	IsDryRun      bool
	IsCI          bool

	// Template variables for hooks and config templates
	Vars map[string]string
}

// NewReleaseContext creates a new release context from configuration.
func NewReleaseContext(cfg *config.Config) *ReleaseContext {
	isCI := cfg.CI || ui.IsCI()
	logger := applog.NewLogger(cfg.Verbose, cfg.DryRun)
	hookRunner := hook.NewHookRunner(&cfg.Hooks, logger, cfg.DryRun)

	var prompter ui.Prompter
	if isCI {
		prompter = &ui.NonInteractivePrompter{}
	} else {
		prompter = &ui.InteractivePrompter{}
	}

	gitClient := git.NewGit(&cfg.Git, logger, cfg.DryRun)
	spinner := ui.NewSpinner(isCI)

	return &ReleaseContext{
		Config:     cfg,
		Logger:     logger,
		Git:        gitClient,
		Prompter:   prompter,
		HookRunner: hookRunner,
		Spinner:    spinner,
		IsDryRun:   cfg.DryRun,
		IsCI:       isCI,
		Vars:       make(map[string]string),
	}
}

// UpdateVars updates the template variables based on current state.
func (ctx *ReleaseContext) UpdateVars() {
	ctx.Vars["version"] = ctx.Version
	ctx.Vars["latestVersion"] = ctx.LatestVersion
	ctx.Vars["tagName"] = ctx.TagName
	ctx.Vars["changelog"] = ctx.Changelog
	ctx.Vars["releaseUrl"] = ctx.ReleaseURL
	ctx.Vars["branchName"] = ctx.BranchName

	if ctx.RepoInfo != nil {
		ctx.Vars["repo.remote"] = ctx.RepoInfo.Remote
		ctx.Vars["repo.protocol"] = ctx.RepoInfo.Protocol
		ctx.Vars["repo.host"] = ctx.RepoInfo.Host
		ctx.Vars["repo.owner"] = ctx.RepoInfo.Owner
		ctx.Vars["repo.repository"] = ctx.RepoInfo.Repository
	}

	ctx.HookRunner.SetVars(ctx.Vars)
}
