package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// captureStdout mirrors captureStderr for fmt.Println output.
func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = w
	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		done <- buf.Bytes()
	}()
	fn()
	_ = w.Close()
	os.Stdout = oldStdout
	return <-done
}

// setupReleaseRepo creates a temp git repo (cwd switched) with a v1.0.0 tag,
// one feat commit on top, and a config that never pushes. Every Execute test
// below MUST run inside such a repo: the previous versions of these tests ran
// `--ci` from the package directory inside the real repository, where the
// project's push:false config is not found — on a clean tree they performed
// an actual release (internal/cli/CHANGELOG.md was committed that way).
func setupReleaseRepo(t *testing.T) string {
	t.Helper()
	dir := setupGitRepo(t)
	writeConfig(t, dir, "git:\n  push: false\n  requireUpstream: false\n  tagName: \"v${version}\"\n")
	cfgFile = "" // Execute must find the config via cwd search like a real run
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-q", "-m", "chore: init")
	mustRun(t, dir, "git", "tag", "v1.0.0")
	if err := os.WriteFile(filepath.Join(dir, "feature.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-q", "-m", "feat: add feature")
	return dir
}

func tagExists(t *testing.T, dir, tag string) bool {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "tag", "-l", tag).Output()
	if err != nil {
		t.Fatalf("git tag -l: %v", err)
	}
	return strings.TrimSpace(string(out)) == tag
}

func TestExecute_DryRun_CreatesNothing(t *testing.T) {
	saveFlagGlobals(t)
	dir := setupReleaseRepo(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--dry-run", "--ci"})
	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Errorf("dry-run release failed: %v", err)
		}
	})

	if tagExists(t, dir, "v1.1.0") {
		t.Error("dry-run must not create a tag")
	}
	if !strings.Contains(string(stderr), "dry-run") {
		t.Errorf("expected dry-run banner in output, got:\n%s", stderr)
	}
}

func TestExecute_CI_PerformsReleaseInTempRepo(t *testing.T) {
	saveFlagGlobals(t)
	dir := setupReleaseRepo(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--ci"})
	_ = captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("release failed: %v", err)
		}
	})

	// feat commit → minor bump, v-prefixed template
	if !tagExists(t, dir, "v1.1.0") {
		t.Error("expected v1.1.0 tag after a --ci release")
	}
}

func TestExecute_PositionalIncrement_EndToEnd(t *testing.T) {
	saveFlagGlobals(t)
	dir := setupReleaseRepo(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"major", "--ci"})
	_ = captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("release failed: %v", err)
		}
	})

	if !tagExists(t, dir, "v2.0.0") {
		t.Error("positional 'major' through cobra should create v2.0.0")
	}
	if tagExists(t, dir, "v1.1.0") {
		t.Error("auto-detected minor must not win over the positional increment")
	}
}

func TestExecute_IncrementFlag_DryRun(t *testing.T) {
	saveFlagGlobals(t)
	dir := setupReleaseRepo(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--increment", "major", "--ci", "--dry-run"})
	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("dry-run failed: %v", err)
		}
	})

	if !strings.Contains(string(stderr), "1.0.0 → 2.0.0") {
		t.Errorf("expected the major bump to be announced, got:\n%s", stderr)
	}
	if tagExists(t, dir, "v2.0.0") {
		t.Error("dry-run must not create the tag")
	}
}

func TestExecute_VerboseFlag_DryRun(t *testing.T) {
	saveFlagGlobals(t)
	setupReleaseRepo(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"-V", "--ci", "--dry-run"})
	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("verbose dry-run failed: %v", err)
		}
	})
	// -V shows the ↳ detail lines (e.g. hook/git commands)
	if !strings.Contains(string(stderr), "↳") {
		t.Errorf("expected verbose detail lines, got:\n%s", stderr)
	}
}

