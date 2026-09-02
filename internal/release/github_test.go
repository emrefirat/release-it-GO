package release

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"release-it-go/internal/config"
	"release-it-go/internal/git"
	applog "release-it-go/internal/log"
)

func testRepoInfo() *git.RepoInfo {
	return &git.RepoInfo{
		Host:       "github.com",
		Owner:      "testowner",
		Repository: "testrepo",
		Protocol:   "https",
	}
}

func testLogger() *applog.Logger {
	return applog.NewLogger(0, false)
}

func TestResolveGitHubBaseURL(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{"", "https://api.github.com"},
		{"github.com", "https://api.github.com"},
		{"api.github.com", "https://api.github.com"},
		{"github.example.com", "https://github.example.com/api/v3"},
	}

	for _, tt := range tests {
		got := resolveGitHubBaseURL(tt.host)
		if got != tt.want {
			t.Errorf("resolveGitHubBaseURL(%q) = %q, want %q", tt.host, got, tt.want)
		}
	}
}

func TestNewGitHubClient_DefaultConfig_UsesPublicAPIBaseURL(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "test-token")
	cfg := config.DefaultConfig()

	c, err := NewGitHubClient(&cfg.GitHub, testRepoInfo(), testLogger(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Regression guard: the shipped default config must target the public
	// GitHub API, never the Enterprise-style https://.../api/v3 path.
	if c.baseURL != "https://api.github.com" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://api.github.com")
	}
}

