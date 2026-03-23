package release

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"release-it-go/internal/config"
	"release-it-go/internal/git"
	applog "release-it-go/internal/log"
)

// GitLabClient implements ReleaseProvider for GitLab.
type GitLabClient struct {
	config    *config.GitLabConfig
	repoInfo  *git.RepoInfo
	logger    *applog.Logger
	dryRun    bool
	client    *http.Client
	baseURL   string
	token     string
	projectID string // URL-encoded "owner/repo"
}

// gitlabCreateReleaseRequest is the GitLab API request body for creating a release.
type gitlabCreateReleaseRequest struct {
	Name        string   `json:"name"`
	TagName     string   `json:"tag_name"`
	Description string   `json:"description"`
	Milestones  []string `json:"milestones,omitempty"`
}

// gitlabCreateReleaseResponse is the GitLab API response for creating a release.
type gitlabCreateReleaseResponse struct {
	TagName     string `json:"tag_name"`
	Description string `json:"description"`
	Links       struct {
		Self string `json:"self"`
	} `json:"_links"`
}

// gitlabReleaseLinkRequest creates a link to an uploaded asset.
type gitlabReleaseLinkRequest struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	LinkType string `json:"link_type"`
}

// gitlabCommentRequest is the GitLab API request for posting a note.
type gitlabCommentRequest struct {
	Body string `json:"body"`
}

// NewGitLabClient creates a new GitLab API client.
func NewGitLabClient(cfg *config.GitLabConfig, repoInfo *git.RepoInfo, logger *applog.Logger, dryRun bool) (*GitLabClient, error) {
	token, err := getToken(cfg.TokenRef, cfg.SkipChecks)
	if err != nil {
		return nil, err
	}

	c := &GitLabClient{
		config:    cfg,
		repoInfo:  repoInfo,
		logger:    logger,
		dryRun:    dryRun,
		token:     token,
		baseURL:   resolveGitLabBaseURL(cfg.Origin, repoInfo),
		projectID: url.PathEscape(repoInfo.Owner + "/" + repoInfo.Repository),
	}

	c.client = c.createHTTPClient()
	return c, nil
}

// resolveGitLabBaseURL determines the GitLab API base URL.
func resolveGitLabBaseURL(origin string, repoInfo *git.RepoInfo) string {
	if origin != "" {
		return strings.TrimRight(origin, "/") + "/api/v4"
	}
	return fmt.Sprintf("https://%s/api/v4", repoInfo.Host)
}

// createHTTPClient creates an HTTP client with TLS and CA certificate support.
func (c *GitLabClient) createHTTPClient() *http.Client {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	caFile := c.config.CertificateAuthorityFile
	if caFile == "" && c.config.CertificateAuthorityFileRef != "" {
		caFile = os.Getenv(c.config.CertificateAuthorityFileRef)
	}

	if caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			c.logger.Warn("failed to load CA certificate %s: %v", caFile, err)
		} else {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
	}

	if !c.config.Secure {
		tlsConfig.InsecureSkipVerify = true //nolint:gosec // User explicitly disabled TLS verification
	}

	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   30 * time.Second,
	}
}

// CreateRelease creates a new release on GitLab.
func (c *GitLabClient) CreateRelease(opts ReleaseOptions) (*ReleaseResult, error) {
	endpoint := fmt.Sprintf("%s/projects/%s/releases", c.baseURL, c.projectID)

	if c.dryRun {
		c.logger.DryRun("POST %s", endpoint)
		c.logger.DryRun("  tag_name: %s", opts.TagName)
		c.logger.DryRun("  name: %s", opts.ReleaseName)
		return &ReleaseResult{ID: opts.TagName, URL: "(dry-run)"}, nil
	}

	reqBody := gitlabCreateReleaseRequest{
		Name:        opts.ReleaseName,
		TagName:     opts.TagName,
		Description: opts.ReleaseNotes,
	}

	// Add milestones from config
	if len(c.config.Milestones) > 0 {
		reqBody.Milestones = c.config.Milestones
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling release request: %w", err)
	}

	resp, err := c.doRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp, "creating release")
	}

	var result gitlabCreateReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding release response: %w", err)
	}

	return &ReleaseResult{
		ID:  result.TagName,
		URL: result.Links.Self,
	}, nil
}

