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

func TestGetLatestPreReleaseTagMerged(t *testing.T) {
	tests := []struct {
		name         string
		preReleaseID string
		gitOutput    string
		gitErr       error
		tagMatch     string
		tagExclude   string
		wantTag      string
		wantErr      bool
	}{
		{
			name:         "finds matching pre-release tag",
			preReleaseID: "deneme",
			gitOutput:    "v1.2.5-deneme.1\nv1.2.5-deneme.0\nv1.2.4",
			wantTag:      "v1.2.5-deneme.1",
		},
		{
			name:         "skips non-matching pre-release IDs",
			preReleaseID: "deneme",
			gitOutput:    "v2.0.0-beta.0\nv1.2.4",
			wantTag:      "",
		},
		{
			name:         "does not match partial ID (beta vs beta2)",
			preReleaseID: "beta",
			gitOutput:    "v1.0.0-beta2.0\nv1.0.0",
			wantTag:      "",
		},
		{
			name:         "empty preReleaseID returns empty",
			preReleaseID: "",
			gitOutput:    "v1.0.0-beta.0",
			wantTag:      "",
		},
		{
			name:         "no tags returns empty",
			preReleaseID: "deneme",
			gitOutput:    "",
			wantTag:      "",
		},
		{
			name:         "respects TagMatch filter",
			preReleaseID: "deneme",
			gitOutput:    "release-1.0.0-deneme.0\nv1.0.0-deneme.0",
			tagMatch:     "v*",
			wantTag:      "v1.0.0-deneme.0",
		},
		{
			name:         "respects TagExclude filter",
			preReleaseID: "deneme",
			gitOutput:    "v1.0.0-deneme.1\nv1.0.0-deneme.0",
			tagExclude:   "*deneme.1*",
			wantTag:      "v1.0.0-deneme.0",
		},
		{
			name:         "git error returns error",
			preReleaseID: "deneme",
			gitErr:       fmt.Errorf("git failed"),
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := commandExecutor
			defer func() { commandExecutor = original }()

			commandExecutor = func(name string, args ...string) (string, error) {
				if tt.gitErr != nil {
					return "", tt.gitErr
				}
				return tt.gitOutput, nil
			}

			cfg := &config.GitConfig{
				TagMatch:   tt.tagMatch,
				TagExclude: tt.tagExclude,
			}
			g := newTestGitWithConfig(cfg, false)

			got, err := g.GetLatestPreReleaseTagMerged(tt.preReleaseID)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantTag {
				t.Errorf("got = %q, want %q", got, tt.wantTag)
			}
		})
	}
}

func TestGetLatestStableTagMerged(t *testing.T) {
	tests := []struct {
		name       string
		gitOutput  string
		gitErr     error
		tagMatch   string
		tagExclude string
		wantTag    string
		wantErr    bool
	}{
		{
			name:      "finds stable tag skipping pre-release",
			gitOutput: "v2.0.0-beta.0\nv1.5.0\nv1.0.0",
			wantTag:   "v1.5.0",
		},
		{
			name:      "all pre-release returns empty",
			gitOutput: "v2.0.0-beta.0\nv1.0.0-alpha.1",
			wantTag:   "",
		},
		{
			name:      "first tag is stable",
			gitOutput: "v3.0.0\nv2.0.0\nv1.0.0",
			wantTag:   "v3.0.0",
		},
		{
			name:      "no tags returns empty",
			gitOutput: "",
			wantTag:   "",
		},
		{
			name:      "respects TagMatch filter",
			gitOutput: "release-2.0.0\nv1.5.0\nv1.0.0",
			tagMatch:  "v*",
			wantTag:   "v1.5.0",
		},
		{
			name:       "respects TagExclude filter",
			gitOutput:  "v2.0.0\nv1.5.0",
			tagExclude: "v2*",
			wantTag:    "v1.5.0",
		},
		{
			name:    "git error returns error",
			gitErr:  fmt.Errorf("git failed"),
			wantErr: true,
		},
		{
			name:      "invalid version tags are skipped",
			gitOutput: "not-a-version\nv1.0.0",
			wantTag:   "v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := commandExecutor
			defer func() { commandExecutor = original }()

			commandExecutor = func(name string, args ...string) (string, error) {
				if tt.gitErr != nil {
					return "", tt.gitErr
				}
				return tt.gitOutput, nil
			}

			cfg := &config.GitConfig{
				TagMatch:   tt.tagMatch,
				TagExclude: tt.tagExclude,
			}
			g := newTestGitWithConfig(cfg, false)

			got, err := g.GetLatestStableTagMerged()
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantTag {
				t.Errorf("got = %q, want %q", got, tt.wantTag)
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
