package cli

import (
	"fmt"

	"release-it-go/internal/config"
	"release-it-go/internal/githook"

	"github.com/spf13/cobra"
)

var (
	installForce  bool
	installRemove bool
)

func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install git hooks from configuration",
		Long: `Install git hooks defined in the hooks section of your
release-it-go config file into .git/hooks/.

Supported git hooks: pre-commit, commit-msg, pre-push,
post-commit, post-merge, prepare-commit-msg.

Use --remove to uninstall all managed hooks.
Use --force to overwrite existing non-managed hooks.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runInstall,
	}

	cmd.Flags().BoolVar(&installForce, "force", false, "overwrite existing non-managed hooks")
	cmd.Flags().BoolVar(&installRemove, "remove", false, "remove all managed git hooks")

	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	gitDir, err := githook.FindGitDir()
	if err != nil {
		return err
	}

	installer := githook.NewInstaller(gitDir, installForce)

	if installRemove {
		fmt.Println("Removing managed git hooks...")
		return installer.Remove()
	}

	hooks := githook.HooksFromConfig(&cfg.Hooks)
	if len(hooks) == 0 {
		fmt.Println("No git hooks configured in hooks section.")
		fmt.Println("Add pre-commit, commit-msg, or pre-push to your config file.")
		return nil
	}

	fmt.Println("Installing git hooks...")
	return installer.Install(hooks)
}