// UploadAssets uploads files to a GitLab release using Generic Package Repository.
func (c *GitLabClient) UploadAssets(releaseID string, assets []string) error {
	if c.dryRun {
		for _, asset := range assets {
			c.logger.DryRun("Upload asset: %s (via Generic Package)", asset)
		}
		return nil
	}

	for _, assetPath := range assets {
		packageURL, err := c.uploadToGenericPackage(releaseID, assetPath)
		if err != nil {
			return fmt.Errorf("uploading asset %s: %w", assetPath, err)
		}

		filename := filepath.Base(assetPath)
		if err := c.createReleaseLink(releaseID, filename, packageURL); err != nil {
			return fmt.Errorf("creating release link for %s: %w", assetPath, err)
		}
	}
	return nil
}

// uploadToGenericPackage uploads a file to the GitLab Generic Package Registry.
func (c *GitLabClient) uploadToGenericPackage(tagName string, assetPath string) (string, error) {
	f, err := os.Open(assetPath)
	if err != nil {
		return "", fmt.Errorf("asset file not found: %s", assetPath)
	}
	defer func() { _ = f.Close() }()

	filename := filepath.Base(assetPath)
	packageName := c.repoInfo.Repository
	version := strings.TrimPrefix(tagName, "v")

	endpoint := fmt.Sprintf("%s/projects/%s/packages/generic/%s/%s/%s",
		c.baseURL, c.projectID, packageName, version, url.PathEscape(filename))

	req, err := http.NewRequest("PUT", endpoint, f)
	if err != nil {
		return "", fmt.Errorf("creating upload request: %w", err)
	}

	c.setAuthHeader(req)
	req.Header.Set("Content-Type", DetectContentType(assetPath))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("uploading to generic package: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return "", c.handleErrorResponse(resp, "uploading to generic package")
	}

	c.logger.Info("Uploaded asset to generic package: %s", filename)
	return endpoint, nil
}

// createReleaseLink creates a link in a GitLab release to an uploaded asset.
func (c *GitLabClient) createReleaseLink(tagName string, name string, assetURL string) error {
	endpoint := fmt.Sprintf("%s/projects/%s/releases/%s/assets/links",
		c.baseURL, c.projectID, url.PathEscape(tagName))

	reqBody := gitlabReleaseLinkRequest{
		Name:     name,
		URL:      assetURL,
		LinkType: "package",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling release link: %w", err)
	}

	resp, err := c.doRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating release link: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return c.handleErrorResponse(resp, "creating release link")
	}

	return nil
}

// PostComment posts a comment on a GitLab issue or merge request.
func (c *GitLabClient) PostComment(target CommentTarget, message string) error {
	var resource string
	switch target.Type {
	case "mr":
		resource = "merge_requests"
	default:
		resource = "issues"
	}

	endpoint := fmt.Sprintf("%s/projects/%s/%s/%d/notes",
		c.baseURL, c.projectID, resource, target.Number)

	if c.dryRun {
		c.logger.DryRun("POST %s", endpoint)
		c.logger.DryRun("  body: %s", truncate(message, 100))
		return nil
	}

	body, err := json.Marshal(gitlabCommentRequest{Body: message})
	if err != nil {
		return fmt.Errorf("marshaling comment: %w", err)
	}

	resp, err := c.doRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("posting comment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return c.handleErrorResponse(resp, "posting comment")
	}

	return nil
}

// ValidateToken checks if the GitLab token is valid.
func (c *GitLabClient) ValidateToken() error {
	if c.config.SkipChecks {
		return nil
	}

	if c.token == "" {
		return fmt.Errorf("GitLab token is not set")
	}

	endpoint := fmt.Sprintf("%s/user", c.baseURL)

	if c.dryRun {
		c.logger.DryRun("GET %s (validate token)", endpoint)
		return nil
	}

	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("validating GitLab token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("GitLab token is invalid (HTTP 401)")
	}

	if resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp, "validating token")
	}

	return nil
}

// doRequest executes an HTTP request with GitLab authentication.
func (c *GitLabClient) doRequest(method, reqURL string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}

	c.setAuthHeader(req)
	if body != nil && method != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.client.Do(req)
}

// setAuthHeader sets the authentication header based on config.
func (c *GitLabClient) setAuthHeader(req *http.Request) {
	header := c.config.TokenHeader
	if header == "" {
		header = "Private-Token"
	}
	req.Header.Set(header, c.token)
}

// handleErrorResponse creates a descriptive error from an HTTP response.
func (c *GitLabClient) handleErrorResponse(resp *http.Response, context string) error {
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := strings.TrimSpace(string(bodyBytes))

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("project %s/%s not found (HTTP 404)", c.repoInfo.Owner, c.repoInfo.Repository)
	case http.StatusUnauthorized:
		return fmt.Errorf("GitLab token is invalid (HTTP 401)")
	default:
		return fmt.Errorf("%s failed (HTTP %d): %s", context, resp.StatusCode, truncate(bodyStr, 200))
	}
}