func TestExecute_ChangelogFlag_PrintsWithoutReleasing(t *testing.T) {
	saveFlagGlobals(t)
	dir := setupReleaseRepo(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--changelog", "--ci"})
	var stdout []byte
	_ = captureStderr(t, func() {
		stdout = captureStdout(t, func() {
			if err := cmd.Execute(); err != nil {
				t.Fatalf("--changelog failed: %v", err)
			}
		})
	})

	if !strings.Contains(string(stdout), "add feature") {
		t.Errorf("expected the feat commit in the changelog output, got:\n%s", stdout)
	}
	if tagExists(t, dir, "v1.1.0") {
		t.Error("--changelog must not create a tag")
	}
}

func TestExecute_ReleaseVersionFlag_PrintsNextVersion(t *testing.T) {
	saveFlagGlobals(t)
	dir := setupReleaseRepo(t)

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--release-version", "--ci"})
	var stdout []byte
	_ = captureStderr(t, func() {
		stdout = captureStdout(t, func() {
			if err := cmd.Execute(); err != nil {
				t.Fatalf("--release-version failed: %v", err)
			}
		})
	})

	if strings.TrimSpace(string(stdout)) != "1.1.0" {
		t.Errorf("stdout = %q, want exactly \"1.1.0\" (scripting mode)", stdout)
	}
	if tagExists(t, dir, "v1.1.0") {
		t.Error("--release-version must not create a tag")
	}
}

func TestExecute_OutsideGitRepo_ReturnsError(t *testing.T) {
	saveFlagGlobals(t)
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origCwd) })

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--ci", "--dry-run"})
	var err error
	_ = captureStderr(t, func() { err = cmd.Execute() })
	if err == nil {
		t.Fatal("expected an error outside a git repository, got nil")
	}
	if !strings.Contains(err.Error(), "git") {
		t.Errorf("error should explain the git precondition, got: %v", err)
	}
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
	if !bytesContains(stderr, "Invalid commit message") {
		t.Errorf("expected the diagnostic header in stderr, got: %s", stderr)
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

func TestResolveIncrementArg(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		flag    string
		want    string
		wantErr bool
	}{
		{"no arg, flag only", nil, "minor", "minor", false},
		{"increment keyword arg", []string{"major"}, "", "major", false},
		{"prerelease keyword arg", []string{"prerelease"}, "", "prerelease", false},
		{"explicit version arg", []string{"1.5.0"}, "", "1.5.0", false},
		{"v-prefixed version arg", []string{"v1.5.0"}, "", "v1.5.0", false},
		{"invalid arg", []string{"bogus"}, "", "", true},
		{"conflicting arg and flag", []string{"minor"}, "patch", "", true},
		{"agreeing arg and flag", []string{"minor"}, "minor", "minor", false},
		{"empty everything", nil, "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveIncrementArg(tt.args, tt.flag)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewRootCommand_RejectsMultipleArgs(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"minor", "extra"})
	if err := cmd.Execute(); err == nil {
		t.Error("expected error for two positional args")
	}
}

func TestFirstSubjectLine(t *testing.T) {
	tests := []struct{ in, want string }{
		{"feat: x\n\nbody", "feat: x"},
		{"\n\nfeat: x\n", "feat: x"},                                 // leading blank lines (git strips them later)
		{"# Please enter the commit message\n#\nfix: y\n", "fix: y"}, // template comments
		{"  feat: z  ", "feat: z"},
		{"", ""},
		{"# only comments\n#\n", ""},
	}
	for _, tt := range tests {
		if got := firstSubjectLine(tt.in); got != tt.want {
			t.Errorf("firstSubjectLine(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRunCheckMsg_FileMode_SkipsLeadingBlankAndCommentLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "COMMIT_EDITMSG")
	content := "\n# Please enter the commit message for your changes.\n#\nfeat: real subject\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := runCheckMsg(path, false); err != nil {
		t.Errorf("expected nil (subject is the first real line), got %v", err)
	}
}

func TestRunCheckMsg_Output_IsScannable(t *testing.T) {
	stderr := captureStderr(t, func() {
		_ = runCheckMsg("fic: deneme", false)
	})
	out := string(stderr)
	for _, want := range []string{"Invalid commit message", "message:", "problem:", "Expected:", "Example:", "Types:"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "commitlint —") {
		t.Errorf("old cramped header must be gone, got:\n%s", out)
	}
}

func TestRunCheckMsg_UnknownType_SuggestsClosestAndRewritesExample(t *testing.T) {
	stderr := captureStderr(t, func() {
		_ = runCheckMsg("fic: deneme", false)
	})
	out := string(stderr)
	if !strings.Contains(out, `did you mean "fix"`) {
		t.Errorf("expected a suggestion for the typo, got:\n%s", out)
	}
	if !strings.Contains(out, "fix: deneme") {
		t.Errorf("example should be the user's own message with the corrected type, got:\n%s", out)
	}
}

func TestRunCheckMsg_ReturnsSentinel(t *testing.T) {
	var err error
	_ = captureStderr(t, func() { err = runCheckMsg("bogus", false) })
	if !errors.Is(err, ErrCheckFailed) {
		t.Errorf("expected ErrCheckFailed so Execute skips the redundant Error: line, got %v", err)
	}
}

func TestRunCheckMsg_Verbose_ListsEveryAcceptedType(t *testing.T) {
	stderr := captureStderr(t, func() { _ = runCheckMsg("bogus", true) })
	if !strings.Contains(string(stderr), "build") {
		t.Errorf("build is accepted by the linter and must appear in the verbose type list, got:\n%s", stderr)
	}
}

func TestRunCheckMsg_Verbose_DoesNotClaimUnenforcedRules(t *testing.T) {
	stderr := captureStderr(t, func() { _ = runCheckMsg("bogus", true) })
	out := string(stderr)
	if strings.Contains(out, "must start with lowercase") {
		t.Errorf("help claims a description-case rule the linter does not enforce:\n%s", out)
	}
	if !strings.Contains(out, "type must be lowercase") {
		t.Errorf("help should state the rule that IS enforced (type case), got:\n%s", out)
	}
}

func TestRunCheckMsg_LabelColumnsAlign(t *testing.T) {
	stderr := captureStderr(t, func() { _ = runCheckMsg("fic: deneme", false) })
	lines := strings.Split(string(stderr), "\n")
	col := map[string]int{}
	for _, line := range lines {
		for _, label := range []string{"message:", "problem:", "Expected:", "Example:", "Types:"} {
			if idx := strings.Index(line, label); idx >= 0 {
				col[label] = idx
			}
		}
	}
	if len(col) < 5 {
		t.Fatalf("labels missing: %v\n%s", col, stderr)
	}
	for label, idx := range col {
		if idx != col["message:"] {
			t.Errorf("label %q at column %d, want %d (all labels aligned)", label, idx, col["message:"])
		}
	}
}

func TestExecute_CheckMsg_NoConfigWarningSuppressed(t *testing.T) {
	saveFlagGlobals(t)
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origCwd) })

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--check-msg", "feat: ok"})
	stderr := captureStderr(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Errorf("valid message must pass, got %v", err)
		}
	})
	// The commit-msg hook runs this on every commit; a config-file warning
	// for a mode that never reads the config is pure noise.
	if strings.Contains(string(stderr), "No config file found") {
		t.Errorf("no-config warning must be suppressed in --check-msg mode, got:\n%s", stderr)
	}
}

