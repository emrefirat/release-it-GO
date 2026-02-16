package hook

import (
	"os/exec"
	"testing"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
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
