package git

import (
	"fmt"
	"testing"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantProto  string
		wantHost   string
		wantOwner  string
		wantRepo   string
		wantRemote string // expected sanitized Remote URL (empty = same as cleaned https URL)
		wantErr    bool
	}{
		{
			name:       "HTTPS with .git",
			url:        "https://github.com/emfi/release-it-go.git",
			wantProto:  "https",
			wantHost:   "github.com",
			wantOwner:  "emfi",
			wantRepo:   "release-it-go",
			wantRemote: "https://github.com/emfi/release-it-go",
		},
		{
			name:       "HTTPS without .git",
			url:        "https://github.com/emfi/release-it-go",
			wantProto:  "https",
			wantHost:   "github.com",
			wantOwner:  "emfi",
			wantRepo:   "release-it-go",
			wantRemote: "https://github.com/emfi/release-it-go",
		},
		{
			name:      "SSH format",
			url:       "git@github.com:emfi/release-it-go.git",
			wantProto: "ssh",
			wantHost:  "github.com",
			wantOwner: "emfi",
			wantRepo:  "release-it-go",
		},
		{
			name:      "SSH without .git",
			url:       "git@github.com:emfi/release-it-go",
			wantProto: "ssh",
			wantHost:  "github.com",
			wantOwner: "emfi",
			wantRepo:  "release-it-go",
		},
		{
			name:       "GitLab HTTPS",
			url:        "https://gitlab.com/mygroup/myproject.git",
			wantProto:  "https",
			wantHost:   "gitlab.com",
			wantOwner:  "mygroup",
			wantRepo:   "myproject",
			wantRemote: "https://gitlab.com/mygroup/myproject",
		},
		{
			name:      "GitLab SSH",
			url:       "git@gitlab.com:mygroup/myproject.git",
			wantProto: "ssh",
			wantHost:  "gitlab.com",
			wantOwner: "mygroup",
			wantRepo:  "myproject",
		},
		{
			name:       "Self-hosted HTTPS",
			url:        "https://git.example.com/team/repo.git",
			wantProto:  "https",
			wantHost:   "git.example.com",
			wantOwner:  "team",
			wantRepo:   "repo",
			wantRemote: "https://git.example.com/team/repo",
		},
		{
			name:       "HTTPS with credentials (oauth2)",
			url:        "https://oauth2:secret-token@gitlab.com/mygroup/myproject.git",
			wantProto:  "https",
			wantHost:   "gitlab.com",
			wantOwner:  "mygroup",
			wantRepo:   "myproject",
			wantRemote: "https://gitlab.com/mygroup/myproject",
		},
		{
			name:       "HTTPS with user:password credentials",
			url:        "https://user:password123@github.com/owner/repo.git",
			wantProto:  "https",
			wantHost:   "github.com",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantRemote: "https://github.com/owner/repo",
		},
		{
			name:       "HTTPS with token-only credential",
			url:        "https://x-access-token:ghp_abc123@github.com/owner/repo",
			wantProto:  "https",
			wantHost:   "github.com",
			wantOwner:  "owner",
			wantRepo:   "repo",
			wantRemote: "https://github.com/owner/repo",
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseRepoURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRepoURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if info.Protocol != tt.wantProto {
				t.Errorf("Protocol = %q, want %q", info.Protocol, tt.wantProto)
			}
			if info.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", info.Host, tt.wantHost)
			}
			if info.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", info.Owner, tt.wantOwner)
			}
			if info.Repository != tt.wantRepo {
				t.Errorf("Repository = %q, want %q", info.Repository, tt.wantRepo)
			}
			// For HTTPS URLs, Remote should be the clean URL (no credentials)
			if tt.wantRemote != "" {
				if info.Remote != tt.wantRemote {
					t.Errorf("Remote = %q, want %q", info.Remote, tt.wantRemote)
				}
			}
			// For SSH URLs, Remote should be the original URL
			if tt.wantProto == "ssh" && info.Remote != tt.url {
				t.Errorf("Remote = %q, want %q", info.Remote, tt.url)
			}
		})
	}
}

func TestGetRepoInfo(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "https://github.com/emfi/release-it-go.git", nil
	}

	info, err := GetRepoInfo("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Owner != "emfi" {
		t.Errorf("Owner = %q, want 'emfi'", info.Owner)
	}
	if info.Repository != "release-it-go" {
		t.Errorf("Repository = %q, want 'release-it-go'", info.Repository)
	}
}

func TestGetRepoInfo_CustomRemote(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	var capturedArgs []string
	commandExecutor = func(name string, args ...string) (string, error) {
		capturedArgs = args
		return "git@github.com:other/repo.git", nil
	}

	_, err := GetRepoInfo("upstream")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedArgs) < 3 || capturedArgs[2] != "upstream" {
		t.Errorf("expected 'upstream' as remote name, got args: %v", capturedArgs)
	}
}

func TestGetRepoInfo_Error(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", fmt.Errorf("no remote configured")
	}

	_, err := GetRepoInfo("")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetBranchName(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "  main  ", nil
	}

	branch, err := GetBranchName()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" {
		t.Errorf("expected 'main', got %q", branch)
	}
}

func TestGetBranchName_Error(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", fmt.Errorf("not a git repo")
	}

	_, err := GetBranchName()
	if err == nil {
		t.Fatal("expected error")
	}
}
