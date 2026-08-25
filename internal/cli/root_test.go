package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRootCommand_Help(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("help output should not be empty")
	}
}

func TestNewRootCommand_VersionSubcommand(t *testing.T) {
	Version = "1.0.0-test"
	Commit = "abc1234"
	Date = "2026-02-16"

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestNewRootCommand_DryRunFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--dry-run", "--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_CIFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_VerboseFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"-V", "--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_IncrementFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--increment", "major", "--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_ChangelogFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--changelog", "--ci"})

	// This will fail in test environment (no git repo), which is expected
	_ = cmd.Execute()
}

func TestNewRootCommand_ReleaseVersionFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--release-version", "--ci"})

	// This will fail in test environment (no git repo), which is expected
	_ = cmd.Execute()
}

func TestNewRootCommand_HasExpectedFlags(t *testing.T) {
	cmd := NewRootCommand()

	expectedPersistentFlags := []string{"config", "dry-run", "ci", "verbose", "increment", "preReleaseId"}
	for _, name := range expectedPersistentFlags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("missing persistent flag: %s", name)
		}
	}

	expectedFlags := []string{"changelog", "release-version", "only-version", "no-increment", "no-git.commit", "no-git.tag", "no-git.push"}
	for _, name := range expectedFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag: %s", name)
		}
	}
}

func TestNewRootCommand_HasVersionSubcommand(t *testing.T) {
	cmd := NewRootCommand()

	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("root command should have 'version' subcommand")
	}
}

func TestNewRootCommand_HasCompletionSubcommand(t *testing.T) {
	cmd := NewRootCommand()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "completion [bash|zsh|fish|powershell]" {
			found = true
			break
		}
	}
	if !found {
		t.Error("root command should have 'completion' subcommand")
	}
}

func TestNewRootCommand_CompletionBash(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "bash"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("bash completion output should not be empty")
	}
}

func TestNewRootCommand_CompletionZsh(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "zsh"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("completion zsh failed: %v", err)
	}
}

func TestNewRootCommand_CompletionFish(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "fish"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("completion fish failed: %v", err)
	}
}

func TestNewRootCommand_CompletionPowershell(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"completion", "powershell"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("completion powershell failed: %v", err)
	}
}

func TestNewRootCommand_CompletionInvalidShell(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"completion", "invalid"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid shell")
	}
}

