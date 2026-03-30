package cli

import (
	"fmt"

	"release-it-go/internal/config"
	"release-it-go/internal/githook"

	"github.com/spf13/cobra"
)

var hooksForce bool

func newHooksCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hooks",
		Short: "Manage git hooks",
		Long:  "Install or remove git hooks defined in your release-it-go configuration.",
	}

	cmd.AddCommand(newHooksInstallCommand())
	cmd.AddCommand(newHooksRemoveCommand())

	return cmd
}

func newHooksInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install git hooks from configuration",
		Long: `Install git hooks defined in the hooks section of your
release-it-go config file into .git/hooks/.

Supported git hooks: pre-commit, commit-msg, pre-push,
post-commit, post-merge, prepare-commit-msg.

Use --force to overwrite existing non-managed hooks.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runHooksInstall,
	}

	cmd.Flags().BoolVar(&hooksForce, "force", false, "overwrite existing non-managed hooks")

	return cmd
}

func newHooksRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:           "remove",
		Short:         "Remove managed git hooks",
		Long:          "Remove all git hooks that were installed by release-it-go.\nUser-created hooks are left untouched.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runHooksRemove,
	}
}

func runHooksInstall(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	gitDir, err := githook.FindGitDir()
	if err != nil {
		return err
	}

	hooks := githook.HooksFromConfig(&cfg.Hooks)
	if len(hooks) == 0 {
		fmt.Println("No git hooks configured in hooks section.")
		fmt.Println("Add pre-commit, commit-msg, or pre-push to your config file.")
		return nil
	}

	installer := githook.NewInstaller(gitDir, hooksForce)
	fmt.Println("Installing git hooks...")
	return installer.Install(hooks)
}

func runHooksRemove(cmd *cobra.Command, args []string) error {
	gitDir, err := githook.FindGitDir()
	if err != nil {
		return err
	}

	installer := githook.NewInstaller(gitDir, false)
	fmt.Println("Removing managed git hooks...")
	return installer.Remove()
}
