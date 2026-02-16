package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGitRepo creates a new git repo in the given directory with an initial commit.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")

	// Create initial file and commit
	writeFile(t, filepath.Join(dir, "README.md"), "# Test Project\n")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "chore: initial commit")
}

// createCommits creates commits with the given messages in the repo.
// Each commit creates a unique file to ensure there are changes to commit.
func createCommits(t *testing.T, dir string, messages []string) {
	t.Helper()
	for _, msg := range messages {
		filename := fmt.Sprintf("file_%d_%d.txt", os.Getpid(), commitCounter)
		commitCounter++
		writeFile(t, filepath.Join(dir, filename), fmt.Sprintf("content for: %s\n", msg))
		runGit(t, dir, "add", ".")
		runGit(t, dir, "commit", "-m", msg)
	}
}

// commitCounter ensures unique file names across calls within the same test.
var commitCounter int

// createTag creates a git tag in the repo.
func createTag(t *testing.T, dir string, tag string) {
	t.Helper()
	runGit(t, dir, "tag", "-a", tag, "-m", fmt.Sprintf("Release %s", tag))
}

// assertTagExists verifies a tag exists in the repo.
func assertTagExists(t *testing.T, dir string, tag string) {
	t.Helper()
	out := runGit(t, dir, "tag", "-l", tag)
	if strings.TrimSpace(out) != tag {
		t.Errorf("expected tag %q to exist, got %q", tag, out)
	}
}

// assertTagNotExists verifies a tag does not exist in the repo.
func assertTagNotExists(t *testing.T, dir string, tag string) {
	t.Helper()
	out := runGit(t, dir, "tag", "-l", tag)
	if strings.TrimSpace(out) == tag {
		t.Errorf("expected tag %q to not exist", tag)
	}
}

// assertChangelogContains checks that CHANGELOG.md contains the expected string.
func assertChangelogContains(t *testing.T, dir string, expected string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("failed to read CHANGELOG.md: %v", err)
	}
	if !strings.Contains(string(content), expected) {
		t.Errorf("CHANGELOG.md does not contain %q, got:\n%s", expected, string(content))
	}
}

// getLatestCommitMsg returns the latest commit message in the repo.
func getLatestCommitMsg(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(runGit(t, dir, "log", "-1", "--pretty=format:%s"))
}

// runGit executes a git command in the given directory and returns stdout.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\nOutput: %s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

// writeFile writes content to a file, creating directories as needed.
func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

// mockGitHubServer creates a mock GitHub API server that records requests.
func mockGitHubServer(t *testing.T) (*httptest.Server, *requestRecorder) {
	t.Helper()
	recorder := &requestRecorder{}

	mux := http.NewServeMux()

	// Create release endpoint
	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		recorder.record(r)

		if strings.HasSuffix(r.URL.Path, "/releases") && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1,
				"html_url":   "https://github.com/test/repo/releases/v1.0.0",
				"upload_url": "https://uploads.github.com/repos/test/repo/releases/1/assets{?name,label}",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	})

	server := httptest.NewServer(mux)
	return server, recorder
}

// mockGitLabServer creates a mock GitLab API server.
func mockGitLabServer(t *testing.T) (*httptest.Server, *requestRecorder) {
	t.Helper()
	recorder := &requestRecorder{}

	mux := http.NewServeMux()

	// Create release endpoint
	mux.HandleFunc("/api/v4/projects/", func(w http.ResponseWriter, r *http.Request) {
		recorder.record(r)

		if strings.Contains(r.URL.Path, "/releases") && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"tag_name": "v1.0.0",
				"_links": map[string]interface{}{
					"self": "https://gitlab.com/test/repo/-/releases/v1.0.0",
				},
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	})

	server := httptest.NewServer(mux)
	return server, recorder
}

// requestRecorder records HTTP requests for later inspection.
type requestRecorder struct {
	requests []recordedRequest
}

type recordedRequest struct {
	Method string
	Path   string
	Body   string
}

func (r *requestRecorder) record(req *http.Request) {
	rec := recordedRequest{
		Method: req.Method,
		Path:   req.URL.Path,
	}
	r.requests = append(r.requests, rec)
}

func (r *requestRecorder) hasRequest(method, pathContains string) bool {
	for _, req := range r.requests {
		if req.Method == method && strings.Contains(req.Path, pathContains) {
			return true
		}
	}
	return false
}
