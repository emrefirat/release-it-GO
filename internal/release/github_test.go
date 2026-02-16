package release

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emfi/release-it-go/internal/config"
	"github.com/emfi/release-it-go/internal/git"
	applog "github.com/emfi/release-it-go/internal/log"
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
		{"github.example.com", "https://github.example.com/api/v3"},
	}

	for _, tt := range tests {
		got := resolveGitHubBaseURL(tt.host)
		if got != tt.want {
			t.Errorf("resolveGitHubBaseURL(%q) = %q, want %q", tt.host, got, tt.want)
		}
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
			w.Write([]byte(`{"login":"testuser"}`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"Bad credentials"}`))
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
		json.NewDecoder(r.Body).Decode(&req)

		resp := githubCreateReleaseResponse{
			ID:        42,
			HTMLURL:   "https://github.com/testowner/testrepo/releases/tag/v1.0.0",
			UploadURL: "https://uploads.github.com/repos/testowner/testrepo/releases/42/assets{?name,label}",
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
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
		w.Write([]byte(`{"message":"Validation Failed"}`))
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
		w.Write([]byte(`{"id":1}`))
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
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			c.client = server.Client()
			c.baseURL = server.URL

			resp, _ := c.doRequest("GET", server.URL+"/test", nil)
			err := c.handleErrorResponse(resp, "test")
			resp.Body.Close()

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
	os.Setenv("TEST_GH_TOKEN", "test-token-123")
	defer os.Unsetenv("TEST_GH_TOKEN")

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
	os.Unsetenv("MISSING_TOKEN")

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
		w.Write([]byte(`{"id":1,"name":"test.zip"}`))
	}))
	defer server.Close()

	// Create a temp file to upload
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test.zip"
	os.WriteFile(testFile, []byte("fake zip content"), 0644)

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
		json.NewDecoder(r.Body).Decode(&receivedReq)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(githubCreateReleaseResponse{ID: 1, HTMLURL: "https://example.com"})
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
