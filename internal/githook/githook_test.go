package githook

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"release-it-go/internal/config"
)

func TestGenerateScript(t *testing.T) {
	script := generateScript("pre-commit", []string{"go fmt ./...", "go vet ./..."})

	if !strings.HasPrefix(script, "#!/bin/sh\n") {
		t.Error("expected shebang line")
	}
	if !strings.Contains(script, managedHeader) {
		t.Error("expected managed header")
	}
	if !strings.Contains(script, "# Hook: pre-commit") {
		t.Error("expected hook name comment")
	}
	if !strings.Contains(script, "set -e") {
		t.Error("expected set -e")
	}
	if !strings.Contains(script, "go fmt ./...") {
		t.Error("expected first command")
	}
	if !strings.Contains(script, "go vet ./...") {
		t.Error("expected second command")
	}
}

func TestGenerateScript_SingleCommand(t *testing.T) {
	script := generateScript("commit-msg", []string{"./release-it-go --check-commits --file ${1}"})

	if !strings.Contains(script, "${1}") {
		t.Error("expected ${1} to be preserved literally in script")
	}
}

func TestIsManagedHook(t *testing.T) {
	dir := t.TempDir()
	installer := NewInstaller(dir, false)

	// Managed hook
	managedPath := filepath.Join(dir, "managed")
	_ = os.WriteFile(managedPath, []byte("#!/bin/sh\n"+managedHeader+"\necho hello"), 0755)

	managed, err := installer.isManagedHook(managedPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !managed {
		t.Error("expected managed=true for hook with managed header")
	}

	// User hook
	userPath := filepath.Join(dir, "user")
	_ = os.WriteFile(userPath, []byte("#!/bin/sh\necho my custom hook"), 0755)

	managed, err = installer.isManagedHook(userPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if managed {
		t.Error("expected managed=false for user hook")
	}
}

func TestInstall_NewHooks(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0755)

	installer := NewInstaller(gitDir, false)
	hooks := map[string][]string{
		"pre-commit": {"go fmt ./..."},
		"pre-push":   {"go test ./..."},
	}

	err := installer.Install(hooks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify files created
	for _, name := range []string{"pre-commit", "pre-push"} {
		path := filepath.Join(gitDir, "hooks", name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("hook %s not created: %v", name, err)
		}
		if info.Mode().Perm() != 0755 {
			t.Errorf("hook %s permission = %o, want 0755", name, info.Mode().Perm())
		}
		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), managedHeader) {
			t.Errorf("hook %s missing managed header", name)
		}
	}
}

func TestInstall_CreatesHooksDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0755)
	// Don't create hooks/ dir — Install should create it

	installer := NewInstaller(gitDir, false)
	err := installer.Install(map[string][]string{"pre-commit": {"echo test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists(filepath.Join(gitDir, "hooks", "pre-commit")) {
		t.Error("expected hooks dir and pre-commit to be created")
	}
}

func TestInstall_RejectsUserHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	_ = os.MkdirAll(hooksDir, 0755)

	// Create existing user hook
	_ = os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte("#!/bin/sh\necho user hook"), 0755)

	installer := NewInstaller(gitDir, false)
	err := installer.Install(map[string][]string{"pre-commit": {"go fmt ./..."}})
	if err == nil {
		t.Fatal("expected error for existing user hook")
	}
	if !strings.Contains(err.Error(), "not managed") {
		t.Errorf("expected 'not managed' in error, got: %v", err)
	}
}

func TestInstall_ForceOverwritesUserHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	_ = os.MkdirAll(hooksDir, 0755)

	// Create existing user hook
	_ = os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte("#!/bin/sh\necho user hook"), 0755)

	installer := NewInstaller(gitDir, true) // force=true
	err := installer.Install(map[string][]string{"pre-commit": {"go fmt ./..."}})
	if err != nil {
		t.Fatalf("unexpected error with --force: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	if !strings.Contains(string(content), managedHeader) {
		t.Error("expected managed header after force overwrite")
	}
}

