package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHooksCommand_Registered(t *testing.T) {
	rootCmd := NewRootCommand()

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "hooks" {
			found = true
			// Verify subcommands
			installFound := false
			removeFound := false
			for _, sub := range cmd.Commands() {
				if sub.Use == "install" {
					installFound = true
				}
				if sub.Use == "remove" {
					removeFound = true
				}
			}
			if !installFound {
				t.Error("expected 'install' subcommand under 'hooks'")
			}
			if !removeFound {
				t.Error("expected 'remove' subcommand under 'hooks'")
			}
		}
	}
	if !found {
		t.Error("expected 'hooks' command to be registered")
	}
}

func TestHooksInstallCommand_HasForceFlag(t *testing.T) {
	cmd := newHooksInstallCommand()
	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Error("expected --force flag on hooks install command")
	}
}

func TestHooksRemoveCommand_NoForceFlag(t *testing.T) {
	cmd := newHooksRemoveCommand()
	flag := cmd.Flags().Lookup("force")
	if flag != nil {
		t.Error("remove command should not have --force flag")
	}
}

// setupGitRepo creates a fresh git repo in a tempdir, chdirs into it, and
// returns the directory. Cleanup (chdir back, clear cfgFile) is registered
// via t.Cleanup.
func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustRun(t, dir, "git", "init")
	mustRun(t, dir, "git", "config", "user.email", "test@example.com")
	mustRun(t, dir, "git", "config", "user.name", "Test User")

	origCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	origCfgFile := cfgFile
	origHooksForce := hooksForce
	t.Cleanup(func() {
		_ = os.Chdir(origCwd)
		cfgFile = origCfgFile
		hooksForce = origHooksForce
	})
	return dir
}

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func writeConfig(t *testing.T, dir, content string) {
	t.Helper()
	path := filepath.Join(dir, ".release-it-go.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfgFile = path
}

func TestRunHooksInstall_NotAGitRepo_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origCwd) })

	err := runHooksInstall(nil, nil)
	if err == nil {
		t.Error("expected error when not inside a git repository")
	}
}

func TestRunHooksInstall_NoHooksConfigured_ReturnsNil(t *testing.T) {
	dir := setupGitRepo(t)
	writeConfig(t, dir, "git:\n  commit: true\n") // No hooks section
	if err := runHooksInstall(nil, nil); err != nil {
		t.Errorf("expected nil when no hooks configured, got %v", err)
	}
	// No .hooks/ directory should exist
	if _, err := os.Stat(filepath.Join(dir, ".hooks")); !os.IsNotExist(err) {
		t.Error(".hooks directory should not exist when no hooks configured")
	}
}

func TestRunHooksInstall_WritesHookWithManagedHeader(t *testing.T) {
	dir := setupGitRepo(t)
	writeConfig(t, dir, `hooks:
  "pre-commit":
    - "echo from pre-commit"
`)
	if err := runHooksInstall(nil, nil); err != nil {
		t.Fatalf("install: %v", err)
	}

	path := filepath.Join(dir, ".hooks", "pre-commit")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read hook: %v", err)
	}
	if !strings.Contains(string(content), "Managed by release-it-go") {
		t.Error("expected managed header in hook file")
	}
	if !strings.Contains(string(content), "echo from pre-commit") {
		t.Error("expected hook body to include configured command")
	}

	// File should be executable (0755 or at least user-executable)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm()&0o100 == 0 {
		t.Error("hook file should be user-executable")
	}
}

func TestRunHooksInstall_SetsCoreHooksPath(t *testing.T) {
	dir := setupGitRepo(t)
	writeConfig(t, dir, `hooks:
  "pre-commit":
    - "echo hi"
`)
	if err := runHooksInstall(nil, nil); err != nil {
		t.Fatalf("install: %v", err)
	}

	out, err := exec.Command("git", "-C", dir, "config", "--local", "core.hooksPath").Output()
	if err != nil {
		t.Fatalf("read core.hooksPath: %v", err)
	}
	if strings.TrimSpace(string(out)) != ".hooks" {
		t.Errorf("expected core.hooksPath=.hooks, got %q", strings.TrimSpace(string(out)))
	}
}

func TestRunHooksInstall_MultipleHooks_AllInstalled(t *testing.T) {
	dir := setupGitRepo(t)
	writeConfig(t, dir, `hooks:
  "pre-commit":
    - "echo pc"
  "commit-msg":
    - "echo cm"
  "pre-push":
    - "echo pp"
`)
	if err := runHooksInstall(nil, nil); err != nil {
		t.Fatalf("install: %v", err)
	}
	for _, name := range []string{"pre-commit", "commit-msg", "pre-push"} {
		if _, err := os.Stat(filepath.Join(dir, ".hooks", name)); err != nil {
			t.Errorf("expected %s hook file, got error: %v", name, err)
		}
	}
}

func TestRunHooksInstall_NonManagedHookWithoutForce_ReturnsError(t *testing.T) {
	dir := setupGitRepo(t)
	// Pre-create a non-managed hook file in .hooks/
	hooksDir := filepath.Join(dir, ".hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"),
		[]byte("#!/bin/sh\necho user hook\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	writeConfig(t, dir, `hooks:
  "pre-commit":
    - "echo replaced"
`)
	hooksForce = false
	err := runHooksInstall(nil, nil)
	if err == nil {
		t.Error("expected error when overwriting a non-managed hook without --force")
	}
}

func TestRunHooksInstall_NonManagedHookWithForce_Overwrites(t *testing.T) {
	dir := setupGitRepo(t)
	hooksDir := filepath.Join(dir, ".hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(hooksDir, "pre-commit")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho user hook\n"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	writeConfig(t, dir, `hooks:
  "pre-commit":
    - "echo replaced"
`)
	hooksForce = true
	if err := runHooksInstall(nil, nil); err != nil {
		t.Fatalf("install with --force: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(content), "echo replaced") {
		t.Error("expected hook to be overwritten with configured command")
	}
	if !strings.Contains(string(content), "Managed by release-it-go") {
		t.Error("expected managed header after force install")
	}
}

func TestRunHooksRemove_NotAGitRepo_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origCwd) })

	if err := runHooksRemove(nil, nil); err == nil {
		t.Error("expected error when not inside a git repository")
	}
}

func TestRunHooksRemove_RemovesManagedHook(t *testing.T) {
	dir := setupGitRepo(t)
	writeConfig(t, dir, `hooks:
  "pre-commit":
    - "echo x"
`)
	if err := runHooksInstall(nil, nil); err != nil {
		t.Fatalf("install: %v", err)
	}
	path := filepath.Join(dir, ".hooks", "pre-commit")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("hook should exist after install: %v", err)
	}

	if err := runHooksRemove(nil, nil); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected managed hook to be removed")
	}
}

func TestRunHooksRemove_LeavesNonManagedHookUntouched(t *testing.T) {
	dir := setupGitRepo(t)
	hooksDir := filepath.Join(dir, ".hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	userHook := filepath.Join(hooksDir, "pre-commit")
	userContent := "#!/bin/sh\necho handcrafted\n"
	if err := os.WriteFile(userHook, []byte(userContent), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := runHooksRemove(nil, nil); err != nil {
		t.Fatalf("remove: %v", err)
	}

	content, err := os.ReadFile(userHook)
	if err != nil {
		t.Fatalf("user hook should still exist: %v", err)
	}
	if string(content) != userContent {
		t.Error("user hook content was modified")
	}
}
