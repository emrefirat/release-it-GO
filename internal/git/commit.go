package git

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