func TestInstall_OverwritesManagedHook(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	_ = os.MkdirAll(hooksDir, 0755)

	// Create existing managed hook
	_ = os.WriteFile(filepath.Join(hooksDir, "pre-commit"),
		[]byte("#!/bin/sh\n"+managedHeader+"\necho old"), 0755)

	installer := NewInstaller(gitDir, false) // No force needed for managed hooks
	err := installer.Install(map[string][]string{"pre-commit": {"echo new"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(hooksDir, "pre-commit"))
	if !strings.Contains(string(content), "echo new") {
		t.Error("expected updated content")
	}
}

func TestInstall_EmptyHooks(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0755)

	installer := NewInstaller(gitDir, false)
	err := installer.Install(map[string][]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemove_ManagedOnly(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	_ = os.MkdirAll(hooksDir, 0755)

	// Create one managed and one user hook
	_ = os.WriteFile(filepath.Join(hooksDir, "pre-commit"),
		[]byte("#!/bin/sh\n"+managedHeader+"\necho managed"), 0755)
	_ = os.WriteFile(filepath.Join(hooksDir, "pre-push"),
		[]byte("#!/bin/sh\necho user hook"), 0755)

	installer := NewInstaller(gitDir, false)
	err := installer.Remove()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Managed hook should be removed
	if exists(filepath.Join(hooksDir, "pre-commit")) {
		t.Error("expected managed pre-commit to be removed")
	}
	// User hook should remain
	if !exists(filepath.Join(hooksDir, "pre-push")) {
		t.Error("expected user pre-push to remain")
	}
}

func TestRemove_NoHooksDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0755)
	// No hooks/ dir

	installer := NewInstaller(gitDir, false)
	err := installer.Remove()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindGitDir(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "/repo/.git", nil
	}

	gitDir, err := FindGitDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gitDir != "/repo/.git" {
		t.Errorf("expected /repo/.git, got %q", gitDir)
	}
}

func TestFindGitDir_NotRepo(t *testing.T) {
	original := commandExecutor
	defer func() { commandExecutor = original }()

	commandExecutor = func(name string, args ...string) (string, error) {
		return "", os.ErrNotExist
	}

	_, err := FindGitDir()
	if err == nil {
		t.Error("expected error for non-git directory")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("expected 'not a git repository' in error, got: %v", err)
	}
}

func TestHooksFromConfig(t *testing.T) {
	cfg := &config.HooksConfig{
		PreCommit:  []string{"go fmt ./..."},
		CommitMsg:  []string{"./release-it-go --check-commits"},
		PrePush:    []string{"go test ./..."},
		BeforeInit: []string{"echo lifecycle"}, // Should be ignored
	}

	hooks := HooksFromConfig(cfg)

	if len(hooks) != 3 {
		t.Errorf("expected 3 git hooks, got %d", len(hooks))
	}
	if _, ok := hooks["pre-commit"]; !ok {
		t.Error("expected pre-commit in hooks")
	}
	if _, ok := hooks["commit-msg"]; !ok {
		t.Error("expected commit-msg in hooks")
	}
	if _, ok := hooks["pre-push"]; !ok {
		t.Error("expected pre-push in hooks")
	}
	if _, ok := hooks["before:init"]; ok {
		t.Error("lifecycle hooks should NOT be in git hooks map")
	}
}

func TestHooksFromConfig_Empty(t *testing.T) {
	cfg := &config.HooksConfig{}
	hooks := HooksFromConfig(cfg)
	if len(hooks) != 0 {
		t.Errorf("expected 0 hooks for empty config, got %d", len(hooks))
	}
}

func TestHooksFromConfig_AllGitHooks(t *testing.T) {
	cfg := &config.HooksConfig{
		PreCommit:        []string{"fmt"},
		CommitMsg:        []string{"check"},
		PrePush:          []string{"test"},
		PostCommit:       []string{"notify"},
		PostMerge:        []string{"install"},
		PrepareCommitMsg: []string{"template"},
	}

	hooks := HooksFromConfig(cfg)
	if len(hooks) != 6 {
		t.Errorf("expected 6 git hooks, got %d", len(hooks))
	}
	for _, name := range []string{"pre-commit", "commit-msg", "pre-push", "post-commit", "post-merge", "prepare-commit-msg"} {
		if _, ok := hooks[name]; !ok {
			t.Errorf("expected %s in hooks", name)
		}
	}
}

func TestInstall_SkipsEmptyCommands(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0755)

	installer := NewInstaller(gitDir, false)
	hooks := map[string][]string{
		"pre-commit": {"go fmt ./..."},
		"pre-push":   {}, // Empty — should be skipped
	}

	err := installer.Install(hooks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !exists(filepath.Join(gitDir, "hooks", "pre-commit")) {
		t.Error("expected pre-commit to be created")
	}
	if exists(filepath.Join(gitDir, "hooks", "pre-push")) {
		t.Error("expected pre-push to be skipped (empty commands)")
	}
}

func TestInstall_ScriptContainsShellArgs(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0755)

	installer := NewInstaller(gitDir, false)
	hooks := map[string][]string{
		"commit-msg": {"./scripts/validate.sh ${1}"},
	}

	err := installer.Install(hooks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(filepath.Join(gitDir, "hooks", "commit-msg"))
	if !strings.Contains(string(content), "${1}") {
		t.Error("expected ${1} preserved in generated script")
	}
}

func TestRemove_MultipleManaged(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")
	_ = os.MkdirAll(hooksDir, 0755)

	// Create multiple managed hooks
	for _, name := range []string{"pre-commit", "commit-msg", "pre-push"} {
		_ = os.WriteFile(filepath.Join(hooksDir, name),
			[]byte("#!/bin/sh\n"+managedHeader+"\necho "+name), 0755)
	}

	installer := NewInstaller(gitDir, false)
	err := installer.Remove()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"pre-commit", "commit-msg", "pre-push"} {
		if exists(filepath.Join(hooksDir, name)) {
			t.Errorf("expected %s to be removed", name)
		}
	}
}
