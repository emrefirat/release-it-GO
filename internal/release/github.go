package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"release-it-go/internal/config"
	"release-it-go/internal/git"
	applog "release-it-go/internal/log"
)

// GitHubClient implements ReleaseProvider for GitHub.
type GitHubClient struct {
	config   *config.GitHubConfig
	repoInfo *git.RepoInfo
	logger   *applog.Logger
	dryRun   bool
	client   *http.Client
	baseURL  string
	token    string
}

// githubCreateReleaseRequest is the GitHub API request body for creating a release.
type githubCreateReleaseRequest struct {
	TagName                string `json:"tag_name"`
	Name                   string `json:"name"`
	Body                   string `json:"body"`
	Draft                  bool   `json:"draft"`
	Prerelease             bool   `json:"prerelease"`
	MakeLatest             string `json:"make_latest,omitempty"`
	GenerateReleaseNotes   bool   `json:"generate_release_notes,omitempty"`
	DiscussionCategoryName string `json:"discussion_category_name,omitempty"`
}

// githubCreateReleaseResponse is the GitHub API response for creating a release.
type githubCreateReleaseResponse struct {
	ID        int    `json:"id"`
	HTMLURL   string `json:"html_url"`
	UploadURL string `json:"upload_url"`
}

// githubCommentRequest is the GitHub API request body for posting a comment.
type githubCommentRequest struct {
	Body string `json:"body"`
}

// NewGitHubClient creates a new GitHub API client.
// The token is read from the environment variable specified by config.TokenRef.
func NewGitHubClient(cfg *config.GitHubConfig, repoInfo *git.RepoInfo, logger *applog.Logger, dryRun bool) (*GitHubClient, error) {
	token, err := getToken(cfg.TokenRef, cfg.SkipChecks)
	if err != nil {
		return nil, err
	}

	c := &GitHubClient{
		config:   cfg,
		repoInfo: repoInfo,
		logger:   logger,
		dryRun:   dryRun,
		token:    token,
		baseURL:  resolveGitHubBaseURL(cfg.Host),
	}

	c.client = c.createHTTPClient()
	return c, nil
}

// resolveGitHubBaseURL determines the API base URL based on the host.
func resolveGitHubBaseURL(host string) string {
	if host == "" || host == "github.com" {
		return "https://api.github.com"
	}
	// GitHub Enterprise
	return fmt.Sprintf("https://%s/api/v3", host)
}

// createHTTPClient creates an HTTP client with proxy and timeout support.
func (c *GitHubClient) createHTTPClient() *http.Client {
	transport := &http.Transport{}

	if c.config.Proxy != "" {
		proxyURL, err := url.Parse(c.config.Proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	timeout := time.Duration(c.config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}

// CreateRelease creates a new release on GitHub.
func (c *GitHubClient) CreateRelease(opts ReleaseOptions) (*ReleaseResult, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/%s/releases", c.baseURL, c.repoInfo.Owner, c.repoInfo.Repository)

	if c.dryRun {
		c.logger.DryRun("POST %s", endpoint)
		c.logger.DryRun("  tag_name: %s", opts.TagName)
		c.logger.DryRun("  name: %s", opts.ReleaseName)
		c.logger.DryRun("  draft: %v, prerelease: %v", opts.Draft, opts.PreRelease)
		return &ReleaseResult{ID: "0", URL: "(dry-run)", UploadURL: "(dry-run)"}, nil
	}

	reqBody := githubCreateReleaseRequest{
		TagName:                opts.TagName,
		Name:                   opts.ReleaseName,
		Body:                   opts.ReleaseNotes,
		Draft:                  opts.Draft,
		Prerelease:             opts.PreRelease,
		GenerateReleaseNotes:   opts.AutoGenerate,
		DiscussionCategoryName: opts.DiscussionCategory,
	}

	if opts.MakeLatest {
		reqBody.MakeLatest = "true"
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling release request: %w", err)
	}

	resp, err := c.doRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp, "creating release")
	}

	var result githubCreateReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding release response: %w", err)
	}

	// Clean upload URL template
	uploadURL := result.UploadURL
	if idx := strings.Index(uploadURL, "{"); idx > 0 {
		uploadURL = uploadURL[:idx]
	}

	return &ReleaseResult{
		ID:        strconv.Itoa(result.ID),
		URL:       result.HTMLURL,
		UploadURL: uploadURL,
	}, nil
}

