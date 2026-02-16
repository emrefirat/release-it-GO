package cli

import (
	"bytes"
	"testing"
)

func TestNewRootCommand_Help(t *testing.T) {
	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("help output should not be empty")
	}
}

func TestNewRootCommand_VersionSubcommand(t *testing.T) {
	Version = "1.0.0-test"
	Commit = "abc1234"
	Date = "2026-02-16"

	cmd := NewRootCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"version"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestNewRootCommand_DryRunFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--dry-run", "--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_CIFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_VerboseFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"-V", "--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_IncrementFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--increment", "major", "--ci"})

	// May fail without git repo, but should not crash
	_ = cmd.Execute()
}

func TestNewRootCommand_ChangelogFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--changelog", "--ci"})

	// This will fail in test environment (no git repo), which is expected
	_ = cmd.Execute()
}

func TestNewRootCommand_ReleaseVersionFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--release-version", "--ci"})

	// This will fail in test environment (no git repo), which is expected
	_ = cmd.Execute()
}

func TestNewRootCommand_HasExpectedFlags(t *testing.T) {
	cmd := NewRootCommand()

	expectedPersistentFlags := []string{"config", "dry-run", "ci", "verbose", "increment", "preReleaseId"}
	for _, name := range expectedPersistentFlags {
		if cmd.PersistentFlags().Lookup(name) == nil {
			t.Errorf("missing persistent flag: %s", name)
		}
	}

	expectedFlags := []string{"changelog", "release-version", "only-version", "no-increment", "no-git.commit", "no-git.tag", "no-git.push"}
	for _, name := range expectedFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("missing flag: %s", name)
		}
	}
}

func TestNewRootCommand_HasVersionSubcommand(t *testing.T) {
	cmd := NewRootCommand()

	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("root command should have 'version' subcommand")
	}
}
