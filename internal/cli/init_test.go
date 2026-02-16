package cli

import (
	"os"
	"testing"

	"release-it-go/internal/config"
	"release-it-go/internal/ui"
)

// mockPrompter records calls and returns preset answers.
type mockPrompter struct {
	confirmAnswers []bool
	confirmIdx     int
	selectAnswers  []int
	selectIdx      int
	inputAnswers   []string
	inputIdx       int
}

func (m *mockPrompter) SelectVersion(current string, recommended string, options []ui.VersionOption) (string, error) {
	return recommended, nil
}

func (m *mockPrompter) Confirm(message string, defaultYes bool) (bool, error) {
	if m.confirmIdx < len(m.confirmAnswers) {
		ans := m.confirmAnswers[m.confirmIdx]
		m.confirmIdx++
		return ans, nil
	}
	return defaultYes, nil
}

func (m *mockPrompter) Input(message string, defaultValue string) (string, error) {
	if m.inputIdx < len(m.inputAnswers) {
		ans := m.inputAnswers[m.inputIdx]
		m.inputIdx++
		return ans, nil
	}
	return defaultValue, nil
}

func (m *mockPrompter) Select(question string, options []string, defaultIndex int) (int, error) {
	if m.selectIdx < len(m.selectAnswers) {
		ans := m.selectAnswers[m.selectIdx]
		m.selectIdx++
		return ans, nil
	}
	return defaultIndex, nil
}

func TestRunInit_WizardCreatesConfig(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{0, 0},                                                            // GitHub, Conventional Changelog
		confirmAnswers: []bool{true},                                                           // git enabled
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}", "main"}, // commit msg, tag format, branch
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	if !config.DetectNativeConfig() {
		t.Fatal("expected .release-it-go.json to be created")
	}

	// Verify the written config can be loaded
	cfg, err := config.LoadConfig(config.NativeConfigFile)
	if err != nil {
		t.Fatalf("loading created config: %v", err)
	}

	if !cfg.GitHub.Release {
		t.Error("expected github.release=true")
	}
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("expected git.tagName=v${version}, got %s", cfg.Git.TagName)
	}
}

func TestRunInit_GitLabPlatform(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{1, 1},  // GitLab, Keep a Changelog
		confirmAnswers: []bool{true}, // git enabled
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}", "main"},
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	cfg, err := config.LoadConfig(config.NativeConfigFile)
	if err != nil {
		t.Fatalf("loading created config: %v", err)
	}

	if !cfg.GitLab.Release {
		t.Error("expected gitlab.release=true")
	}
	if !cfg.Changelog.KeepAChangelog {
		t.Error("expected changelog.keepAChangelog=true")
	}
}

func TestRunInit_GitTagOnly_NoChangelog(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{2, 2},   // Git tag only, No changelog
		confirmAnswers: []bool{false}, // git disabled
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}", "main"},
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	cfg, err := config.LoadConfig(config.NativeConfigFile)
	if err != nil {
		t.Fatalf("loading created config: %v", err)
	}

	if cfg.GitHub.Release {
		t.Error("expected github.release=false")
	}
	if cfg.GitLab.Release {
		t.Error("expected gitlab.release=false")
	}
	if cfg.Changelog.Enabled {
		t.Error("expected changelog.enabled=false")
	}
	if cfg.Git.Commit {
		t.Error("expected git.commit=false")
	}
	if cfg.Git.RequireUpstream {
		t.Error("expected git.requireUpstream=false when push disabled")
	}
}

func TestRunInit_MigrateLegacy(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create legacy config
	legacy := `{"github": {"release": true}, "git": {"tagName": "v${version}"}}`
	if err := os.WriteFile(config.LegacyConfigFile, []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	p := &mockPrompter{
		confirmAnswers: []bool{true}, // yes, migrate
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	if !config.DetectNativeConfig() {
		t.Fatal("expected .release-it-go.json to be created after migration")
	}

	// Check backup exists
	if _, err := os.Stat(config.LegacyConfigFile + ".bak"); err != nil {
		t.Error("expected backup file after migration")
	}
}

func TestRunInit_ExistingNativeConfig_Abort(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create existing native config
	if err := os.WriteFile(config.NativeConfigFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	p := &mockPrompter{
		confirmAnswers: []bool{false}, // don't overwrite
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	// File should remain unchanged
	data, _ := os.ReadFile(config.NativeConfigFile)
	if string(data) != "{}" {
		t.Error("config file should not have been modified")
	}
}

func TestRunInit_ExistingNativeConfig_Overwrite(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create existing native config
	if err := os.WriteFile(config.NativeConfigFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	p := &mockPrompter{
		confirmAnswers: []bool{true, true}, // overwrite, git enabled
		selectAnswers:  []int{0, 0},        // GitHub, Conventional
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}", "main"},
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	// File should be rewritten
	data, _ := os.ReadFile(config.NativeConfigFile)
	if string(data) == "{}" {
		t.Error("config file should have been overwritten with wizard output")
	}
}

func TestInitCommand_Exists(t *testing.T) {
	root := NewRootCommand()
	for _, cmd := range root.Commands() {
		if cmd.Use == "init" {
			return
		}
	}
	t.Error("init command not found in root command")
}
