package hook

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emrefirat/release-it-GO/internal/config"
	applog "github.com/emrefirat/release-it-GO/internal/log"
)

func TestNewHookRunner(t *testing.T) {
	cfg := &config.HooksConfig{}
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(cfg, logger, false)

	if runner == nil {
		t.Fatal("expected non-nil HookRunner")
	}
	if runner.dryRun {
		t.Error("expected dryRun to be false")
	}
}

func TestHookRunner_RunHooks_NoConfig(t *testing.T) {
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(nil, logger, false)

	err := runner.RunHooks("before:init")
	if err != nil {
		t.Errorf("expected no error for nil config, got: %v", err)
	}
}

func TestHookRunner_RunHooks_EmptyHooks(t *testing.T) {
	cfg := &config.HooksConfig{}
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(cfg, logger, false)

	err := runner.RunHooks("before:init")
	if err != nil {
		t.Errorf("expected no error for empty hooks, got: %v", err)
	}
}

func TestHookRunner_RunHooks_DryRun(t *testing.T) {
	cfg := &config.HooksConfig{
		BeforeInit: []string{"echo hello"},
	}
	logger := applog.NewLogger(0, true)
	runner := NewHookRunner(cfg, logger, true)

	err := runner.RunHooks("before:init")
	if err != nil {
		t.Errorf("expected no error for dry run, got: %v", err)
	}
}

func TestHookRunner_RunHooks_Success(t *testing.T) {
	cfg := &config.HooksConfig{
		BeforeInit: []string{"echo hello"},
	}
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(cfg, logger, false)

	// Mock exec.Command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "mocked")
	}

	err := runner.RunHooks("before:init")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestHookRunner_RunHooks_Failure(t *testing.T) {
	cfg := &config.HooksConfig{
		BeforeInit: []string{"false"},
	}
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(cfg, logger, false)

	// Mock exec.Command to return a failing command
	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("false")
	}

	err := runner.RunHooks("before:init")
	if err == nil {
		t.Error("expected error for failing hook")
	}
}

func TestHookRunner_SetVars(t *testing.T) {
	cfg := &config.HooksConfig{}
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(cfg, logger, false)

	vars := map[string]string{
		"version": "1.0.0",
		"tagName": "v1.0.0",
	}
	runner.SetVars(vars)

	if runner.vars["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", runner.vars["version"])
	}
}

