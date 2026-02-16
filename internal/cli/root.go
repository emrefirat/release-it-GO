// Package cli provides the command-line interface for release-it-go.
package cli

import (
	"fmt"
	"os"

	"github.com/emfi/release-it-go/internal/config"
	applog "github.com/emfi/release-it-go/internal/log"
	"github.com/emfi/release-it-go/internal/runner"
	"github.com/spf13/cobra"
)

// Build information, set via ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// CLI flag variables
var (
	cfgFile        string
	dryRun         bool
	ciMode         bool
	verboseCount   int
	increment      string
	preReleaseID   string
	showChangelog  bool
	releaseVersion bool
	onlyVersion    bool
	noIncrement    bool
	noCommit       bool
	noTag          bool
	noPush         bool
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

	// Mode flags
	rootCmd.Flags().BoolVar(&showChangelog, "changelog", false, "show changelog only")
	rootCmd.Flags().BoolVar(&releaseVersion, "release-version", false, "show next version only")
	rootCmd.Flags().BoolVar(&onlyVersion, "only-version", false, "prompt for version only")
	rootCmd.Flags().BoolVar(&noIncrement, "no-increment", false, "skip version increment")

	// Disable flags
	rootCmd.Flags().BoolVar(&noCommit, "no-git.commit", false, "skip git commit")
	rootCmd.Flags().BoolVar(&noTag, "no-git.tag", false, "skip git tag")
	rootCmd.Flags().BoolVar(&noPush, "no-git.push", false, "skip git push")

	// Subcommands
	rootCmd.AddCommand(newVersionCommand())

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

// runRelease is the main entry point for the release command.
func runRelease(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
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

	// Create logger
	logger := applog.NewLogger(cfg.Verbose, cfg.DryRun)

	if cfg.DryRun {
		logger.DryRun("running in dry-run mode")
	}

	logger.Debug("config loaded successfully")

	// Handle no-increment flag
	if noIncrement {
		cfg.Increment = "no-increment"
	}

	// Validate CalVer + SemVer conflict
	if cfg.CalVer.Enabled && cfg.PreReleaseID != "" {
		return fmt.Errorf("CalVer and pre-release cannot be used together")
	}

	// Create runner and handle special modes
	r := runner.NewRunner(cfg)

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