func TestFileExists_RegularFile_ReturnsTrue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(path, []byte("content"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !fileExists(path) {
		t.Error("expected true for existing regular file")
	}
}

func TestFileExists_Directory_ReturnsFalse(t *testing.T) {
	dir := t.TempDir()
	if fileExists(dir) {
		t.Error("expected false for a directory path")
	}
}

func TestFileExists_Missing_ReturnsFalse(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist")
	if fileExists(missing) {
		t.Error("expected false for missing path")
	}
}

func TestFileExists_EmptyString_ReturnsFalse(t *testing.T) {
	if fileExists("") {
		t.Error("expected false for empty path")
	}
}

func TestRunCheckMsg_DirectString_Valid_ReturnsNil(t *testing.T) {
	if err := runCheckMsg("feat: add login", false); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunCheckMsg_DirectString_Invalid_ReturnsError(t *testing.T) {
	stderr := captureStderr(t, func() {
		if err := runCheckMsg("oops no type", false); err == nil {
			t.Error("expected error for non-conventional message")
		}
	})
	if !bytesContains(stderr, "commitlint") {
		t.Errorf("expected commitlint header in stderr, got: %s", stderr)
	}
}

func TestRunCheckMsg_EmptyString_ReturnsEmptyError(t *testing.T) {
	err := runCheckMsg("", false)
	if err == nil || err.Error() != "commit message is empty" {
		t.Errorf("expected 'commit message is empty', got %v", err)
	}
}

func TestRunCheckMsg_FileMode_Valid_ReturnsNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	if err := os.WriteFile(path, []byte("fix(auth): handle expired tokens\n\nbody line"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := runCheckMsg(path, false); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunCheckMsg_FileMode_Invalid_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	if err := os.WriteFile(path, []byte("update docs\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = captureStderr(t, func() {
		if err := runCheckMsg(path, false); err == nil {
			t.Error("expected error for non-conventional file content")
		}
	})
}

func TestRunCheckMsg_FileMode_OnlyFirstLineChecked(t *testing.T) {
	// Valid first line but invalid rest — should still pass (only subject is linted).
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	if err := os.WriteFile(path, []byte("feat: valid subject\nthis body is free-form text"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := runCheckMsg(path, false); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestRunCheckMsg_Stdin_Valid_ReturnsNil(t *testing.T) {
	withStdin(t, "feat: via stdin\n", func() {
		if err := runCheckMsg("-", false); err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})
}

func TestRunCheckMsg_Stdin_Invalid_ReturnsError(t *testing.T) {
	withStdin(t, "no type here\n", func() {
		_ = captureStderr(t, func() {
			if err := runCheckMsg("-", false); err == nil {
				t.Error("expected error for non-conventional stdin content")
			}
		})
	})
}

func TestRunCheckMsg_Verbose_PrintsValidTypes(t *testing.T) {
	stderr := captureStderr(t, func() {
		_ = runCheckMsg("bad msg no colon", true)
	})
	// Verbose output should list valid types — check for a few.
	for _, expected := range []string{"feat", "fix", "chore", "Valid types"} {
		if !bytesContains(stderr, expected) {
			t.Errorf("expected %q in verbose output, got: %s", expected, stderr)
		}
	}
}

func TestRunCheckMsg_FileMode_UnreadableFile_StillTreatedAsString(t *testing.T) {
	// fileExists returns false for directories, so a directory path falls into
	// the default "direct string" branch. This documents current behavior.
	dir := t.TempDir()
	_ = captureStderr(t, func() {
		if err := runCheckMsg(dir, false); err == nil {
			t.Error("expected error — directory path is treated as string and isn't conventional")
		}
	})
}

// captureStderr temporarily replaces os.Stderr for the duration of fn and
// returns whatever was written. Restores os.Stderr on return.
func captureStderr(t *testing.T, fn func()) []byte {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	oldStderr := os.Stderr
	os.Stderr = w
	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.Bytes()
	}()
	fn()
	_ = w.Close()
	os.Stderr = oldStderr
	return <-done
}

// withStdin temporarily replaces os.Stdin with a pipe feeding `input` for the
// duration of fn. Restores os.Stdin on return.
func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()
	go func() {
		_, _ = w.WriteString(input)
		_ = w.Close()
	}()
	fn()
}

func bytesContains(haystack []byte, needle string) bool {
	return bytes.Contains(haystack, []byte(needle))
}

// saveFlagGlobals snapshots the package-level flag variables mutated by
// ParseFlags and restores them via t.Cleanup.
func saveFlagGlobals(t *testing.T) {
	t.Helper()
	origDryRun, origCI, origVerbose := dryRun, ciMode, verboseCount
	t.Cleanup(func() {
		dryRun, ciMode, verboseCount = origDryRun, origCI, origVerbose
	})
}

func TestBuildFlagOverrides_UnsetFlags_AreNil(t *testing.T) {
	saveFlagGlobals(t)
	cmd := NewRootCommand()
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	o := buildFlagOverrides(cmd)
	// Nil means "keep the config-file value" — passing the flag defaults
	// here clobbered config ci/dry-run/verbose on every run.
	if o.DryRun != nil {
		t.Error("DryRun must be nil when --dry-run is not given")
	}
	if o.CI != nil {
		t.Error("CI must be nil when --ci is not given")
	}
	if o.Verbose != nil {
		t.Error("Verbose must be nil when -V is not given")
	}
}

func TestBuildFlagOverrides_SetFlags_ArePresent(t *testing.T) {
	saveFlagGlobals(t)
	cmd := NewRootCommand()
	if err := cmd.ParseFlags([]string{"--ci", "--dry-run", "-VV"}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	o := buildFlagOverrides(cmd)
	if o.DryRun == nil || !*o.DryRun {
		t.Error("expected DryRun override when --dry-run given")
	}
	if o.CI == nil || !*o.CI {
		t.Error("expected CI override when --ci given")
	}
	if o.Verbose == nil || *o.Verbose != 2 {
		t.Error("expected Verbose=2 override when -VV given")
	}
}

func TestReasonDescription(t *testing.T) {
	tests := []struct {
		name     string
		reason   string
		expected string
	}{
		{
			name:     "not conventional substring",
			reason:   "commit 'abc' not in conventional commit format",
			expected: "message must follow conventional commit format",
		},
		{
			name:     "unknown type prefix",
			reason:   "unknown type: fic",
			expected: "type is not in the allowed list",
		},
		{
			name:     "fallback passes reason through",
			reason:   "some other reason",
			expected: "some other reason",
		},
		{
			name:     "empty reason falls through",
			reason:   "",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reasonDescription(tt.reason)
			if got != tt.expected {
				t.Errorf("reasonDescription(%q) = %q, want %q", tt.reason, got, tt.expected)
			}
		})
	}
}
