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

func testGitLabRepoInfo() *git.RepoInfo {
	return &git.RepoInfo{
		Host:       "gitlab.com",
		Owner:      "testgroup",
		Repository: "testproject",
		Protocol:   "https",
	}
}

func TestResolveGitLabBaseURL(t *testing.T) {
	tests := []struct {
		origin string
		host   string
		want   string
	}{
		{"", "gitlab.com", "https://gitlab.com/api/v4"},
		{"https://gitlab.example.com", "gitlab.example.com", "https://gitlab.example.com/api/v4"},
		{"https://custom.host/", "gitlab.com", "https://custom.host/api/v4"},
	}

	for _, tt := range tests {
		repo := &git.RepoInfo{Host: tt.host}
		got := resolveGitLabBaseURL(tt.origin, repo)
		if got != tt.want {
			t.Errorf("resolveGitLabBaseURL(%q) = %q, want %q", tt.origin, got, tt.want)
		}
	}
}

func TestGitLabClient_ValidateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		token := r.Header.Get("Private-Token")
		if token == "valid-token" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"username":"testuser"}`))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	t.Run("valid token", func(t *testing.T) {
		c := &GitLabClient{
			config:    &config.GitLabConfig{},
			repoInfo:  testGitLabRepoInfo(),
			logger:    applog.NewLogger(0, false),
			client:    server.Client(),
			baseURL:   server.URL,
			token:     "valid-token",
			projectID: "testgroup%2Ftestproject",
		}

		err := c.ValidateToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		c := &GitLabClient{
			config:    &config.GitLabConfig{},
			repoInfo:  testGitLabRepoInfo(),
			logger:    applog.NewLogger(0, false),
			client:    server.Client(),
			baseURL:   server.URL,
			token:     "bad-token",
			projectID: "testgroup%2Ftestproject",
		}

		err := c.ValidateToken()
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("skip checks", func(t *testing.T) {
		c := &GitLabClient{
			config:   &config.GitLabConfig{SkipChecks: true},
			repoInfo: testGitLabRepoInfo(),
			logger:   applog.NewLogger(0, false),
		}

		err := c.ValidateToken()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestGitLabClient_CreateRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !strings.Contains(r.URL.Path, "/releases") {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Verify Private-Token header
		if r.Header.Get("Private-Token") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req gitlabCreateReleaseRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := gitlabCreateReleaseResponse{
			TagName:     req.TagName,
			Description: req.Description,
		}
		resp.Links.Self = "https://gitlab.com/testgroup/testproject/-/releases/" + req.TagName

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := &GitLabClient{
		config:    &config.GitLabConfig{},
		repoInfo:  testGitLabRepoInfo(),
		logger:    applog.NewLogger(0, false),
		client:    server.Client(),
		baseURL:   server.URL,
		token:     "test-token",
		projectID: "testgroup%2Ftestproject",
	}

	result, err := c.CreateRelease(ReleaseOptions{
		TagName:      "v1.0.0",
		ReleaseName:  "Release v1.0.0",
		ReleaseNotes: "Initial release",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "v1.0.0" {
		t.Errorf("ID = %q, want 'v1.0.0'", result.ID)
	}
}

func TestGitLabClient_CreateRelease_WithMilestones(t *testing.T) {
	var receivedReq gitlabCreateReleaseRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedReq)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(gitlabCreateReleaseResponse{TagName: "v1.0.0"})
	}))
	defer server.Close()

	c := &GitLabClient{
		config:    &config.GitLabConfig{Milestones: []string{"1.0", "Q1-2026"}},
		repoInfo:  testGitLabRepoInfo(),
		logger:    applog.NewLogger(0, false),
		client:    server.Client(),
		baseURL:   server.URL,
		token:     "test-token",
		projectID: "testgroup%2Ftestproject",
	}

	_, err := c.CreateRelease(ReleaseOptions{TagName: "v1.0.0", ReleaseName: "v1.0.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(receivedReq.Milestones) != 2 {
		t.Errorf("expected 2 milestones, got %d", len(receivedReq.Milestones))
	}
}

func TestGitLabClient_CreateRelease_DryRun(t *testing.T) {
	c := &GitLabClient{
		config:    &config.GitLabConfig{},
		repoInfo:  testGitLabRepoInfo(),
		logger:    applog.NewLogger(0, false),
		dryRun:    true,
		token:     "test-token",
		baseURL:   "https://gitlab.com/api/v4",
		projectID: "testgroup%2Ftestproject",
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

func TestGitLabClient_PostComment(t *testing.T) {
	tests := []struct {
		name         string
		targetType   string
		wantResource string
	}{
		{"merge request", "mr", "merge_requests"},
		{"issue", "issue", "issues"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedPath = r.URL.Path
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id":1}`))
			}))
			defer server.Close()

			c := &GitLabClient{
				config:    &config.GitLabConfig{},
				repoInfo:  testGitLabRepoInfo(),
				logger:    applog.NewLogger(0, false),
				client:    server.Client(),
				baseURL:   server.URL,
				token:     "test-token",
				projectID: "testgroup%2Ftestproject",
			}

			err := c.PostComment(CommentTarget{Type: tt.targetType, Number: 5}, "test comment")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(receivedPath, tt.wantResource) {
				t.Errorf("path = %q, want to contain %q", receivedPath, tt.wantResource)
			}
		})
	}
}

