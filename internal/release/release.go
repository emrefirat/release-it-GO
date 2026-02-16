// Package release provides GitHub and GitLab release management
// using direct REST API calls via net/http.
package release

import (
	"fmt"
	"os"
)

// ReleaseProvider defines the interface for creating releases on remote platforms.
type ReleaseProvider interface {
	// CreateRelease creates a new release on the platform.
	CreateRelease(opts ReleaseOptions) (*ReleaseResult, error)

	// UploadAssets uploads asset files to an existing release.
	UploadAssets(releaseID string, assets []string) error

	// PostComment posts a comment on an issue, PR, or MR.
	PostComment(target CommentTarget, message string) error

	// ValidateToken checks if the API token is valid.
	ValidateToken() error
}

// ReleaseOptions holds parameters for creating a release.
type ReleaseOptions struct {
	TagName            string
	ReleaseName        string
	ReleaseNotes       string
	Draft              bool
	PreRelease         bool
	MakeLatest         bool
	AutoGenerate       bool   // GitHub auto-generate release notes
	DiscussionCategory string // GitHub discussions category
}

// ReleaseResult holds the result of a release creation.
type ReleaseResult struct {
	ID        string // release ID
	URL       string // release web URL
	UploadURL string // asset upload URL (GitHub)
}

// CommentTarget identifies where to post a comment.
type CommentTarget struct {
	Type   string // "issue", "pr", "mr"
	Number int
}

// getToken reads an API token from the environment variable specified by tokenRef.
// Returns an error if the token is empty and skipChecks is false.
func getToken(tokenRef string, skipChecks bool) (string, error) {
	if tokenRef == "" {
		if skipChecks {
			return "", nil
		}
		return "", fmt.Errorf("token reference is not configured")
	}

	token := os.Getenv(tokenRef)
	if token == "" && !skipChecks {
		return "", fmt.Errorf("environment variable %s is not set; "+
			"create a token and set it as %s", tokenRef, tokenRef)
	}
	return token, nil
}
