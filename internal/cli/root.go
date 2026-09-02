// Package cli provides the command-line interface for release-it-go.
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"release-it-go/internal/changelog"
	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
	"release-it-go/internal/runner"
	"release-it-go/internal/ui"
	"release-it-go/internal/version"
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
	checkMsgFile     string
)

// NewRootCommand creates the root cobra command for release-it-go.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "release-it-go [increment]",
		Short: "Release automation tool for Git projects",
		Long: `release-it-go is a release automation tool that handles
Git tagging, changelog generation, and GitHub/GitLab releases.
It is a Go reimplementation of release-it without Node.js dependencies.

The optional positional argument sets the version increment
(major|minor|patch|premajor|preminor|prepatch|prerelease) or an
explicit target version (e.g. 1.5.0), matching npm release-it:

  release-it-go minor
  release-it-go 1.5.0 --ci`,
		Args:          cobra.MaximumNArgs(1),
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
	rootCmd.Flags().StringVar(&checkMsgFile, "check-msg", "", "validate a single commit message file (for commit-msg hook)")

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

// resolveIncrementArg reconciles the positional increment argument with the
// --increment flag. The positional form (release-it-go minor, release-it-go
// 1.5.0) is npm release-it's primary documented usage; it must be validated —
// silently ignoring it would run a full release with an auto-detected
// increment the user did not ask for.
func resolveIncrementArg(args []string, flagIncrement string) (string, error) {
	if len(args) == 0 {
		return flagIncrement, nil
	}

	arg := args[0]
	if !version.IsIncrementType(arg) {
		if _, err := version.ParseVersion(arg); err != nil {
			return "", fmt.Errorf("invalid argument %q: expected an increment (major|minor|patch|premajor|preminor|prepatch|prerelease) or an explicit version like 1.5.0", arg)
		}
	}

	if flagIncrement != "" && flagIncrement != arg {
		return "", fmt.Errorf("conflicting increments: positional %q vs --increment %q", arg, flagIncrement)
	}

	return arg, nil
}

