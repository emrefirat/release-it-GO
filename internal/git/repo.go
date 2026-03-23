package git

import (
	"fmt"
	"regexp"
	"strings"
)

// RepoInfo holds parsed repository information from a git remote URL.
type RepoInfo struct {
	Remote     string // full remote URL
	Protocol   string // "https" or "ssh"
	Host       string // "github.com"
	Owner      string // "emfi"
	Repository string // "release-it-go"
}

// httpsPattern matches HTTPS remote URLs like https://github.com/owner/repo.git
// Also handles URLs with credentials like https://user:token@github.com/owner/repo.git
// Supports nested groups: https://gitlab.com/group/subgroup/repo.git
var httpsPattern = regexp.MustCompile(`^https?://(?:[^@]+@)?([^/]+)/(.+)/([^/]+?)(?:\.git)?$`)

// sshPattern matches SSH remote URLs like git@github.com:owner/repo.git
// Supports nested groups: git@gitlab.com:group/subgroup/repo.git
var sshPattern = regexp.MustCompile(`^git@([^:]+):(.+)/([^/]+?)(?:\.git)?$`)

// sshURLPattern matches ssh:// protocol URLs like ssh://git@host:22/owner/repo.git
var sshURLPattern = regexp.MustCompile(`^ssh://(?:[^@]+@)?([^:/]+)(?::\d+)?/(.+)/([^/]+?)(?:\.git)?$`)

// GetRepoInfo parses repository information from the given git remote.
func GetRepoInfo(remoteName string) (*RepoInfo, error) {
	if remoteName == "" {
		remoteName = "origin"
	}

	url, err := commandExecutor("git", "remote", "get-url", remoteName)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL for %s: %w", remoteName, err)
	}

	url = strings.TrimSpace(url)
	return ParseRepoURL(url)
}

// ParseRepoURL parses a git remote URL into RepoInfo.
// Supports both HTTPS and SSH formats.
// Credentials in HTTPS URLs (user:token@host) are stripped for security.
func ParseRepoURL(url string) (*RepoInfo, error) {
	if matches := httpsPattern.FindStringSubmatch(url); matches != nil {
		// Build a clean remote URL without credentials
		cleanRemote := fmt.Sprintf("https://%s/%s/%s", matches[1], matches[2], matches[3])
		return &RepoInfo{
			Remote:     cleanRemote,
			Protocol:   "https",
			Host:       matches[1],
			Owner:      matches[2],
			Repository: matches[3],
		}, nil
	}

	if matches := sshPattern.FindStringSubmatch(url); matches != nil {
		return &RepoInfo{
			Remote:     url,
			Protocol:   "ssh",
			Host:       matches[1],
			Owner:      matches[2],
			Repository: matches[3],
		}, nil
	}

	if matches := sshURLPattern.FindStringSubmatch(url); matches != nil {
		return &RepoInfo{
			Remote:     url,
			Protocol:   "ssh",
			Host:       matches[1],
			Owner:      matches[2],
			Repository: matches[3],
		}, nil
	}

	return nil, fmt.Errorf("unsupported remote URL format: %s", url)
}

// GetBranchName returns the current branch name.
func GetBranchName() (string, error) {
	out, err := commandExecutor("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}
