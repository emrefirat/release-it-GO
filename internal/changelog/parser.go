package changelog

import (
	"regexp"
	"strings"
)

// commitPattern matches conventional commit format:
// type(scope)!: description
var commitPattern = regexp.MustCompile(
	`^(?P<type>\w+)(?:\((?P<scope>[^)]+)\))?(?P<breaking>!)?:\s+(?P<description>.+)$`,
)

// ParseCommit parses a single commit message into a Commit struct.
// Returns nil if the message does not match conventional commit format.
func ParseCommit(raw string, hash string) *Commit {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	lines := strings.SplitN(raw, "\n", 2)
	subject := strings.TrimSpace(lines[0])

	matches := commitPattern.FindStringSubmatch(subject)
	if matches == nil {
		return nil
	}

	commit := &Commit{
		Hash:        hash,
		Type:        matches[commitPattern.SubexpIndex("type")],
		Description: matches[commitPattern.SubexpIndex("description")],
		Raw:         raw,
	}

	if idx := commitPattern.SubexpIndex("scope"); idx >= 0 && idx < len(matches) {
		commit.Scope = matches[idx]
	}

	if idx := commitPattern.SubexpIndex("breaking"); idx >= 0 && idx < len(matches) && matches[idx] == "!" {
		commit.BreakingChange = true
		commit.BreakingMessage = commit.Description
	}

	// Parse body and footers
	if len(lines) > 1 {
		remaining := strings.TrimSpace(lines[1])
		commit.Body, commit.Footers = parseBodyAndFooters(remaining)

		// Check for BREAKING CHANGE footer
		for _, f := range commit.Footers {
			if f.Token == "BREAKING CHANGE" || f.Token == "BREAKING-CHANGE" {
				commit.BreakingChange = true
				commit.BreakingMessage = f.Value
			}
		}
	}

	return commit
}

// parseBodyAndFooters splits the remaining commit text into body and footers.
// Footers follow the git trailer convention: "Token: value" or "Token #value".
func parseBodyAndFooters(text string) (string, []Footer) {
	if text == "" {
		return "", nil
	}

	lines := strings.Split(text, "\n")
	var footers []Footer
	var bodyLines []string

	// Scan from the end to find where footers begin
	footerStartIdx := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			// Empty line breaks footer section from body
			break
		}
		if isFooterLine(line) {
			footerStartIdx = i
		} else {
			break
		}
	}

	bodyLines = lines[:footerStartIdx]
	for i := footerStartIdx; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		token, value := parseFooterLine(line)
		if token != "" {
			footers = append(footers, Footer{Token: token, Value: value})
		}
	}

	body := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	return body, footers
}

// footerPattern matches git trailer format: "Token: value" or "BREAKING CHANGE: value"
var footerPattern = regexp.MustCompile(`^([A-Za-z-]+(?:\s+[A-Z]+)?)\s*[:#]\s*(.+)$`)

// isFooterLine checks if a line looks like a git trailer.
func isFooterLine(line string) bool {
	return footerPattern.MatchString(line)
}

// parseFooterLine extracts token and value from a footer line.
func parseFooterLine(line string) (string, string) {
	matches := footerPattern.FindStringSubmatch(line)
	if matches == nil {
		return "", ""
	}
	return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
}

// ParseCommits parses multiple raw commits. Non-conventional commits are skipped.
func ParseCommits(rawCommits []RawCommit) []*Commit {
	commits := make([]*Commit, 0, len(rawCommits))
	for _, rc := range rawCommits {
		c := ParseCommit(rc.Message, rc.Hash)
		if c != nil {
			commits = append(commits, c)
		}
	}
	return commits
}
