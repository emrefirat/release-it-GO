package cli

import (
	"testing"
)

func TestHooksCommand_Registered(t *testing.T) {
	rootCmd := NewRootCommand()

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "hooks" {
			found = true
			// Verify subcommands
			installFound := false
			removeFound := false
			for _, sub := range cmd.Commands() {
				if sub.Use == "install" {
					installFound = true
				}
				if sub.Use == "remove" {
					removeFound = true
				}
			}
			if !installFound {
				t.Error("expected 'install' subcommand under 'hooks'")
			}
			if !removeFound {
				t.Error("expected 'remove' subcommand under 'hooks'")
			}
		}
	}
	if !found {
		t.Error("expected 'hooks' command to be registered")
	}
}

func TestHooksInstallCommand_HasForceFlag(t *testing.T) {
	cmd := newHooksInstallCommand()
	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Error("expected --force flag on hooks install command")
	}
}

func TestHooksRemoveCommand_NoForceFlag(t *testing.T) {
	cmd := newHooksRemoveCommand()
	flag := cmd.Flags().Lookup("force")
	if flag != nil {
		t.Error("remove command should not have --force flag")
	}
}
