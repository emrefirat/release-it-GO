// Package hook provides lifecycle hook execution for the release pipeline.
// Hooks are shell commands that run before/after each pipeline step.
package hook

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
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
func shellCommandFor(goos string, comspec string, command string) *exec.Cmd {
	if goos == "windows" {
		shell := comspec
		if shell == "" {
			shell = "cmd.exe"
		}
		return execCommand(shell, "/C", command)
	}
	return execCommand("sh", "-c", command)
}

// envVars exposes the template variables as RELEASE_* environment variables
// (version → RELEASE_VERSION, repo.owner → RELEASE_REPO_OWNER, ...) so hook
// scripts can read values without ${var} splicing and its quoting pitfalls.
func (h *HookRunner) envVars() []string {
	env := make([]string, 0, len(h.vars))
	for key, value := range h.vars {
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

// renderTemplate replaces ${var} placeholders in the command string.
func renderTemplate(cmd string, vars map[string]string) string {
	result := cmd
	for key, value := range vars {
		placeholder := "${" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}
