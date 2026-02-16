package git

import (
	"fmt"
	"strings"
	"testing"

	"release-it-go/internal/config"
)

func TestCreateTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		cmd := strings.Join(args, " ")
		// TagExists check
		if strings.Contains(cmd, "tag -l") {
			return "", nil // tag does not exist
		}
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	err := g.CreateTag("v1.0.0", "Release v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "--annotate") {
		t.Error("expected --annotate flag")
	}
	if !strings.Contains(cmd, "--message") {
		t.Error("expected --message flag")
	}
	if !strings.Contains(cmd, "v1.0.0") {
		t.Error("expected tag name in args")
	}
}

func TestCreateTag_WithArgs(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var executedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		cmd := strings.Join(args, " ")
		if strings.Contains(cmd, "tag -l") {
			return "", nil
		}
		executedArgs = args
		return "", nil
	}

	cfg := &config.GitConfig{
		TagArgs: []string{"--sign"},
	}
	g := newTestGitWithConfig(cfg, false)

	err := g.CreateTag("v1.0.0", "Release v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := strings.Join(executedArgs, " ")
	if !strings.Contains(cmd, "--sign") {
		t.Error("expected --sign in args")
	}
}

func TestCreateTag_AlreadyExists(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "v1.0.0", nil // tag exists
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	err := g.CreateTag("v1.0.0", "Release v1.0.0")
	if err == nil {
		t.Fatal("expected error for existing tag")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestCreateTag_DryRun(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	tagCreated := false
	commandExecutor = func(name string, args ...string) (string, error) {
		cmd := strings.Join(args, " ")
		if strings.Contains(cmd, "tag -l") {
			return "", nil // tag doesn't exist
		}
		if strings.Contains(cmd, "--annotate") {
			tagCreated = true
		}
		return "", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, true)

	err := g.CreateTag("v1.0.0", "Release v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tagCreated {
		t.Error("expected tag creation to be skipped in dry-run mode")
	}
}

func TestGetLatestTag(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "v1.2.3", nil
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	tag, err := g.GetLatestTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.2.3" {
		t.Errorf("expected 'v1.2.3', got %q", tag)
	}
}

func TestGetLatestTag_FromAllRefs(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "v2.0.0\nv1.5.0\nv1.0.0", nil
	}

	cfg := &config.GitConfig{GetLatestTagFromAllRefs: true}
	g := newTestGitWithConfig(cfg, false)

	tag, err := g.GetLatestTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v2.0.0" {
		t.Errorf("expected 'v2.0.0', got %q", tag)
	}
}

func TestGetLatestTag_FromAllRefs_WithMatch(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "v2.0.0\nrelease-1.5.0\nv1.0.0", nil
	}

	cfg := &config.GitConfig{
		GetLatestTagFromAllRefs: true,
		TagMatch:                "v*",
	}
	g := newTestGitWithConfig(cfg, false)

	tag, err := g.GetLatestTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v2.0.0" {
		t.Errorf("expected 'v2.0.0', got %q", tag)
	}
}

func TestGetLatestTag_FromAllRefs_WithExclude(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "v2.0.0-beta\nv1.5.0\nv1.0.0", nil
	}

	cfg := &config.GitConfig{
		GetLatestTagFromAllRefs: true,
		TagExclude:              "*-beta",
	}
	g := newTestGitWithConfig(cfg, false)

	tag, err := g.GetLatestTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.5.0" {
		t.Errorf("expected 'v1.5.0', got %q", tag)
	}
}

func TestGetLatestTag_NoTags(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", fmt.Errorf("fatal: No names found, cannot describe anything")
	}

	cfg := &config.GitConfig{}
	g := newTestGitWithConfig(cfg, false)

	_, err := g.GetLatestTag()
	if err == nil {
		t.Fatal("expected error for no tags")
	}
}

func TestTagExists(t *testing.T) {
	tests := []struct {
		name   string
		output string
		exists bool
	}{
		{"tag exists", "v1.0.0", true},
		{"tag does not exist", "", false},
		{"different tag", "v2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := commandExecutor
			defer func() { commandExecutor = original }()

			commandExecutor = func(name string, args ...string) (string, error) {
				return tt.output, nil
			}

			cfg := &config.GitConfig{}
			g := newTestGitWithConfig(cfg, false)

			exists, err := g.TagExists("v1.0.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if exists != tt.exists {
				t.Errorf("TagExists() = %v, want %v", exists, tt.exists)
			}
		})
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    bool
	}{
		{"wildcard all", "*", "anything", true},
		{"prefix match", "v*", "v1.0.0", true},
		{"prefix no match", "v*", "release-1.0", false},
		{"suffix match", "*-beta", "v1.0.0-beta", true},
		{"suffix no match", "*-beta", "v1.0.0", false},
		{"contains match", "*beta*", "v1.0.0-beta.1", true},
		{"contains no match", "*beta*", "v1.0.0", false},
		{"exact match", "v1.0.0", "v1.0.0", true},
		{"exact no match", "v1.0.0", "v2.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGlob(tt.pattern, tt.input)
			if got != tt.want {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.input, got, tt.want)
			}
		})
	}
}