// --- mode flags: silent if-order precedence replaced by explicit errors ---

func TestExecute_ConflictingModeFlags_AreErrors(t *testing.T) {
	saveFlagGlobals(t)
	dir := t.TempDir()
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origCwd) })

	cases := [][]string{
		{"--changelog", "--release-version"},
		{"--check-msg", "feat: x", "--dry-run"},
		{"--only-version", "--no-increment"},
		{"minor", "--no-increment"},
		{"--increment", "major", "--no-increment"},
		{"--check-commits", "--changelog"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			cmd := NewRootCommand()
			cmd.SetArgs(args)
			var err error
			_ = captureStderr(t, func() { err = cmd.Execute() })
			if err == nil {
				t.Fatalf("%v: the losing flag used to be dropped silently; expected an error", args)
			}
			if !strings.Contains(err.Error(), "cannot be combined") {
				t.Errorf("%v: error should say the flags cannot be combined, got: %v", args, err)
			}
		})
	}
}

func TestExecute_InvalidIncrementFlag_FailsBeforePipeline(t *testing.T) {
	saveFlagGlobals(t)
	dir := t.TempDir() // not a git repo on purpose: validation must fire first
	origCwd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origCwd) })

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--increment", "bogus", "--ci"})
	var err error
	_ = captureStderr(t, func() { err = cmd.Execute() })
	if err == nil {
		t.Fatal("-i bogus must be rejected up front (it used to fail after init and prerequisites)")
	}
	if !strings.Contains(err.Error(), "increment") {
		t.Errorf("error should name the increment flag, got: %v", err)
	}
	if strings.Contains(err.Error(), "git") {
		t.Errorf("validation must run before any git interaction, got: %v", err)
	}
}
