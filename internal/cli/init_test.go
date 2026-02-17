package cli

import (
	"encoding/json"
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
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{0, 0},                                                    // GitHub, JSON
		confirmAnswers: []bool{true, true, true},                                       // writeChangelog, commit/tag, push
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"}, // commit msg, tag format
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

func TestRunInit_WizardWritesExplicitFields(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	// User picks GitHub + all defaults
	// Even though commit=true, tag=true, push=true are defaults,
	// they should appear in the config because the wizard explicitly asked about them.
	p := &mockPrompter{
		selectAnswers:  []int{0, 0},                                                    // GitHub, JSON
		confirmAnswers: []bool{true, true, true},                                       // writeChangelog=YES, commit/tag=YES, push=YES
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"}, // commit msg, tag format
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	data, err := os.ReadFile(config.NativeConfigFile)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing JSON: %v", err)
	}

	// git section must contain wizard-configured fields even if they match defaults
	var gitMap map[string]interface{}
	if err := json.Unmarshal(raw["git"], &gitMap); err != nil {
		t.Fatalf("parsing git section: %v", err)
	}
	for _, key := range []string{"commit", "tag", "push", "commitMessage", "tagName"} {
		if _, ok := gitMap[key]; !ok {
			t.Errorf("expected git.%s to be explicitly written in config", key)
		}
	}

	// changelog section must contain wizard-configured fields
	var clMap map[string]interface{}
	if err := json.Unmarshal(raw["changelog"], &clMap); err != nil {
		t.Fatalf("parsing changelog section: %v", err)
	}
	for _, key := range []string{"enabled", "preset", "infile"} {
		if _, ok := clMap[key]; !ok {
			t.Errorf("expected changelog.%s to be explicitly written in config", key)
		}
	}

	// github section must contain release field
	var ghMap map[string]interface{}
	if err := json.Unmarshal(raw["github"], &ghMap); err != nil {
		t.Fatalf("parsing github section: %v", err)
	}
	if _, ok := ghMap["release"]; !ok {
		t.Error("expected github.release to be explicitly written in config")
	}
}

func TestRunInit_GitLabPlatform(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{1, 0},              // GitLab, JSON
		confirmAnswers: []bool{true, true, true}, // writeChangelog, commit/tag, push
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"},
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
	if !cfg.Changelog.Enabled {
		t.Error("expected changelog.enabled=true")
	}
	if cfg.Changelog.Preset != "angular" {
		t.Errorf("expected changelog.preset=angular, got %s", cfg.Changelog.Preset)
	}
}

func TestRunInit_GitTagOnly_NoChangelog(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{2, 0},                // Git tag only, JSON
		confirmAnswers: []bool{true, false, false}, // writeChangelog, commit/tag disabled, push disabled
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"},
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
	if !cfg.Changelog.Enabled {
		t.Error("expected changelog.enabled=true (default conventional)")
	}
	if cfg.Git.Commit {
		t.Error("expected git.commit=false")
	}
	if cfg.Git.Push {
		t.Error("expected git.push=false")
	}
	if cfg.Git.RequireUpstream {
		t.Error("expected git.requireUpstream=false when push disabled")
	}
}

func TestRunInit_CommitTagEnabled_PushDisabled(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{2, 0},               // Git tag only, JSON
		confirmAnswers: []bool{true, true, false}, // writeChangelog, commit/tag YES, push NO
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"},
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	cfg, err := config.LoadConfig(config.NativeConfigFile)
	if err != nil {
		t.Fatalf("loading created config: %v", err)
	}

	if !cfg.Git.Commit {
		t.Error("expected git.commit=true")
	}
	if !cfg.Git.Tag {
		t.Error("expected git.tag=true")
	}
	if cfg.Git.Push {
		t.Error("expected git.push=false")
	}
	if cfg.Git.RequireUpstream {
		t.Error("expected git.requireUpstream=false when push disabled")
	}
	if !cfg.Git.RequireCleanWorkingDir {
		t.Error("expected git.requireCleanWorkingDir=true regardless of push")
	}
}

func TestRunInit_MigrateLegacy(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	// Create legacy config
	legacy := `{"github": {"release": true}, "git": {"tagName": "v${version}"}}`
	if err := os.WriteFile(config.LegacyConfigFile, []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	p := &mockPrompter{
		confirmAnswers: []bool{true}, // yes, migrate
		selectAnswers:  []int{0},     // JSON format (asked after migration decision)
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
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	// Create existing native config
	if err := os.WriteFile(config.NativeConfigFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	p := &mockPrompter{
		confirmAnswers: []bool{false}, // don't overwrite (no select questions asked)
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
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	// Create existing native config
	if err := os.WriteFile(config.NativeConfigFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	p := &mockPrompter{
		confirmAnswers: []bool{true, true, true, true}, // overwrite, writeChangelog, commit/tag, push
		selectAnswers:  []int{0, 0},                  // GitHub, JSON
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"},
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

func TestRunInit_FormatSwitch_RenamesOldConfig(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	// Create existing JSON config
	if err := os.WriteFile(config.NativeConfigFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	p := &mockPrompter{
		selectAnswers:  []int{0, 1},                     // GitHub, YAML
		confirmAnswers: []bool{true, true, true, true}, // overwrite, writeChangelog, commit/tag, push
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"},
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	// Old JSON should be renamed to .bak
	if _, err := os.Stat(config.NativeConfigFile); err == nil {
		t.Error("expected old .release-it-go.json to be renamed, but it still exists")
	}
	if _, err := os.Stat(config.NativeConfigFile + ".bak"); err != nil {
		t.Error("expected .release-it-go.json.bak to exist after format switch")
	}

	// New YAML should be created
	if _, err := os.Stat(config.NativeConfigFileYAML); err != nil {
		t.Fatalf("expected .release-it-go.yaml to be created: %v", err)
	}

	// Verify YAML is loadable
	cfg, err := config.LoadConfig(config.NativeConfigFileYAML)
	if err != nil {
		t.Fatalf("loading YAML config: %v", err)
	}
	if !cfg.GitHub.Release {
		t.Error("expected github.release=true")
	}
}

func TestRunInit_ChangelogDefaultConventional(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{0, 0},                                                    // GitHub, JSON
		confirmAnswers: []bool{true, true, true},                                       // writeChangelog, commit/tag, push
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"}, // commit msg, tag format
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	cfg, err := config.LoadConfig(config.NativeConfigFile)
	if err != nil {
		t.Fatalf("loading created config: %v", err)
	}

	if !cfg.Changelog.Enabled {
		t.Error("expected changelog.enabled=true")
	}
	if cfg.Changelog.Preset != "angular" {
		t.Errorf("expected changelog.preset=angular, got %s", cfg.Changelog.Preset)
	}
}

func TestRunInit_FullExample(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	if err := runInitFullExample(); err != nil {
		t.Fatalf("runInitFullExample failed: %v", err)
	}

	// Verify YAML file was created
	data, err := os.ReadFile(fullExampleFile)
	if err != nil {
		t.Fatalf("expected %s to be created: %v", fullExampleFile, err)
	}

	content := string(data)

	// Verify key sections are present (YAML format)
	checks := []string{
		"git:",
		"github:",
		"gitlab:",
		"hooks:",
		"changelog:",
		"bumper:",
		"calver:",
		"notification:",
		"commitMessage:",
		"tagName:",
		"tokenRef:",
		"SLACK_WEBHOOK_URL",
		"TEAMS_WEBHOOK_URL",
	}
	for _, check := range checks {
		if !contains(content, check) {
			t.Errorf("expected full example to contain %s", check)
		}
	}

	// Should contain YAML comments
	if !contains(content, "# ") {
		t.Error("expected YAML comments in full example")
	}

	// Should NOT have runtime flags
	for _, flag := range []string{"ci:", "dry-run:", "verbose:", "preReleaseId:"} {
		if contains(content, flag) {
			t.Errorf("full example should not contain runtime flag %s", flag)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRunInit_YAMLFormat(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(dir)

	p := &mockPrompter{
		selectAnswers:  []int{0, 1},                                                    // GitHub, YAML
		confirmAnswers: []bool{true, true, true},                                       // writeChangelog, commit/tag, push
		inputAnswers:   []string{"chore(release): release v${version}", "v${version}"}, // commit msg, tag format
	}

	if err := runInitWithPrompter(p); err != nil {
		t.Fatalf("runInitWithPrompter failed: %v", err)
	}

	// Should create .release-it-go.yaml (not .json)
	if _, err := os.Stat(config.NativeConfigFileYAML); err != nil {
		t.Fatalf("expected %s to be created: %v", config.NativeConfigFileYAML, err)
	}

	// Should NOT create .json
	if _, err := os.Stat(config.NativeConfigFile); err == nil {
		t.Error("expected .release-it-go.json NOT to be created when YAML is selected")
	}

	// Verify the written YAML config can be loaded
	cfg, err := config.LoadConfig(config.NativeConfigFileYAML)
	if err != nil {
		t.Fatalf("loading created YAML config: %v", err)
	}

	if !cfg.GitHub.Release {
		t.Error("expected github.release=true")
	}
	if cfg.Git.TagName != "v${version}" {
		t.Errorf("expected git.tagName=v${version}, got %s", cfg.Git.TagName)
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