func TestHookRunner_GetHooks_AllLifecycles(t *testing.T) {
	cfg := &config.HooksConfig{
		BeforeInit:          []string{"cmd1"},
		AfterInit:           []string{"cmd2"},
		BeforeBump:          []string{"cmd3"},
		AfterBump:           []string{"cmd4"},
		BeforeRelease:       []string{"cmd5"},
		AfterRelease:        []string{"cmd6"},
		BeforeGitRelease:    []string{"cmd7"},
		AfterGitRelease:     []string{"cmd8"},
		BeforeGitHubRelease: []string{"cmd9"},
		AfterGitHubRelease:  []string{"cmd10"},
		BeforeGitLabRelease: []string{"cmd11"},
		AfterGitLabRelease:  []string{"cmd12"},
	}
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(cfg, logger, false)

	tests := []struct {
		lifecycle string
		expected  string
	}{
		{"before:init", "cmd1"},
		{"after:init", "cmd2"},
		{"before:bump", "cmd3"},
		{"after:bump", "cmd4"},
		{"before:release", "cmd5"},
		{"after:release", "cmd6"},
		{"before:git:release", "cmd7"},
		{"after:git:release", "cmd8"},
		{"before:github:release", "cmd9"},
		{"after:github:release", "cmd10"},
		{"before:gitlab:release", "cmd11"},
		{"after:gitlab:release", "cmd12"},
		{"unknown:lifecycle", ""},
	}

	for _, tt := range tests {
		t.Run(tt.lifecycle, func(t *testing.T) {
			hooks := runner.getHooks(tt.lifecycle)
			if tt.expected == "" {
				if len(hooks) != 0 {
					t.Errorf("expected no hooks for %s, got %v", tt.lifecycle, hooks)
				}
			} else {
				if len(hooks) != 1 || hooks[0] != tt.expected {
					t.Errorf("expected [%s] for %s, got %v", tt.expected, tt.lifecycle, hooks)
				}
			}
		})
	}
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		vars     map[string]string
		expected string
	}{
		{
			name:     "simple replacement",
			cmd:      "echo ${version}",
			vars:     map[string]string{"version": "1.0.0"},
			expected: "echo 1.0.0",
		},
		{
			name:     "multiple replacements",
			cmd:      "git tag ${tagName} -m 'Release ${version}'",
			vars:     map[string]string{"version": "1.0.0", "tagName": "v1.0.0"},
			expected: "git tag v1.0.0 -m 'Release 1.0.0'",
		},
		{
			name:     "no replacement needed",
			cmd:      "echo hello",
			vars:     map[string]string{"version": "1.0.0"},
			expected: "echo hello",
		},
		{
			name:     "empty vars",
			cmd:      "echo ${version}",
			vars:     map[string]string{},
			expected: "echo ${version}",
		},
		{
			name:     "repo vars",
			cmd:      "echo ${repo.owner}/${repo.repository}",
			vars:     map[string]string{"repo.owner": "emfi", "repo.repository": "release-it-go"},
			expected: "echo emfi/release-it-go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderTemplate(tt.cmd, tt.vars)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHookRunner_RunHooks_MultipleCommands(t *testing.T) {
	cfg := &config.HooksConfig{
		BeforeInit: []string{"echo one", "echo two", "echo three"},
	}
	logger := applog.NewLogger(0, false)
	runner := NewHookRunner(cfg, logger, false)

	origExecCommand := execCommand
	defer func() { execCommand = origExecCommand }()

	callCount := 0
	execCommand = func(name string, arg ...string) *exec.Cmd {
		callCount++
		return exec.Command("echo", "mocked")
	}

	err := runner.RunHooks("before:init")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestEnvKey(t *testing.T) {
	tests := []struct{ in, want string }{
		{"version", "VERSION"},
		{"latestVersion", "LATEST_VERSION"},
		{"tagName", "TAG_NAME"},
		{"branchName", "BRANCH_NAME"},
		{"repo.owner", "REPO_OWNER"},
		{"repo.repository", "REPO_REPOSITORY"},
		{"changelog", "CHANGELOG"},
	}
	for _, tt := range tests {
		if got := envKey(tt.in); got != tt.want {
			t.Errorf("envKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestRunCommand_ExportsReleaseEnvVars(t *testing.T) {
	var captured *exec.Cmd
	original := execCommand
	defer func() { execCommand = original }()
	execCommand = func(name string, arg ...string) *exec.Cmd {
		captured = exec.Command("true") // never actually runs the hook
		return captured
	}

	h := NewHookRunner(&config.HooksConfig{BeforeBump: []string{"echo hi"}}, applog.NewLogger(0, false), false)
	h.SetVars(map[string]string{"version": "1.2.3", "repo.owner": "acme"})

	if err := h.RunHooks("before:bump"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured == nil {
		t.Fatal("execCommand was not called")
	}

	env := strings.Join(captured.Env, "\n")
	if !strings.Contains(env, "RELEASE_VERSION=1.2.3") {
		t.Errorf("expected RELEASE_VERSION in hook env, got:\n%s", env)
	}
	if !strings.Contains(env, "RELEASE_REPO_OWNER=acme") {
		t.Errorf("expected RELEASE_REPO_OWNER in hook env, got:\n%s", env)
	}
	if !strings.Contains(env, "PATH=") {
		t.Error("hook env must inherit the parent environment")
	}
}

func TestShellCommandFor_Platforms(t *testing.T) {
	c := shellCommandFor("linux", "", "echo hi")
	if filepath.Base(c.Args[0]) != "sh" || c.Args[1] != "-c" || c.Args[2] != "echo hi" {
		t.Errorf("linux: args = %v, want sh -c 'echo hi'", c.Args)
	}

	c = shellCommandFor("windows", "", "echo hi")
	if filepath.Base(c.Args[0]) != "cmd.exe" || c.Args[1] != "/C" || c.Args[2] != "echo hi" {
		t.Errorf("windows default: args = %v, want cmd.exe /C 'echo hi'", c.Args)
	}

	c = shellCommandFor("windows", `C:\Windows\system32\cmd.exe`, "echo hi")
	if c.Args[0] != `C:\Windows\system32\cmd.exe` || c.Args[1] != "/C" {
		t.Errorf("windows COMSPEC: args = %v", c.Args)
	}
}

func TestGetHooks_EveryPipelineStepHasAField(t *testing.T) {
	// The runner fires before:/after: for each of these step names; a config
	// key for any of them used to be silently dropped.
	cfg := &config.HooksConfig{
		BeforePrerequisites: []string{"a"}, AfterPrerequisites: []string{"b"},
		BeforeCommitlint: []string{"c"}, AfterCommitlint: []string{"d"},
		BeforeVersion: []string{"e"}, AfterVersion: []string{"f"},
		BeforeChangelog: []string{"g"}, AfterChangelog: []string{"h"},
		BeforeNotification: []string{"i"}, AfterNotification: []string{"j"},
	}
	h := NewHookRunner(cfg, applog.NewLogger(0, false), true)
	steps := []string{"prerequisites", "commitlint", "version", "changelog", "notification"}
	for _, step := range steps {
		if got := h.getHooks("before:" + step); len(got) != 1 {
			t.Errorf("before:%s not mapped (got %v)", step, got)
		}
		if got := h.getHooks("after:" + step); len(got) != 1 {
			t.Errorf("after:%s not mapped (got %v)", step, got)
		}
	}
}
