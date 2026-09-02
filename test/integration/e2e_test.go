package integration

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/emrefirat/release-it-GO/internal/config"
	"github.com/emrefirat/release-it-GO/internal/runner"
)

var (
	binaryOnce sync.Once
	binaryPath string
	binaryDir  string
	binaryErr  error
)

// releaseItBinary builds cmd/release-it-go once per test run and returns its
// path. The build runs from the module root so it works after t.Chdir; the
// directory is removed by TestMain.
func releaseItBinary(t *testing.T) string {
	t.Helper()
	binaryOnce.Do(func() {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			binaryErr = errors.New("cannot locate the module root")
			return
		}
		root := filepath.Join(filepath.Dir(file), "..", "..")
		binaryDir, binaryErr = os.MkdirTemp("", "release-it-go-bin-*")
		if binaryErr != nil {
			return
		}
		name := "release-it-go"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		binaryPath = filepath.Join(binaryDir, name)
		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/release-it-go")
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			binaryErr = fmt.Errorf("go build: %w\n%s", err, out)
		}
	})
	if binaryErr != nil {
		t.Fatalf("building release-it-go: %v", binaryErr)
	}
	return binaryPath
}

func removeBuiltBinary() {
	if binaryDir != "" {
		_ = os.RemoveAll(binaryDir)
	}
}

// The commit-msg git hook, end to end: the real binary installs the hook,
// git invokes it with the message file, and the hook calls the binary back.
func TestIntegration_CommitMsgHook_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	bin := releaseItBinary(t)
	dir := t.TempDir()
	t.Chdir(dir)
	initGitRepo(t, dir)

	writeFile(t, filepath.Join(dir, ".release-it-go.yaml"), fmt.Sprintf(
		"git:\n  push: false\n  requireUpstream: false\nhooks:\n  \"commit-msg\": ['%s --check-msg \"$1\"']\n", bin))

	install := exec.Command(bin, "hooks", "install")
	install.Dir = dir
	if out, err := install.CombinedOutput(); err != nil {
		t.Fatalf("hooks install: %v\n%s", err, out)
	}
	assertFileContains(t, filepath.Join(dir, ".hooks", "commit-msg"), "--check-msg")
	if got := strings.TrimSpace(runGit(t, dir, "config", "core.hooksPath")); got != ".hooks" {
		t.Fatalf("core.hooksPath = %q, want .hooks", got)
	}

	before := getHeadHash(t, dir)

	bad := exec.Command("git", "commit", "--allow-empty", "-m", "bad message")
	bad.Dir = dir
	badOut, badErr := bad.CombinedOutput()
	if badErr == nil {
		t.Fatalf("a non-conventional message must be rejected by the hook; output:\n%s", badOut)
	}
	if !strings.Contains(string(badOut), "Invalid commit message") {
		t.Errorf("hook output should carry the --check-msg diagnostics, got:\n%s", badOut)
	}
	if getHeadHash(t, dir) != before {
		t.Error("a rejected commit must not move HEAD")
	}

	good := exec.Command("git", "commit", "--allow-empty", "-m", "feat: hook accepts conventional messages")
	good.Dir = dir
	if out, err := good.CombinedOutput(); err != nil {
		t.Fatalf("a conventional message must pass the hook: %v\n%s", err, out)
	}
	if getHeadHash(t, dir) == before {
		t.Error("an accepted commit must create a new HEAD")
	}
}

// Lifecycle hooks receive the release context as RELEASE_* environment
// variables — checked from a real shell, not from the env-building helper.
func TestIntegration_LifecycleHook_ReceivesReleaseEnv(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := t.TempDir()
	t.Chdir(dir)
	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"feat: add feature"})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"
	cfg.Hooks.AfterBump = []string{`printf '%s|%s|%s' "$RELEASE_VERSION" "$RELEASE_TAG_NAME" "$RELEASE_LATEST_VERSION" > release-env.txt`}

	if err := runner.NewRunner(cfg).Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "release-env.txt"))
	if err != nil {
		t.Fatalf("hook did not write its environment: %v", err)
	}
	if got, want := string(data), "1.1.0|v1.1.0|1.0.0"; got != want {
		t.Errorf("RELEASE_* seen by the hook = %q, want %q", got, want)
	}
}

// The bumper must leave everything but the version value untouched when the
// whole pipeline runs — byte-for-byte, including a same-looking dependency
// version, indentation, key order and comments.
func TestIntegration_Bumper_PreservesFormatting_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	dir := t.TempDir()
	t.Chdir(dir)

	pkgJSON := filepath.Join(dir, "package.json")
	writeFile(t, pkgJSON, "{\n    \"name\": \"test-app\",\n    \"version\": \"1.0.0\",\n    \"dependencies\": {\n        \"left-pad\": \"1.0.0\"\n    }\n}\n")
	chart := filepath.Join(dir, "Chart.yaml")
	writeFile(t, chart, "# Helm chart\napiVersion: v2\nname: app   # chart name\nversion: 1.0.0   # bumped by release-it-go\nappVersion: \"1.0.0\"\n")

	initGitRepo(t, dir)
	createTag(t, dir, "v1.0.0")
	createCommits(t, dir, []string{"feat: add feature"})

	cfg := newTestConfig(dir)
	cfg.Increment = "minor"
	cfg.Bumper.Enabled = true
	cfg.Bumper.Out = []config.BumperFile{
		{File: pkgJSON, Path: "version"},
		{File: chart, Path: "version"},
	}

	if err := runner.NewRunner(cfg).Run(); err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	wantJSON := "{\n    \"name\": \"test-app\",\n    \"version\": \"1.1.0\",\n    \"dependencies\": {\n        \"left-pad\": \"1.0.0\"\n    }\n}\n"
	if got, _ := os.ReadFile(pkgJSON); string(got) != wantJSON {
		t.Errorf("package.json after release:\n%s\nwant:\n%s", got, wantJSON)
	}
	wantYAML := "# Helm chart\napiVersion: v2\nname: app   # chart name\nversion: 1.1.0   # bumped by release-it-go\nappVersion: \"1.0.0\"\n"
	if got, _ := os.ReadFile(chart); string(got) != wantYAML {
		t.Errorf("Chart.yaml after release:\n%s\nwant:\n%s", got, wantYAML)
	}
	assertWorkingDirClean(t, dir)
}
