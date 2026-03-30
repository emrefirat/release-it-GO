// Package githook installs and manages git hooks from release-it-go configuration.
// Git hooks (pre-commit, commit-msg, etc.) are defined in the hooks section of the
// config alongside release lifecycle hooks. The install command filters git hook names
// and writes them as executable shell scripts to .git/hooks/.
package githook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"release-it-go/internal/config"
)

// managedHeader is the marker comment in generated hook scripts.
// Used to identify hooks managed by release-it-go vs user-created hooks.
const managedHeader = "# Managed by release-it-go — DO NOT EDIT"

// supportedGitHooks lists the git hook names that can be installed.
var supportedGitHooks = []string{
	"pre-commit",
	"commit-msg",
	"pre-push",
	"post-commit",
	"post-merge",
	"prepare-commit-msg",
}

// commandExecutor is the function used to run external commands.
// Replaced in tests for isolation.
var commandExecutor = func(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// Installer handles git hook installation and removal.
type Installer struct {
	gitDir   string
	hooksDir string
	force    bool
}

// NewInstaller creates an Installer for the given git directory.
func NewInstaller(gitDir string, force bool) *Installer {
	return &Installer{
		gitDir:   gitDir,
		hooksDir: filepath.Join(gitDir, "hooks"),
		force:    force,
	}
}

// FindGitDir locates the .git directory for the current repository.
func FindGitDir() (string, error) {
	out, err := commandExecutor("git", "rev-parse", "--git-dir")
	if err != nil {
		return "", fmt.Errorf("not a git repository (run from project root)")
	}
	return out, nil
}

// Install writes git hook scripts to .git/hooks/ for each configured hook.
// Returns an error if a non-managed hook exists and force is false.
func (i *Installer) Install(hooks map[string][]string) error {
	if err := os.MkdirAll(i.hooksDir, 0755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	installed := 0
	for name, commands := range hooks {
		if len(commands) == 0 {
			continue
		}

		hookPath := filepath.Join(i.hooksDir, name)

		// Check for existing non-managed hooks
		if exists(hookPath) {
			managed, err := i.isManagedHook(hookPath)
			if err != nil {
				return fmt.Errorf("checking hook %s: %w", name, err)
			}
			if !managed && !i.force {
				return fmt.Errorf("hook %s already exists and is not managed by release-it-go (use --force to overwrite)", name)
			}
		}

		script := generateScript(name, commands)
		if err := os.WriteFile(hookPath, []byte(script), 0755); err != nil {
			return fmt.Errorf("writing hook %s: %w", name, err)
		}
		installed++
		fmt.Printf("  ✓ Installed %s\n", name)
	}

	if installed == 0 {
		fmt.Println("No git hooks to install.")
	} else {
		fmt.Printf("\n%d git hook(s) installed to %s\n", installed, i.hooksDir)
	}
	return nil
}

// Remove deletes all managed git hooks from .git/hooks/.
// Non-managed (user-created) hooks are left untouched.
func (i *Installer) Remove() error {
	if !exists(i.hooksDir) {
		fmt.Println("No hooks directory found.")
		return nil
	}

	removed := 0
	for _, name := range supportedGitHooks {
		hookPath := filepath.Join(i.hooksDir, name)
		if !exists(hookPath) {
			continue
		}

		managed, err := i.isManagedHook(hookPath)
		if err != nil {
			return fmt.Errorf("checking hook %s: %w", name, err)
		}
		if !managed {
			continue
		}

		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("removing hook %s: %w", name, err)
		}
		removed++
		fmt.Printf("  ✓ Removed %s\n", name)
	}

	if removed == 0 {
		fmt.Println("No managed hooks found to remove.")
	} else {
		fmt.Printf("\n%d managed hook(s) removed.\n", removed)
	}
	return nil
}

// HooksFromConfig extracts git hook definitions from HooksConfig.
// Only returns entries for supported git hook names, ignoring lifecycle hooks.
func HooksFromConfig(cfg *config.HooksConfig) map[string][]string {
	hooks := make(map[string][]string)
	if len(cfg.PreCommit) > 0 {
		hooks["pre-commit"] = cfg.PreCommit
	}
	if len(cfg.CommitMsg) > 0 {
		hooks["commit-msg"] = cfg.CommitMsg
	}
	if len(cfg.PrePush) > 0 {
		hooks["pre-push"] = cfg.PrePush
	}
	if len(cfg.PostCommit) > 0 {
		hooks["post-commit"] = cfg.PostCommit
	}
	if len(cfg.PostMerge) > 0 {
		hooks["post-merge"] = cfg.PostMerge
	}
	if len(cfg.PrepareCommitMsg) > 0 {
		hooks["prepare-commit-msg"] = cfg.PrepareCommitMsg
	}
	return hooks
}

// generateScript creates a shell script for a git hook.
func generateScript(hookName string, commands []string) string {
	var b strings.Builder
	b.WriteString("#!/bin/sh\n")
	b.WriteString(managedHeader + "\n")
	fmt.Fprintf(&b, "# Hook: %s\n", hookName)
	fmt.Fprintf(&b, "# Generated: %s\n", time.Now().Format("2006-01-02"))
	b.WriteString("\nset -e\n\n")
	for _, cmd := range commands {
		b.WriteString(cmd + "\n")
	}
	return b.String()
}

// isManagedHook checks if an existing hook file was created by release-it-go.
func (i *Installer) isManagedHook(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(data), managedHeader), nil
}

// exists checks if a file or directory exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
