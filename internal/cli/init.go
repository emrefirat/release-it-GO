package cli

import (
	"fmt"
	"os"

	"release-it-go/internal/config"
	"release-it-go/internal/ui"

	"github.com/spf13/cobra"
)

// fullExampleFlag controls whether to generate a full example config.
var fullExampleFlag bool

// newInitCommand creates the "init" subcommand for interactive config setup.
func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new release-it-go configuration",
		Long: `Interactively create a .release-it-go.json or .release-it-go.yaml configuration file.
If a legacy .release-it.json file is found, it can be migrated automatically.

Use --full-example to generate a comprehensive YAML example config file
showing all available options with documentation comments.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runInit,
	}

	cmd.Flags().BoolVar(&fullExampleFlag, "full-example", false, "Generate a full example config with all options")

	return cmd
}

// fullExampleFile is the output filename for --full-example (YAML for comment support).
const fullExampleFile = ".release-it-go-full.yaml"

func runInit(cmd *cobra.Command, args []string) error {
	if fullExampleFlag {
		return runInitFullExample()
	}

	var prompter ui.Prompter
	if ciMode {
		prompter = &ui.NonInteractivePrompter{}
	} else {
		prompter = &ui.InteractivePrompter{}
	}

	return runInitWithPrompter(prompter)
}

// runInitFullExample generates a full example config file with all options (YAML).
func runInitFullExample() error {
	if err := config.WriteFullExampleYAML(fullExampleFile); err != nil {
		return fmt.Errorf("writing full example config: %w", err)
	}

	fmt.Printf("Created %s with all available options.\n", fullExampleFile)
	fmt.Println("Copy the options you need to .release-it-go.yaml and customize them.")
	return nil
}

// runInitWithPrompter runs the init wizard with the given prompter (testable).
func runInitWithPrompter(prompter ui.Prompter) error {
	// Check for existing native config (format-independent)
	var existingFile string
	if ef, found := config.DetectNativeConfigAny(); found {
		existingFile = ef
		overwrite, err := prompter.Confirm(
			fmt.Sprintf("%s already exists. Overwrite?", existingFile),
			false,
		)
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Check for legacy config (format-independent — if migration accepted, ask format and return)
	if legacyPath, found := config.DetectLegacyConfig(); found {
		migrate, err := prompter.Confirm(
			fmt.Sprintf("Found %s. Migrate to native format?", legacyPath),
			true,
		)
		if err != nil {
			return err
		}
		if migrate {
			return runMigrationWithFormat(prompter, legacyPath)
		}
	}

	// Run wizard
	cfg := config.DefaultConfig()

	// Track every field the wizard explicitly configures
	force := config.ForceFields{
		"git":       {},
		"github":    {},
		"gitlab":    {},
		"changelog": {},
	}

	// Platform selection
	platformIdx, err := prompter.Select("Select platform:", []string{
		"GitHub",
		"GitLab",
		"Git tag only (no releases)",
	}, 0)
	if err != nil {
		return err
	}

	switch platformIdx {
	case 0: // GitHub
		cfg.GitHub.Release = true
		force["github"]["release"] = true
	case 1: // GitLab
		cfg.GitLab.Release = true
		force["gitlab"]["release"] = true
	case 2: // Git tag only
		// defaults: no release
	}

	// Changelog: default Conventional Changelog (angular preset)
	cfg.Changelog.Enabled = true
	cfg.Changelog.Preset = "angular"
	force["changelog"]["enabled"] = true
	force["changelog"]["preset"] = true

	// Ask whether to write CHANGELOG.md file
	writeFile, err := prompter.Confirm("Write CHANGELOG.md file?", true)
	if err != nil {
		return err
	}
	force["changelog"]["infile"] = true
	if !writeFile {
		cfg.Changelog.Infile = ""
	}

	// Git commit and tag
	commitTag, err := prompter.Confirm("Enable git commit and tag?", true)
	if err != nil {
		return err
	}
	cfg.Git.Commit = commitTag
	cfg.Git.Tag = commitTag
	force["git"]["commit"] = true
	force["git"]["tag"] = true

	// Git push (separate from commit/tag)
	pushEnabled, err := prompter.Confirm("Enable git push?", true)
	if err != nil {
		return err
	}
	cfg.Git.Push = pushEnabled
	force["git"]["push"] = true

	// When push is disabled, upstream check is irrelevant
	if !pushEnabled {
		cfg.Git.RequireUpstream = false
	}

	// Commit message template
	commitMsg, err := prompter.Input(
		"Commit message template",
		"chore(release): release v${version}",
	)
	if err != nil {
		return err
	}
	cfg.Git.CommitMessage = commitMsg
	force["git"]["commitMessage"] = true

	// Tag format
	tagName, err := prompter.Input("Tag name format", "v${version}")
	if err != nil {
		return err
	}
	cfg.Git.TagName = tagName
	force["git"]["tagName"] = true

	// Format selection (last question before writing)
	formatIdx, err := prompter.Select("Config format:", []string{
		"JSON",
		"YAML",
	}, 0)
	if err != nil {
		return err
	}

	format := "json"
	if formatIdx == 1 {
		format = "yaml"
	}
	nativeFile := config.NativeConfigFileForFormat(format)

	// If switching format (e.g. JSON→YAML), rename old file to .bak
	if existingFile != "" && existingFile != nativeFile {
		backupPath := existingFile + ".bak"
		if err := os.Rename(existingFile, backupPath); err != nil {
			return fmt.Errorf("renaming old config %s → %s: %w", existingFile, backupPath, err)
		}
		fmt.Printf("Renamed %s → %s\n", existingFile, backupPath)
	}

	// Write config in selected format, including all wizard-configured fields
	switch format {
	case "yaml":
		if err := config.WriteConfigYAMLWith(cfg, nativeFile, force); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	default:
		if err := config.WriteConfigJSONWith(cfg, nativeFile, force); err != nil {
			return fmt.Errorf("writing config: %w", err)
		}
	}

	fmt.Printf("Created %s\n", nativeFile)
	return nil
}

// runMigrationWithFormat asks for format and performs legacy config migration.
func runMigrationWithFormat(prompter ui.Prompter, legacyPath string) error {
	formatIdx, err := prompter.Select("Config format:", []string{
		"JSON",
		"YAML",
	}, 0)
	if err != nil {
		return err
	}

	format := "json"
	if formatIdx == 1 {
		format = "yaml"
	}
	nativeFile := config.NativeConfigFileForFormat(format)

	if err := config.MigrateLegacyConfigTo(legacyPath, format); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	fmt.Printf("Migrated %s → %s (backup: %s.bak)\n", legacyPath, nativeFile, legacyPath)
	return nil
}
