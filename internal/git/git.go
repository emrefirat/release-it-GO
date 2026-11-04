// Package git provides git command execution and repository operations
// for release-it-go. All operations use the git CLI binary directly
// to preserve user's git config (gpg signing, hooks, credentials).
package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/emrefirat/release-it-GO/internal/config"
	applog "github.com/emrefirat/release-it-GO/internal/log"
)

// commandExecutor runs a command and returns its trimmed stdout. On failure
// git's stderr is folded into the returned error so callers keep the reason.
// This is a function variable to allow mocking in tests.
//
// stdout and stderr are captured separately on purpose: every parser in this
// package (branch name, porcelain status, describe, tag lists, rev-list
// counts) reads the returned string, and with CombinedOutput any stderr
// chatter — GIT_TRACE=1 lines, "warning: tag 'x' is externally known as",
// "refname 'HEAD' is ambiguous" — was parsed as if it were data.
var commandExecutor = func(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimSpace(stdout.String())
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return out, fmt.Errorf("%s: %w", redactCredentials(msg), err)
		}
		return out, err
	}
	return out, nil
}

// credentialPattern matches the userinfo part of a URL (https://user:token@host).
var credentialPattern = regexp.MustCompile(`://[^@/\s]+@`)

// redactCredentials hides URL credentials in text that is logged or returned
// in errors: a pushRepo given as https://user:TOKEN@host/... appears in the
// push arguments, the dry-run/debug command line and git's own messages.
func redactCredentials(s string) string {
	return credentialPattern.ReplaceAllString(s, "://***@")
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
	cmdStr := redactCredentials(fmt.Sprintf("git %s", strings.Join(args, " ")))

	if g.dryRun && isWriteOperation(args) {
		g.logger.DryRun(cmdStr)
		return "", nil
	}

	g.logger.Debug("exec: %s", cmdStr)
	out, err := commandExecutor("git", args...)
	if err != nil {
		// Keep both halves: git's stderr for the user, the root cause for
		// errors.Is/As (project convention).
		if strings.TrimSpace(out) == "" {
			return out, newGitError(fmt.Sprintf("%s: %v", cmdStr, err), err)
		}
		return out, newGitError(fmt.Sprintf("%s: %s: %v", cmdStr, strings.TrimSpace(out), err), err)
	}
	return out, nil
}

// gitError carries a user-facing message with URL credentials redacted while
// still unwrapping to the root cause for errors.Is/As.
type gitError struct {
	msg   string
	cause error
}

func (e *gitError) Error() string { return e.msg }
func (e *gitError) Unwrap() error { return e.cause }

// newGitError builds a gitError whose message is credential-redacted.
func newGitError(msg string, cause error) error {
	return &gitError{msg: redactCredentials(msg), cause: cause}
}

// runSilent executes a git command without logging the output.
// On failure, git's stderr is folded into the returned error so callers
// wrapping it don't lose the actual reason.
func (g *Git) runSilent(args ...string) (string, error) {
	out, err := commandExecutor("git", args...)
	if err != nil {
		if strings.TrimSpace(out) == "" {
			return out, newGitError(err.Error(), err)
		}
		return out, newGitError(fmt.Sprintf("%s: %v", strings.TrimSpace(out), err), err)
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