// buildFlagOverrides collects CLI overrides for config.ApplyFlags. Bool and
// count flags only override the config when the user actually passed them
// (Flags().Changed) — otherwise a config-file "ci: true", "dry-run: true",
// or "verbose: N" would be clobbered by the flag defaults on every run.
// String flags and the --no-* disables carry their own "unset" sentinels
// (empty string / false), so their pointers are always passed.
func buildFlagOverrides(cmd *cobra.Command) config.FlagOverrides {
	overrides := config.FlagOverrides{
		Increment:    &increment,
		PreReleaseID: &preReleaseID,
		NoCommit:     &noCommit,
		NoTag:        &noTag,
		NoPush:       &noPush,
	}
	if cmd.Flags().Changed("dry-run") {
		overrides.DryRun = &dryRun
	}
	if cmd.Flags().Changed("ci") {
		overrides.CI = &ciMode
	}
	if cmd.Flags().Changed("verbose") {
		overrides.Verbose = &verboseCount
	}
	return overrides
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
	config.ApplyFlags(cfg, buildFlagOverrides(cmd))

	// Positional increment/version argument (release-it-go minor, release-it-go 1.5.0)
	resolvedIncrement, err := resolveIncrementArg(args, increment)
	if err != nil {
		return err
	}
	if resolvedIncrement != "" {
		cfg.Increment = resolvedIncrement
	}

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
	} else if checkMsgFile != "" || checkCommits {
		// Lint-only modes never read the config; the commit-msg hook runs
		// --check-msg on every commit, so this warning would be pure noise.
		logger.Verbose("No config file found, using defaults")
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

	// Check single commit message file (for commit-msg hook)
	if checkMsgFile != "" {
		return runCheckMsg(checkMsgFile, verboseCount > 0)
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

// runCheckMsg validates a single commit message file against conventional commit format.
// Used in commit-msg git hooks: ./release-it-go --check-msg $1
// Compact output by default, verbose (-V) shows detailed help.
// Accepts: file path, "-" for stdin, or direct message string.
func runCheckMsg(input string, verbose bool) error {
	var raw string

	switch {
	case input == "-":
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		raw = string(data)
	case fileExists(input):
		data, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("reading commit message file: %w", err)
		}
		raw = string(data)
	default:
		raw = input
	}

	subject := firstSubjectLine(raw)
	if subject == "" {
		return fmt.Errorf("commit message is empty")
	}

	_, failed := changelog.LintCommits([]changelog.LintInput{{Subject: subject}})
	if len(failed) == 0 {
		return nil
	}

	fmt.Fprint(os.Stderr, formatCheckMsgFailure(subject, failed[0], verbose))
	return ErrCheckFailed
}

// ErrCheckFailed marks a failed --check-msg run whose diagnostics were already
// printed; Execute exits non-zero without appending a redundant "Error:" line.
var ErrCheckFailed = errors.New("commit message is not conventional")

// firstSubjectLine returns the first meaningful line of a commit message:
// leading blank lines and "#" comment lines (git's COMMIT_EDITMSG template)
// are skipped, as git itself strips them after the commit-msg hook runs.
func firstSubjectLine(message string) string {
	for _, line := range strings.Split(message, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

// checkMsgLabelWidth aligns the label column of the --check-msg diagnostic.
// Labels are padded as plain text — padding a colored string counts the ANSI
// escape bytes and misaligns the columns on a real terminal.
const checkMsgLabelWidth = 10

// formatCheckMsgFailure renders the --check-msg diagnostic: what was written,
// what is wrong (with a type suggestion when one is close), what is expected,
// and an example — ideally the user's own message with the type corrected.
func formatCheckMsgFailure(subject string, f changelog.LintResult, verbose bool) string {
	var b strings.Builder
	row := func(label, value string) {
		fmt.Fprintf(&b, "  %-*s %s\n", checkMsgLabelWidth, label, value)
	}

	fmt.Fprintf(&b, "\n%s %s\n\n", ui.FormatError(ui.IconFail), ui.FormatBold("Invalid commit message"))
	row("message:", subject)
	row("problem:", ui.FormatError(describeProblem(f)))
	b.WriteString("\n")
	row("Expected:", "<type>(<scope>): <description>   "+ui.FormatDim("scope is optional"))
	row("Example:", ui.FormatSuccess(exampleFor(subject, f)))
	row("Types:", strings.Join(changelog.ValidTypeNames(), ", "))

	if verbose {
		b.WriteString("\n  Valid types:\n")
		for _, t := range changelog.ValidTypes() {
			fmt.Fprintf(&b, "    %-10s %s\n", t.Name, t.Description)
		}
		b.WriteString("\n  Rules:\n")
		b.WriteString("    · type must be lowercase (feat, not Feat)\n")
		b.WriteString("    · scope is optional: feat(auth): ...\n")
		b.WriteString("    · merge, revert, fixup!, squash! and amend! commits are always accepted\n")
	} else {
		fmt.Fprintf(&b, "\n  %s\n", ui.FormatDim("Run with -V for type descriptions and rules."))
	}
	b.WriteString("\n")
	return b.String()
}

// describeProblem turns a lint reason into an actionable sentence.
func describeProblem(f changelog.LintResult) string {
	switch {
	case strings.HasPrefix(f.Reason, "unknown type:"):
		unknown := strings.TrimSpace(strings.TrimPrefix(f.Reason, "unknown type:"))
		if f.Suggestion != "" {
			return fmt.Sprintf("unknown type %q — did you mean %q?", unknown, f.Suggestion)
		}
		return fmt.Sprintf("unknown type %q", unknown)
	case strings.Contains(f.Reason, "not in conventional commit format"):
		return `missing the "type: " prefix (or the ": " separator after the type)`
	default:
		return f.Reason
	}
}

// leadingTypePattern matches the type token at the start of a subject.
var leadingTypePattern = regexp.MustCompile(`^[A-Za-z]+`)

// exampleFor prefers the user's own message with the suggested type swapped
// in, falling back to a generic example.
func exampleFor(subject string, f changelog.LintResult) string {
	if f.Suggestion != "" && leadingTypePattern.MatchString(subject) {
		return leadingTypePattern.ReplaceAllString(subject, f.Suggestion)
	}
	return "feat(auth): add user login"
}

// reasonDescription returns a human-readable description for lint reasons.
func reasonDescription(reason string) string {
	switch {
	case strings.Contains(reason, "not in conventional commit format"):
		return "message must follow conventional commit format"
	case strings.HasPrefix(reason, "unknown type:"):
		return "type is not in the allowed list"
	default:
		return reason
	}
}

// fileExists checks if a path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// Execute runs the root command.
func Execute() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		if errors.Is(err, ErrCheckFailed) {
			os.Exit(1) // diagnostics were already printed
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
