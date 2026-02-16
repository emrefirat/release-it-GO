package cli

import (
	"fmt"

	"release-it-go/internal/config"
	"release-it-go/internal/ui"

	"github.com/spf13/cobra"
)

// newInitCommand creates the "init" subcommand for interactive config setup.
func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new release-it-go configuration",
		Long: `Interactively create a .release-it-go.json configuration file.
If a legacy .release-it.json file is found, it can be migrated automatically.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	var prompter ui.Prompter
	if ciMode {
		prompter = &ui.NonInteractivePrompter{}
	} else {
		prompter = &ui.InteractivePrompter{}
	}

	return runInitWithPrompter(prompter)
}

// runInitWithPrompter runs the init wizard with the given prompter (testable).
func runInitWithPrompter(prompter ui.Prompter) error {
	// Check for existing native config
	if config.DetectNativeConfig() {
		overwrite, err := prompter.Confirm(
			fmt.Sprintf("%s already exists. Overwrite?", config.NativeConfigFile),
			false,
		)
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Check for legacy config
	if legacyPath, found := config.DetectLegacyConfig(); found {
		migrate, err := prompter.Confirm(
			fmt.Sprintf("Found %s. Migrate to %s?", legacyPath, config.NativeConfigFile),
			true,
		)
		if err != nil {
			return err
		}
		if migrate {
			if err := config.MigrateLegacyConfig(legacyPath); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
			fmt.Printf("Migrated %s → %s (backup: %s.bak)\n", legacyPath, config.NativeConfigFile, legacyPath)
			return nil
		}
	}

	// Run wizard
	cfg := config.DefaultConfig()

	// Platform selection
	platformIdx, err := prompter.Select("Select platform:", []string{
		"GitHub",
		"GitLab",
		"Git tag only (no releases)",
	}, 0)
	if err != nil {
		return err
	}

	switch platformIdx {
	case 0: // GitHub
		cfg.GitHub.Release = true
	case 1: // GitLab
		cfg.GitLab.Release = true
	case 2: // Git tag only
		// defaults: no release
	}

	// Changelog format
	changelogIdx, err := prompter.Select("Changelog format:", []string{
		"Conventional Changelog",
		"Keep a Changelog",
		"None",
	}, 0)
	if err != nil {
		return err
	}

	switch changelogIdx {
	case 0: // Conventional Changelog
		cfg.Changelog.Enabled = true
		cfg.Changelog.Preset = "angular"
	case 1: // Keep a Changelog
		cfg.Changelog.Enabled = true
		cfg.Changelog.KeepAChangelog = true
	case 2: // None
		cfg.Changelog.Enabled = false
	}

	// Git commit and tag
	commitTag, err := prompter.Confirm("Enable git commit and tag?", true)
	if err != nil {
		return err
	}
	cfg.Git.Commit = commitTag
	cfg.Git.Tag = commitTag

	// Git push (separate from commit/tag)
	pushEnabled, err := prompter.Confirm("Enable git push?", true)
	if err != nil {
		return err
	}
	cfg.Git.Push = pushEnabled

	// When push is disabled, upstream check is irrelevant
	if !pushEnabled {
		cfg.Git.RequireUpstream = false
	}

	// Require commits before release
	requireCommits, err := prompter.Confirm("Require new commits before release?", true)
	if err != nil {
		return err
	}
	cfg.Git.RequireCommits = requireCommits

	// Commit message template
	commitMsg, err := prompter.Input(
		"Commit message template",
		"chore(release): release v${version}",
	)
	if err != nil {
		return err
	}
	cfg.Git.CommitMessage = commitMsg

	// Tag format
	tagName, err := prompter.Input("Tag name format", "v${version}")
	if err != nil {
		return err
	}
	cfg.Git.TagName = tagName

	// Require conventional commits
	requireConventional, err := prompter.Confirm("Require conventional commits?", true)
	if err != nil {
		return err
	}
	cfg.Git.RequireConventionalCommits = requireConventional

	// requireBranch
	requireBranch, err := prompter.Input("Required branch (empty to skip)", "main")
	if err != nil {
		return err
	}
	cfg.Git.RequireBranch = requireBranch

	// Write config
	if err := config.WriteConfigJSON(cfg, config.NativeConfigFile); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("Created %s\n", config.NativeConfigFile)
	return nil
}
