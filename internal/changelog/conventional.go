package changelog

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"release-it-go/internal/git"
)

// RenderConventional generates changelog content in conventional-changelog format.
// Groups commits by type into sections (Features, Bug Fixes, etc.)
// and includes commit hashes with optional links to the repository.
func RenderConventional(commits []*Commit, version string, prevVersion string, repoInfo *git.RepoInfo) string {
	var sb strings.Builder

	date := time.Now().Format("2006-01-02")

	// Version header with optional compare URL
	if repoInfo != nil && prevVersion != "" {
		compareURL := fmt.Sprintf("https://%s/%s/%s/compare/%s...%s",
			repoInfo.Host, repoInfo.Owner, repoInfo.Repository,
			prevVersion, version)
		fmt.Fprintf(&sb, "## [%s](%s) (%s)\n", version, compareURL, date)
	} else {
		fmt.Fprintf(&sb, "## %s (%s)\n", version, date)
	}

	// Group commits by section
	sections := groupBySection(commits, commitTypeSection)

	// Render sections in consistent order
	sectionOrder := []string{"Features", "Bug Fixes", "Performance Improvements", "Reverts"}
	for _, sectionName := range sectionOrder {
		sectionCommits, ok := sections[sectionName]
		if !ok || len(sectionCommits) == 0 {
			continue
		}

		fmt.Fprintf(&sb, "\n### %s\n\n", sectionName)
		for _, c := range sectionCommits {
			sb.WriteString(formatConventionalEntry(c, repoInfo))
		}
	}

	// Breaking changes section
	breakingCommits := collectBreakingChanges(commits)
	if len(breakingCommits) > 0 {
		sb.WriteString("\n### BREAKING CHANGES\n\n")
		for _, c := range breakingCommits {
			msg := c.BreakingMessage
			if c.Scope != "" {
				fmt.Fprintf(&sb, "* **%s:** %s\n", c.Scope, msg)
			} else {
				fmt.Fprintf(&sb, "* %s\n", msg)
			}
		}
	}

	return sb.String()
}

// formatConventionalEntry formats a single commit entry for conventional-changelog.
func formatConventionalEntry(c *Commit, repoInfo *git.RepoInfo) string {
	var entry string
	shortHash := shortHash(c.Hash)

	hashRef := shortHash
	if repoInfo != nil && shortHash != "" {
		hashRef = fmt.Sprintf("[%s](https://%s/%s/%s/commit/%s)",
			shortHash, repoInfo.Host, repoInfo.Owner, repoInfo.Repository, c.Hash)
	}

	if c.Scope != "" {
		if hashRef != "" {
			entry = fmt.Sprintf("* **%s:** %s (%s)\n", c.Scope, c.Description, hashRef)
		} else {
			entry = fmt.Sprintf("* **%s:** %s\n", c.Scope, c.Description)
		}
	} else {
		if hashRef != "" {
			entry = fmt.Sprintf("* %s (%s)\n", c.Description, hashRef)
		} else {
			entry = fmt.Sprintf("* %s\n", c.Description)
		}
	}

	return entry
}

// groupBySection groups commits by their changelog section heading.
func groupBySection(commits []*Commit, typeMap map[string]string) map[string][]*Commit {
	sections := make(map[string][]*Commit)
	for _, c := range commits {
		section, ok := typeMap[c.Type]
		if !ok {
			continue
		}
		sections[section] = append(sections[section], c)
	}

	// Sort commits within each section by scope then description
	for _, sectionCommits := range sections {
		sort.Slice(sectionCommits, func(i, j int) bool {
			if sectionCommits[i].Scope != sectionCommits[j].Scope {
				return sectionCommits[i].Scope < sectionCommits[j].Scope
			}
			return sectionCommits[i].Description < sectionCommits[j].Description
		})
	}

	return sections
}

// collectBreakingChanges filters commits with breaking changes.
func collectBreakingChanges(commits []*Commit) []*Commit {
	var breaking []*Commit
	for _, c := range commits {
		if c.BreakingChange {
			breaking = append(breaking, c)
		}
	}
	return breaking
}

// shortHash returns the first 7 characters of a hash.
func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
