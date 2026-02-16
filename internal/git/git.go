// Package git provides git command execution and repository operations
// for release-it-go. All operations use the git CLI binary directly
// to preserve user's git config (gpg signing, hooks, credentials).
package git

import (
	"fmt"
	"os/exec"
	"strings"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
)

// commandExecutor runs a command and returns its output.
// This is a function variable to allow mocking in tests.
var commandExecutor = func(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// readOnlyCommands lists git subcommands that don't modify state.
var readOnlyCommands = map[string]bool{
	"status":    true,
	"log":       true,
	"describe":  true,
	"rev-parse": true,
	"remote":    true,
	"diff":      true,
	"show":      true,
	"branch":    true,
	"tag":       false, // tag without -l modifies
}

// Git provides git operations backed by the git CLI.
type Git struct {
	config *config.GitConfig
	logger *applog.Logger
	dryRun bool
}

// NewGit creates a new Git instance.
func NewGit(cfg *config.GitConfig, logger *applog.Logger, dryRun bool) *Git {
	return &Git{
		config: cfg,
		logger: logger,
		dryRun: dryRun,
	}
}

// run executes a git command. Write operations are skipped in dry-run mode.
func (g *Git) run(args ...string) (string, error) {
	cmdStr := fmt.Sprintf("git %s", strings.Join(args, " "))

	if g.dryRun && isWriteOperation(args) {
		g.logger.DryRun(cmdStr)
		return "", nil
	}

	g.logger.Debug("exec: %s", cmdStr)
	out, err := commandExecutor("git", args...)
	if err != nil {
		return out, fmt.Errorf("%s: %s", cmdStr, out)
	}
	return out, nil
}

// runSilent executes a git command without logging the output.
func (g *Git) runSilent(args ...string) (string, error) {
	out, err := commandExecutor("git", args...)
	if err != nil {
		return out, err
	}
	return strings.TrimSpace(out), nil
}

// isWriteOperation determines if the git command modifies state.
func isWriteOperation(args []string) bool {
	if len(args) == 0 {
		return false
	}

	subcmd := args[0]

	// Special case: "tag -l" is read-only
	if subcmd == "tag" {
		for _, a := range args[1:] {
			if a == "-l" || a == "--list" {
				return false
			}
		}
		return true
	}

	isReadOnly, known := readOnlyCommands[subcmd]
	if !known {
		return true // unknown commands treated as write
	}
	return !isReadOnly
}

// IsGitInstalled checks if git is available in PATH.
func IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsGitRepo checks if the current directory is a git repository.
func IsGitRepo() bool {
	out, err := commandExecutor("git", "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}
