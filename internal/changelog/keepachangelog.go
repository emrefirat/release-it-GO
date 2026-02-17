package changelog

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// RenderKeepAChangelog generates changelog content in keep-a-changelog format.
// Groups commits into Added, Changed, Fixed, Removed sections.
func RenderKeepAChangelog(commits []*Commit, version string, date string) string {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "## [%s] - %s\n", version, date)

	// Group commits by keep-a-changelog sections
	sections := groupBySection(commits, commitTypeKeepAChangelog)

	// Render in keep-a-changelog order
	sectionOrder := []string{"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"}
	for _, sectionName := range sectionOrder {
		sectionCommits, ok := sections[sectionName]
		if !ok || len(sectionCommits) == 0 {
			continue
		}

		fmt.Fprintf(&sb, "\n### %s\n\n", sectionName)

		// Sort by description
		sort.Slice(sectionCommits, func(i, j int) bool {
			return sectionCommits[i].Description < sectionCommits[j].Description
		})

		for _, c := range sectionCommits {
			if c.Scope != "" {
				fmt.Fprintf(&sb, "- %s (%s)\n", c.Description, c.Scope)
			} else {
				fmt.Fprintf(&sb, "- %s\n", c.Description)
			}
		}
	}

	return sb.String()
}