// UploadAssets uploads files to an existing GitHub release.
func (c *GitHubClient) UploadAssets(releaseID string, assets []string) error {
	if c.dryRun {
		for _, asset := range assets {
			c.logger.DryRun("Upload asset: %s", asset)
		}
		return nil
	}

	for _, assetPath := range assets {
		if err := c.uploadSingleAsset(releaseID, assetPath); err != nil {
			return err
		}
	}
	return nil
}

// uploadSingleAsset uploads a single file to a GitHub release.
func (c *GitHubClient) uploadSingleAsset(releaseID string, assetPath string) error {
	f, err := os.Open(assetPath)
	if err != nil {
		return fmt.Errorf("asset file not found: %s", assetPath)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("reading asset file info: %w", err)
	}

	filename := filepath.Base(assetPath)
	uploadURL := fmt.Sprintf("%s/repos/%s/%s/releases/%s/assets?name=%s",
		strings.Replace(c.baseURL, "api.github.com", "uploads.github.com", 1),
		c.repoInfo.Owner, c.repoInfo.Repository, releaseID, url.QueryEscape(filename))

	req, err := http.NewRequest("POST", uploadURL, f)
	if err != nil {
		return fmt.Errorf("creating upload request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", DetectContentType(assetPath))
	req.ContentLength = stat.Size()

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("uploading asset %s: %w", filename, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return c.handleErrorResponse(resp, fmt.Sprintf("uploading asset %s", filename))
	}

	c.logger.Info("Uploaded asset: %s", filename)
	return nil
}

// PostComment posts a comment on a GitHub issue or PR.
func (c *GitHubClient) PostComment(target CommentTarget, message string) error {
	// GitHub uses the issues API for both issues and PRs
	endpoint := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments",
		c.baseURL, c.repoInfo.Owner, c.repoInfo.Repository, target.Number)

	if c.dryRun {
		c.logger.DryRun("POST %s", endpoint)
		c.logger.DryRun("  body: %s", truncate(message, 100))
		return nil
	}

	body, err := json.Marshal(githubCommentRequest{Body: message})
	if err != nil {
		return fmt.Errorf("marshaling comment: %w", err)
	}

	resp, err := c.doRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("posting comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return c.handleErrorResponse(resp, "posting comment")
	}

	return nil
}

// ValidateToken checks if the GitHub token is valid.
func (c *GitHubClient) ValidateToken() error {
	if c.config.SkipChecks {
		return nil
	}

	if c.token == "" {
		return fmt.Errorf("GitHub token is not set")
	}

	endpoint := fmt.Sprintf("%s/user", c.baseURL)

	if c.dryRun {
		c.logger.DryRun("GET %s (validate token)", endpoint)
		return nil
	}

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("validating GitHub token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitHub token is invalid (HTTP 401)")
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp, "validating token")
	}

	return nil
}

// doRequest executes an HTTP request with authentication headers.
func (c *GitHubClient) doRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	if body != nil && method != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.client.Do(req)
}

// handleErrorResponse creates a descriptive error from an HTTP response.
func (c *GitHubClient) handleErrorResponse(resp *http.Response, context string) error {
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := strings.TrimSpace(string(bodyBytes))

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("repository %s/%s not found (HTTP 404)", c.repoInfo.Owner, c.repoInfo.Repository)
	case http.StatusUnauthorized:
		return fmt.Errorf("GitHub token is invalid (HTTP 401)")
	case http.StatusForbidden:
		if strings.Contains(bodyStr, "rate limit") {
			retryAfter := resp.Header.Get("Retry-After")
			return fmt.Errorf("GitHub API rate limit exceeded, retry after %s", retryAfter)
		}
		return fmt.Errorf("GitHub API forbidden (HTTP 403): %s", truncate(bodyStr, 200))
	default:
		return fmt.Errorf("%s failed (HTTP %d): %s", context, resp.StatusCode, truncate(bodyStr, 200))
	}
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
