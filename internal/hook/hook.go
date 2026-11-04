// Package hook provides lifecycle hook execution for the release pipeline.
// Hooks are shell commands that run before/after each pipeline step.
package hook

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"unicode"

	"github.com/emrefirat/release-it-GO/internal/config"
	applog "github.com/emrefirat/release-it-GO/internal/log"
)

// execCommand is a function variable for creating exec.Cmd, allowing test mocking.
var execCommand = exec.Command

// HookRunner executes lifecycle hooks defined in configuration.
type HookRunner struct {
	config *config.HooksConfig
	logger *applog.Logger
	dryRun bool
	vars   map[string]string
}

// NewHookRunner creates a new hook runner instance.
func NewHookRunner(cfg *config.HooksConfig, logger *applog.Logger, dryRun bool) *HookRunner {
	return &HookRunner{
		config: cfg,
		logger: logger,
		dryRun: dryRun,
		vars:   make(map[string]string),
	}
}

// SetVars updates the template variables available for hook command rendering.
func (h *HookRunner) SetVars(vars map[string]string) {
	h.vars = vars
}

// RunHooks executes all hooks for the given lifecycle event.
// Hooks run sequentially; execution stops on first failure.
func (h *HookRunner) RunHooks(lifecycle string) error {
	hooks := h.getHooks(lifecycle)
	if len(hooks) == 0 {
		return nil
	}

	for _, cmd := range hooks {
		if err := h.runCommand(cmd); err != nil {
			return fmt.Errorf("hook %q failed for %s: %w", cmd, lifecycle, err)
		}
	}
	return nil
}

// getHooks returns the hook commands for a given lifecycle event.
func (h *HookRunner) getHooks(lifecycle string) []string {
	if h.config == nil {
		return nil
	}

	switch lifecycle {
	case "before:init":
		return h.config.BeforeInit
	case "after:init":
		return h.config.AfterInit
	case "before:bump":
		return h.config.BeforeBump
	case "after:bump":
		return h.config.AfterBump
	case "before:release":
		return h.config.BeforeRelease
	case "after:release":
		return h.config.AfterRelease
	case "before:git:release":
		return h.config.BeforeGitRelease
	case "after:git:release":
		return h.config.AfterGitRelease
	case "before:github:release":
		return h.config.BeforeGitHubRelease
	case "after:github:release":
		return h.config.AfterGitHubRelease
	case "before:gitlab:release":
		return h.config.BeforeGitLabRelease
	case "after:gitlab:release":
		return h.config.AfterGitLabRelease
	case "before:prerequisites":
		return h.config.BeforePrerequisites
	case "after:prerequisites":
		return h.config.AfterPrerequisites
	case "before:commitlint":
		return h.config.BeforeCommitlint
	case "after:commitlint":
		return h.config.AfterCommitlint
	case "before:version":
		return h.config.BeforeVersion
	case "after:version":
		return h.config.AfterVersion
	case "before:changelog":
		return h.config.BeforeChangelog
	case "after:changelog":
		return h.config.AfterChangelog
	case "before:notification":
		return h.config.BeforeNotification
	case "after:notification":
		return h.config.AfterNotification
	default:
		return nil
	}
}

// runCommand executes a single shell command with template variable substitution.
func (h *HookRunner) runCommand(cmd string) error {
	rendered := renderTemplate(cmd, h.vars)

	if h.dryRun {
		h.logger.DryRun("hook: %s", rendered)
		return nil
	}

	h.logger.Verbose("hook: %s", rendered)

	c := shellCommandFor(runtime.GOOS, os.Getenv("COMSPEC"), rendered)
	c.Env = append(os.Environ(), h.envVars()...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// shellCommandFor builds the platform shell invocation for a hook command.
// Unix uses sh -c; Windows uses %COMSPEC% (cmd.exe) /C — the shipped windows
// binaries previously hard-required an sh on PATH, so every configured hook
// failed on stock Windows.
//
// On Windows the command line is passed verbatim (see windowsCmdLine):
// letting Go escape the arguments turned every double quote in a hook into
// \", which cmd.exe does not unescape — `git commit -m "chore: ${version}"`
// got the wrong message.
func shellCommandFor(goos string, comspec string, command string) *exec.Cmd {
	if goos == "windows" {
		shell := comspec
		if shell == "" {
			shell = "cmd.exe"
		}
		c := execCommand(shell, "/S", "/C", command)
		setVerbatimCmdLine(c, windowsCmdLine(shell, command))
		return c
	}
	return execCommand("sh", "-c", command)
}

// windowsCmdLine builds the raw command line for cmd.exe: /S strips the
// outer quotes and leaves everything between them untouched.
func windowsCmdLine(shell string, command string) string {
	return fmt.Sprintf("\"%s\" /S /C \"%s\"", shell, command)
}

// maxEnvValueLen bounds a single RELEASE_* value. Linux rejects any single
// environment string over 128 KiB (MAX_ARG_STRLEN) with "argument list too
// long", which made every hook — even `echo done` — fail on a release with a
// large changelog.
const maxEnvValueLen = 96 * 1024

// envTruncatedMarker is appended to a value that was cut to fit.
const envTruncatedMarker = "\n[truncated by release-it-go: value exceeded 96 KiB]"

// envVars exposes the template variables as RELEASE_* environment variables
// (version → RELEASE_VERSION, repo.owner → RELEASE_REPO_OWNER, ...) so hook
// scripts can read values without ${var} splicing and its quoting pitfalls.
func (h *HookRunner) envVars() []string {
	env := make([]string, 0, len(h.vars))
	for key, value := range h.vars {
		if len(value) > maxEnvValueLen {
			value = value[:maxEnvValueLen] + envTruncatedMarker
			h.logger.Verbose("RELEASE_%s truncated to %d bytes for the hook environment", envKey(key), maxEnvValueLen)
		}
		env = append(env, "RELEASE_"+envKey(key)+"="+value)
	}
	return env
}

// envKey converts a template variable name to SCREAMING_SNAKE_CASE:
// camelCase humps and dots become underscores.
func envKey(key string) string {
	var b strings.Builder
	for i, r := range key {
		switch {
		case r == '.':
			b.WriteRune('_')
		case r >= 'A' && r <= 'Z':
			if i > 0 {
				b.WriteRune('_')
			}
			b.WriteRune(r)
		default:
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

// templateVarPattern matches ${var} placeholders (dotted keys allowed).
var templateVarPattern = regexp.MustCompile(`\$\{([A-Za-z][A-Za-z0-9.]*)\}`)

// renderTemplate replaces ${var} placeholders in the command string in a
// single pass: a substituted value (a changelog quoting "${version}" from a
// commit message) is never expanded again, and the result no longer depends
// on map iteration order. Unknown placeholders are kept literal.
//
// Values are spliced into a shell command line, so ${changelog} — text
// written by commit authors — should be consumed through the
// RELEASE_CHANGELOG environment variable instead.
func renderTemplate(cmd string, vars map[string]string) string {
	return templateVarPattern.ReplaceAllStringFunc(cmd, func(m string) string {
		if v, ok := vars[m[2:len(m)-1]]; ok {
			return v
		}
		return m
	})
}