func TestGitLabClient_PostComment_DryRun(t *testing.T) {
	c := &GitLabClient{
		config:    &config.GitLabConfig{},
		repoInfo:  testGitLabRepoInfo(),
		logger:    applog.NewLogger(0, false),
		dryRun:    true,
		token:     "test-token",
		baseURL:   "https://gitlab.com/api/v4",
		projectID: "testgroup%2Ftestproject",
	}

	err := c.PostComment(CommentTarget{Type: "mr", Number: 1}, "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGitLabClient_UploadAssets_DryRun(t *testing.T) {
	c := &GitLabClient{
		config:    &config.GitLabConfig{},
		repoInfo:  testGitLabRepoInfo(),
		logger:    applog.NewLogger(0, false),
		dryRun:    true,
		token:     "test-token",
		baseURL:   "https://gitlab.com/api/v4",
		projectID: "testgroup%2Ftestproject",
	}

	err := c.UploadAssets("v1.0.0", []string{"dist/app.zip"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGitLabClient_SetAuthHeader(t *testing.T) {
	tests := []struct {
		name        string
		tokenHeader string
		wantHeader  string
	}{
		{"default", "", "Private-Token"},
		{"custom", "Authorization", "Authorization"},
		{"job token", "JOB-TOKEN", "JOB-TOKEN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &GitLabClient{
				config: &config.GitLabConfig{TokenHeader: tt.tokenHeader},
				token:  "test-token",
			}

			req, _ := http.NewRequest("GET", "https://example.com", nil)
			c.setAuthHeader(req)

			got := req.Header.Get(tt.wantHeader)
			if got != "test-token" {
				t.Errorf("header %q = %q, want 'test-token'", tt.wantHeader, got)
			}
		})
	}
}

func TestNewGitLabClient(t *testing.T) {
	os.Setenv("TEST_GL_TOKEN", "gitlab-token-123")
	defer os.Unsetenv("TEST_GL_TOKEN")

	cfg := &config.GitLabConfig{
		TokenRef: "TEST_GL_TOKEN",
	}

	client, err := NewGitLabClient(cfg, testGitLabRepoInfo(), applog.NewLogger(0, false), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.token != "gitlab-token-123" {
		t.Errorf("token = %q, want 'gitlab-token-123'", client.token)
	}
	if !strings.Contains(client.projectID, "testgroup") {
		t.Errorf("projectID = %q, expected to contain 'testgroup'", client.projectID)
	}
}

func TestGitLabClient_HandleErrorResponse(t *testing.T) {
	c := &GitLabClient{
		config:   &config.GitLabConfig{},
		repoInfo: testGitLabRepoInfo(),
		logger:   applog.NewLogger(0, false),
	}

	tests := []struct {
		name       string
		statusCode int
		wantMsg    string
	}{
		{"not found", 404, "not found"},
		{"unauthorized", 401, "401"},
		{"server error", 500, "HTTP 500"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("error body"))
			}))
			defer server.Close()

			c.client = server.Client()
			c.baseURL = server.URL
			c.token = "test"

			resp, _ := c.doRequest("GET", server.URL+"/test", nil)
			err := c.handleErrorResponse(resp, "test")
			resp.Body.Close()

			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("expected %q in error, got: %v", tt.wantMsg, err)
			}
		})
	}
}

func TestGitLabClient_EmptyToken(t *testing.T) {
	c := &GitLabClient{
		config:   &config.GitLabConfig{},
		repoInfo: testGitLabRepoInfo(),
		logger:   applog.NewLogger(0, false),
		token:    "",
	}

	err := c.ValidateToken()
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestGitLabClient_ValidateToken_DryRun(t *testing.T) {
	c := &GitLabClient{
		config:    &config.GitLabConfig{},
		repoInfo:  testGitLabRepoInfo(),
		logger:    applog.NewLogger(0, false),
		dryRun:    true,
		token:     "test-token",
		baseURL:   "https://gitlab.com/api/v4",
		projectID: "testgroup%2Ftestproject",
	}

	err := c.ValidateToken()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
