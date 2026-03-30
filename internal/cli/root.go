// Package cli provides the command-line interface for release-it-go.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
	"release-it-go/internal/runner"
)

// Build information, set via ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// CLI flag variables
var (
	cfgFile          string
	dryRun           bool
	ciMode           bool
	verboseCount     int
	increment        string
	preReleaseID     string
	preRelease       string
	showChangelog    bool
	releaseVersion   bool
	onlyVersion      bool
	noIncrement      bool
	noCommit         bool
	noTag            bool
	noPush           bool
	checkCommits     bool
	ignoreCommitLint bool
)

// NewRootCommand creates the root cobra command for release-it-go.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "release-it-go",
		Short: "Release automation tool for Git projects",
		Long: `release-it-go is a release automation tool that handles
Git tagging, changelog generation, and GitHub/GitLab releases.
It is a Go reimplementation of release-it without Node.js dependencies.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runRelease,
	}

	// Persistent flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "dry-run mode (no changes)")
	rootCmd.PersistentFlags().BoolVar(&ciMode, "ci", false, "CI mode (non-interactive)")
	rootCmd.PersistentFlags().CountVarP(&verboseCount, "verbose", "V", "verbose output (-V for verbose, -VV for debug)")
	rootCmd.PersistentFlags().StringVarP(&increment, "increment", "i", "", "version increment type (major/minor/patch/pre*)")
	rootCmd.PersistentFlags().StringVar(&preReleaseID, "preReleaseId", "", "pre-release identifier (e.g., beta, alpha)")
	rootCmd.PersistentFlags().StringVar(&preRelease, "preRelease", "", "shorthand for pre-release (sets preReleaseId and marks release as pre-release)")

	// Mode flags
	rootCmd.Flags().BoolVar(&showChangelog, "changelog", false, "show changelog only")
	rootCmd.Flags().BoolVar(&releaseVersion, "release-version", false, "show next version only")
	rootCmd.Flags().BoolVar(&onlyVersion, "only-version", false, "prompt for version only")
	rootCmd.Flags().BoolVar(&noIncrement, "no-increment", false, "skip version increment")

	// Disable flags
	rootCmd.Flags().BoolVar(&noCommit, "no-git.commit", false, "skip git commit")
	rootCmd.Flags().BoolVar(&noTag, "no-git.tag", false, "skip git tag")
	rootCmd.Flags().BoolVar(&noPush, "no-git.push", false, "skip git push")

	// Commit lint flags
	rootCmd.Flags().BoolVar(&checkCommits, "check-commits", false, "check commit conventions only (no release)")
	rootCmd.Flags().BoolVar(&ignoreCommitLint, "ignore-commit-lint", false, "skip conventional commit validation")

	// Subcommands
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newCompletionCommand())
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newHooksCommand())

	return rootCmd
}

// newVersionCommand creates the "version" subcommand.
func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of release-it-go",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("release-it-go %s (commit: %s, built: %s)\n", Version, Commit, Date)
		},
	}
}

// newCompletionCommand creates the "completion" subcommand for shell completion generation.
func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for your shell.

To load completions:

Bash:
  $ source <(release-it-go completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ release-it-go completion bash > /etc/bash_completion.d/release-it-go
  # macOS:
  $ release-it-go completion bash > $(brew --prefix)/etc/bash_completion.d/release-it-go

Zsh:
  $ release-it-go completion zsh > "${fpath[1]}/_release-it-go"

Fish:
  $ release-it-go completion fish > ~/.config/fish/completions/release-it-go.fish

PowerShell:
  PS> release-it-go completion powershell | Out-String | Invoke-Expression
`,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(out)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(out)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}

// runRelease is the main entry point for the release command.
func runRelease(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Expand --preRelease shorthand into preReleaseId + github/gitlab preRelease
	if preRelease != "" && preReleaseID == "" {
		preReleaseID = preRelease
	}

	// Apply CLI flag overrides
	config.ApplyFlags(cfg, config.FlagOverrides{
		DryRun:       &dryRun,
		CI:           &ciMode,
		Verbose:      &verboseCount,
		Increment:    &increment,
		PreReleaseID: &preReleaseID,
		NoCommit:     &noCommit,
		NoTag:        &noTag,
		NoPush:       &noPush,
	})

	// When preRelease is set, auto-mark GitHub/GitLab releases as pre-release
	if preRelease != "" {
		cfg.GitHub.PreRelease = true
		cfg.GitLab.PreRelease = true
	}

	// Create logger
	logger := applog.NewLogger(cfg.Verbose, cfg.DryRun)

	if cfg.DryRun {
		logger.DryRun("running in dry-run mode")
	}

	if cfg.ConfigFile != "" {
		logger.Debug("config loaded from %s", cfg.ConfigFile)
	} else {
		logger.Print("  %s No config file found, using defaults", "⚠")
	}

	// Handle no-increment flag
	if noIncrement {
		cfg.Increment = "no-increment"
	}

	// Validate CalVer + SemVer conflict
	if cfg.CalVer.Enabled && cfg.PreReleaseID != "" {
		return fmt.Errorf("CalVer and pre-release cannot be used together")
	}

	// Handle commit lint override
	if ignoreCommitLint {
		cfg.Git.RequireConventionalCommits = false
	}

	// Create runner and handle special modes
	r := runner.NewRunner(cfg)

	if checkCommits {
		return r.RunCheckCommits()
	}

	if showChangelog {
		return r.RunChangelogOnly()
	}

	if releaseVersion {
		return r.RunReleaseVersionOnly()
	}

	if onlyVersion {
		return r.RunOnlyVersion()
	}

	if cfg.Increment == "no-increment" {
		return r.RunNoIncrement()
	}

	// Main release pipeline
	return r.Run()
}

// Execute runs the root command.
func Execute() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
