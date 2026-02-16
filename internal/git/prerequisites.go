package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CheckPrerequisites runs all prerequisite checks in order.
// Returns an error on the first failed check.
func (g *Git) CheckPrerequisites() error {
	if !IsGitInstalled() {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	if !IsGitRepo() {
		return fmt.Errorf("current directory is not a git repository")
	}

	if err := g.checkBranch(); err != nil {
		return err
	}
	if err := g.checkCleanWorkingDir(); err != nil {
		return err
	}
	if err := g.checkUpstream(); err != nil {
		return err
	}
	if err := g.checkGitIdentity(); err != nil {
		return err
	}
	if err := g.checkCommits(); err != nil {
		return err
	}
	return nil
}

// checkBranch verifies the current branch matches the required pattern.
func (g *Git) checkBranch() error {
	if g.config.RequireBranch == "" {
		return nil
	}

	branch, err := g.runSilent("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	branch = strings.TrimSpace(branch)

	matched, err := filepath.Match(g.config.RequireBranch, branch)
	if err != nil {
		// Pattern error, fall back to exact match
		matched = branch == g.config.RequireBranch
	}

	if !matched {
		return fmt.Errorf("required branch is %s, but current branch is %s", g.config.RequireBranch, branch)
	}

	return nil
}

// checkCleanWorkingDir verifies there are no uncommitted changes.
// Untracked files are ignored — only staged and unstaged modifications count.
func (g *Git) checkCleanWorkingDir() error {
	if !g.config.RequireCleanWorkingDir {
		return nil
	}

	// -uno excludes untracked files from the output.
	// Untracked files are not "uncommitted changes" and should not block a release.
	out, err := g.runSilent("status", "--porcelain", "-uno")
	if err != nil {
		return fmt.Errorf("failed to check working directory status: %w", err)
	}

	if strings.TrimSpace(out) != "" {
		return fmt.Errorf("working directory is not clean (uncommitted changes exist)")
	}

	return nil
}

// checkUpstream verifies the current branch has an upstream tracking branch.
func (g *Git) checkUpstream() error {
	if !g.config.RequireUpstream {
		return nil
	}

	_, err := g.runSilent("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		return fmt.Errorf("no upstream configured for current branch")
	}

	return nil
}

// checkGitIdentity verifies that git user.name and user.email are configured.
// This is only checked when commit is enabled, to prevent failures in Docker
// containers or fresh environments where git identity is not set.
func (g *Git) checkGitIdentity() error {
	if !g.config.Commit {
		return nil
	}

	name, _ := commandExecutor("git", "config", "user.name")
	email, _ := commandExecutor("git", "config", "user.email")

	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	if name == "" || email == "" {
		return fmt.Errorf("git user identity is not configured (user.name or user.email is missing);\n" +
			"  run: git config --global user.name \"Your Name\"\n" +
			"       git config --global user.email \"you@example.com\"")
	}

	return nil
}

// checkCommits verifies there are new commits since the latest tag.
func (g *Git) checkCommits() error {
	if !g.config.RequireCommits {
		return nil
	}

	latestTag, err := g.GetLatestTag()
	if err != nil {
		// No tags exist yet, so there must be commits
		return nil
	}

	out, err := g.runSilent("log", latestTag+"..HEAD", "--oneline")
	if err != nil {
		return fmt.Errorf("failed to check commits since %s: %w", latestTag, err)
	}

	if strings.TrimSpace(out) == "" {
		return fmt.Errorf("no commits since latest tag %s", latestTag)
	}

	return nil
}