func TestGitHubClient_ValidateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		auth := r.Header.Get("Authorization")
		if auth == "token valid-token" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"login":"testuser"}`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
		}
	}))
	defer server.Close()

	t.Run("valid token", func(t *testing.T) {
		c := &GitHubClient{
			config:   &config.GitHubConfig{},
			repoInfo: testRepoInfo(),
			logger:   testLogger(),
			client:   server.Client(),
			baseURL:  server.URL,
			token:    "valid-token",
		}

		err := c.ValidateToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		c := &GitHubClient{
			config:   &config.GitHubConfig{},
			repoInfo: testRepoInfo(),
			logger:   testLogger(),
			client:   server.Client(),
			baseURL:  server.URL,
			token:    "invalid-token",
		}

		err := c.ValidateToken()
		if err == nil {
			t.Error("expected error for invalid token")
		}
		if !strings.Contains(err.Error(), "401") {
			t.Errorf("expected 401 error, got: %v", err)
		}
	})

	t.Run("skip checks", func(t *testing.T) {
		c := &GitHubClient{
			config:   &config.GitHubConfig{SkipChecks: true},
			repoInfo: testRepoInfo(),
			logger:   testLogger(),
			client:   server.Client(),
			baseURL:  server.URL,
			token:    "",
		}

		err := c.ValidateToken()
		if err != nil {
			t.Errorf("unexpected error with skipChecks: %v", err)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		c := &GitHubClient{
			config:   &config.GitHubConfig{},
			repoInfo: testRepoInfo(),
			logger:   testLogger(),
			client:   server.Client(),
			baseURL:  server.URL,
			token:    "",
		}

		err := c.ValidateToken()
		if err == nil {
			t.Error("expected error for empty token")
		}
	})
}

func TestGitHubClient_CreateRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/releases") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req githubCreateReleaseRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		resp := githubCreateReleaseResponse{
			ID:        42,
			HTMLURL:   "https://github.com/testowner/testrepo/releases/tag/v1.0.0",
			UploadURL: "https://uploads.github.com/repos/testowner/testrepo/releases/42/assets{?name,label}",
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		client:   server.Client(),
		baseURL:  server.URL,
		token:    "test-token",
	}

	result, err := c.CreateRelease(ReleaseOptions{
		TagName:      "v1.0.0",
		ReleaseName:  "Release v1.0.0",
		ReleaseNotes: "Initial release",
		Draft:        false,
		PreRelease:   false,
		MakeLatest:   true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "42" {
		t.Errorf("ID = %q, want '42'", result.ID)
	}
	if !strings.Contains(result.URL, "releases/tag/v1.0.0") {
		t.Errorf("URL = %q, expected release URL", result.URL)
	}
	// Upload URL should have template stripped
	if strings.Contains(result.UploadURL, "{") {
		t.Errorf("UploadURL should not contain template: %q", result.UploadURL)
	}
}

func TestGitHubClient_CreateRelease_DryRun(t *testing.T) {
	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		dryRun:   true,
		token:    "test-token",
		baseURL:  "https://api.github.com",
	}

	result, err := c.CreateRelease(ReleaseOptions{
		TagName:     "v1.0.0",
		ReleaseName: "Release v1.0.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.URL != "(dry-run)" {
		t.Errorf("expected dry-run URL, got %q", result.URL)
	}
}

func TestGitHubClient_CreateRelease_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"Validation Failed"}`))
	}))
	defer server.Close()

	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		client:   server.Client(),
		baseURL:  server.URL,
		token:    "test-token",
	}

	_, err := c.CreateRelease(ReleaseOptions{TagName: "v1.0.0"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestGitHubClient_PostComment(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer server.Close()

	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		client:   server.Client(),
		baseURL:  server.URL,
		token:    "test-token",
	}

	err := c.PostComment(CommentTarget{Type: "pr", Number: 42}, "Release v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(receivedPath, "/issues/42/comments") {
		t.Errorf("expected issues API path, got %q", receivedPath)
	}
}

func TestGitHubClient_PostComment_DryRun(t *testing.T) {
	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		dryRun:   true,
		token:    "test-token",
		baseURL:  "https://api.github.com",
	}

	err := c.PostComment(CommentTarget{Type: "issue", Number: 1}, "test comment")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGitHubClient_UploadAssets_DryRun(t *testing.T) {
	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		dryRun:   true,
		token:    "test-token",
		baseURL:  "https://api.github.com",
	}

	err := c.UploadAssets("42", []string{"file1.zip", "file2.tar.gz"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGitHubClient_UploadAssets_FileNotFound(t *testing.T) {
	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		client:   http.DefaultClient,
		baseURL:  "https://api.github.com",
		token:    "test-token",
	}

	err := c.UploadAssets("42", []string{"/nonexistent/file.zip"})
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestGitHubClient_HandleErrorResponse(t *testing.T) {
	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
	}

	tests := []struct {
		name       string
		statusCode int
		body       string
		wantMsg    string
	}{
		{"not found", 404, "", "not found"},
		{"unauthorized", 401, "", "401"},
		{"rate limit", 403, "rate limit exceeded", "rate limit"},
		{"other error", 500, "internal error", "HTTP 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c.client = server.Client()
			c.baseURL = server.URL

			resp, _ := c.doRequest("GET", server.URL+"/test", nil)
			err := c.handleErrorResponse(resp, "test")
			_ = resp.Body.Close()

			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("expected error containing %q, got: %v", tt.wantMsg, err)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is a ..."},
		{"", 5, ""},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestNewGitHubClient(t *testing.T) {
	_ = os.Setenv("TEST_GH_TOKEN", "test-token-123")
	defer func() { _ = os.Unsetenv("TEST_GH_TOKEN") }()

	cfg := &config.GitHubConfig{
		TokenRef: "TEST_GH_TOKEN",
		Timeout:  10,
	}

	client, err := NewGitHubClient(cfg, testRepoInfo(), testLogger(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.token != "test-token-123" {
		t.Errorf("token = %q, want 'test-token-123'", client.token)
	}
	if client.baseURL != "https://api.github.com" {
		t.Errorf("baseURL = %q, want 'https://api.github.com'", client.baseURL)
	}
}

func TestNewGitHubClient_MissingToken(t *testing.T) {
	_ = os.Unsetenv("MISSING_TOKEN")

	cfg := &config.GitHubConfig{
		TokenRef: "MISSING_TOKEN",
	}

	_, err := NewGitHubClient(cfg, testRepoInfo(), testLogger(), false)
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestGitHubClient_UploadAssets_Success(t *testing.T) {
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"name":"test.zip"}`))
	}))
	defer server.Close()

	// Create a temp file to upload
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.zip"
	_ = os.WriteFile(testFile, []byte("fake zip content"), 0644)

	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		client:   server.Client(),
		baseURL:  server.URL,
		token:    "test-token",
	}

	err := c.UploadAssets("42", []string{testFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedContentType != "application/zip" {
		t.Errorf("Content-Type = %q, want 'application/zip'", receivedContentType)
	}
}

func TestGitHubClient_CreateRelease_WithAllOptions(t *testing.T) {
	var receivedReq githubCreateReleaseRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedReq)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(githubCreateReleaseResponse{ID: 1, HTMLURL: "https://example.com"})
	}))
	defer server.Close()

	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		client:   server.Client(),
		baseURL:  server.URL,
		token:    "test-token",
	}

	_, err := c.CreateRelease(ReleaseOptions{
		TagName:            "v2.0.0",
		ReleaseName:        "Release v2.0.0",
		ReleaseNotes:       "Notes",
		Draft:              true,
		PreRelease:         true,
		MakeLatest:         true,
		AutoGenerate:       true,
		DiscussionCategory: "Announcements",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !receivedReq.Draft {
		t.Error("expected draft=true")
	}
	if !receivedReq.Prerelease {
		t.Error("expected prerelease=true")
	}
	if receivedReq.MakeLatest != "true" {
		t.Errorf("MakeLatest = %q, want 'true'", receivedReq.MakeLatest)
	}
	if !receivedReq.GenerateReleaseNotes {
		t.Error("expected generate_release_notes=true")
	}
	if receivedReq.DiscussionCategoryName != "Announcements" {
		t.Errorf("DiscussionCategoryName = %q, want 'Announcements'", receivedReq.DiscussionCategoryName)
	}
}

func TestGitHubClient_ValidateToken_DryRun(t *testing.T) {
	c := &GitHubClient{
		config:   &config.GitHubConfig{},
		repoInfo: testRepoInfo(),
		logger:   testLogger(),
		dryRun:   true,
		token:    "test-token",
		baseURL:  "https://api.github.com",
	}

	err := c.ValidateToken()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGitHubClient_CreateHTTPClient_DefaultTimeout(t *testing.T) {
	c := &GitHubClient{
		config: &config.GitHubConfig{},
		logger: testLogger(),
	}
	client := c.createHTTPClient()
	if client.Timeout != 30*time.Second {
		t.Errorf("default timeout = %v, want 30s", client.Timeout)
	}
}

func TestGitHubClient_CreateHTTPClient_CustomTimeout(t *testing.T) {
	c := &GitHubClient{
		config: &config.GitHubConfig{Timeout: 60},
		logger: testLogger(),
	}
	client := c.createHTTPClient()
	if client.Timeout != 60*time.Second {
		t.Errorf("timeout = %v, want 60s", client.Timeout)
	}
}

func TestGitHubClient_CreateHTTPClient_WithProxy(t *testing.T) {
	c := &GitHubClient{
		config: &config.GitHubConfig{Proxy: "http://proxy.example.com:8080"},
		logger: testLogger(),
	}
	client := c.createHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if transport.Proxy == nil {
		t.Error("expected proxy to be configured")
	}
}

func TestGitHubClient_CreateHTTPClient_InvalidProxy(t *testing.T) {
	c := &GitHubClient{
		config: &config.GitHubConfig{Proxy: "://invalid"},
		logger: testLogger(),
	}
	// Should not panic, should log warning and continue without proxy
	client := c.createHTTPClient()
	if client == nil {
		t.Fatal("expected non-nil client even with invalid proxy")
	}
}

func TestGitHubClient_ValidateToken_RetriesTransient503(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &GitHubClient{
		config:  &config.GitHubConfig{},
		client:  server.Client(),
		baseURL: server.URL,
		token:   "test-token",
		logger:  testLogger(),
	}

	if err := c.ValidateToken(); err != nil {
		t.Fatalf("expected transient 503 to be retried, got: %v", err)
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestGitHubClient_CreateHTTPClient_HonorsProxyEnvironment(t *testing.T) {
	t.Setenv("HTTPS_PROXY", "http://proxy.example:3128")
	t.Setenv("NO_PROXY", "")
	c := &GitHubClient{config: &config.GitHubConfig{}, logger: testLogger()}

	client := c.createHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.Proxy == nil {
		t.Fatal("transport must consult the proxy environment (HTTPS_PROXY/NO_PROXY) like the notification client does")
	}
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	proxyURL, err := transport.Proxy(req)
	if err != nil || proxyURL == nil || proxyURL.Host != "proxy.example:3128" {
		t.Errorf("Proxy(req) = %v, %v; want proxy.example:3128", proxyURL, err)
	}
}

func TestGitHubClient_CreateRelease_SendsMakeLatestFalseExplicitly(t *testing.T) {
	var body map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1,"html_url":"u","upload_url":"x{?name}"}`))
	}))
	defer server.Close()

	c := &GitHubClient{
		config: &config.GitHubConfig{}, repoInfo: testRepoInfo(), logger: testLogger(),
		client: server.Client(), baseURL: server.URL, token: "t",
	}
	if _, err := c.CreateRelease(ReleaseOptions{TagName: "v1.0.1", MakeLatest: false}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// GitHub's server-side default is "true": omitting the field makes a
	// support-branch release the repository's "Latest".
	if body["make_latest"] != "false" {
		t.Errorf(`make_latest = %v, want the literal "false" on the wire`, body["make_latest"])
	}
}
