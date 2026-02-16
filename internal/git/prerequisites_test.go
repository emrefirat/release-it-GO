package git

import (
	"fmt"
	"strings"
	"testing"

	"github.com/emfi/release-it-go/internal/config"
	applog "github.com/emfi/release-it-go/internal/log"
)

func newTestGitWithConfig(cfg *config.GitConfig, dryRun bool) *Git {
	logger := applog.NewLogger(0, false)
	return NewGit(cfg, logger, dryRun)
}

func TestCheckPrerequisites_AllDisabled(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		cmd := strings.Join(args, " ")
		if strings.Contains(cmd, "rev-parse --is-inside-work-tree") {
			return "true", nil
		}
		return "", nil
	}

	cfg := &config.GitConfig{
		RequireBranch:          "",
		RequireCleanWorkingDir: false,
		RequireUpstream:        false,
		RequireCommits:         false,
	}

	g := newTestGitWithConfig(cfg, false)
	err := g.CheckPrerequisites()
	if err != nil {
		t.Errorf("expected no error with all checks disabled, got: %v", err)
	}
}

func TestCheckBranch(t *testing.T) {
	tests := []struct {
		name          string
		requireBranch string
		currentBranch string
		wantErr       bool
	}{
		{"no requirement", "", "feature", false},
		{"matching branch", "main", "main", false},
		{"non-matching branch", "main", "develop", true},
		{"wildcard pattern", "release/*", "release/1.0", false},
		{"wildcard no match", "release/*", "main", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := commandExecutor
			defer func() { commandExecutor = original }()

			commandExecutor = func(name string, args ...string) (string, error) {
				return tt.currentBranch, nil
			}

			cfg := &config.GitConfig{RequireBranch: tt.requireBranch}
			g := newTestGitWithConfig(cfg, false)
			err := g.checkBranch()

			if (err != nil) != tt.wantErr {
				t.Errorf("checkBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckCleanWorkingDir(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		output  string
		wantErr bool
	}{
		{"disabled", false, "M file.go", false},
		{"clean", true, "", false},
		{"dirty", true, "M file.go", true},
		{"dirty with whitespace", true, " M file.go\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := commandExecutor
			defer func() { commandExecutor = original }()

			commandExecutor = func(name string, args ...string) (string, error) {
				return tt.output, nil
			}

			cfg := &config.GitConfig{RequireCleanWorkingDir: tt.enabled}
			g := newTestGitWithConfig(cfg, false)
			err := g.checkCleanWorkingDir()

			if (err != nil) != tt.wantErr {
				t.Errorf("checkCleanWorkingDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckUpstream(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		hasErr  bool
		wantErr bool
	}{
		{"disabled", false, true, false},
		{"upstream exists", true, false, false},
		{"no upstream", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := commandExecutor
			defer func() { commandExecutor = original }()

			commandExecutor = func(name string, args ...string) (string, error) {
				if tt.hasErr {
					return "", fmt.Errorf("no upstream")
				}
				return "origin/main", nil
			}

			cfg := &config.GitConfig{RequireUpstream: tt.enabled}
			g := newTestGitWithConfig(cfg, false)
			err := g.checkUpstream()

			if (err != nil) != tt.wantErr {
				t.Errorf("checkUpstream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckCommits(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		tagOutput string
		tagErr    error
		logOutput string
		wantErr   bool
	}{
		{"disabled", false, "", nil, "", false},
		{"has commits", true, "v1.0.0", nil, "abc1234 some commit", false},
		{"no commits", true, "v1.0.0", nil, "", true},
		{"no tags yet", true, "", fmt.Errorf("no tags"), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := commandExecutor
			defer func() { commandExecutor = original }()

			commandExecutor = func(name string, args ...string) (string, error) {
				cmd := strings.Join(args, " ")
				if strings.Contains(cmd, "describe --tags") {
					if tt.tagErr != nil {
						return "", tt.tagErr
					}
					return tt.tagOutput, nil
				}
				if strings.Contains(cmd, "log") {
					return tt.logOutput, nil
				}
				return "", nil
			}

			cfg := &config.GitConfig{RequireCommits: tt.enabled}
			g := newTestGitWithConfig(cfg, false)
			err := g.checkCommits()

			if (err != nil) != tt.wantErr {
				t.Errorf("checkCommits() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
