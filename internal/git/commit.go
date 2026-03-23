package git

import "strings"

// HasStagedChanges returns true if there are staged changes ready to commit.
// In dry-run mode, always returns true since write operations are simulated.
// On git error, assumes true to avoid skipping commits incorrectly.
func (g *Git) HasStagedChanges() bool {
	if g.dryRun {
		return true
	}
	out, err := g.runSilent("diff", "--cached", "--name-only")
	if err != nil {
		// If we can't check, assume there are changes — let git commit decide
		return true
	}
	return strings.TrimSpace(out) != ""
}

// Stage adds changed files to the staging area.
// If AddUntrackedFiles is true, all files (including untracked) are staged.
// Otherwise, only tracked files are updated.
func (g *Git) Stage() error {
	if g.config.AddUntrackedFiles {
		_, err := g.run("add", ".")
		return err
	}
	_, err := g.run("add", ".", "--update")
	return err
}

// StageFile explicitly adds a specific file to the staging area.
// Used for files generated during the release process (e.g. CHANGELOG.md).
func (g *Git) StageFile(path string) error {
	_, err := g.run("add", path)
	return err
}

// Commit creates a commit with the given message.
// Additional commit arguments from config are appended.
func (g *Git) Commit(message string) error {
	args := []string{"commit", "--message", message}
	args = append(args, g.config.CommitArgs...)
	_, err := g.run(args...)
	return err
}
