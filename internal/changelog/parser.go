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

// gitRevertPattern matches the subject `git revert` generates.
var gitRevertPattern = regexp.MustCompile(`^Revert "(.+)"$`)

// ParseCommit parses a single commit message into a Commit struct.
// Returns nil if the message does not match conventional commit format.
func ParseCommit(raw string, hash string) *Commit {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	lines := strings.SplitN(raw, "\n", 2)
	subject := strings.TrimSpace(lines[0])

	// git's own revert subject: Revert "feat(api): add x". Treated as a
	// revert-typed commit so the bump analysis and the "Reverts" section see
	// it; the linter already accepts it.
	if m := gitRevertPattern.FindStringSubmatch(subject); m != nil {
		commit := &Commit{Hash: hash, Type: "revert", Description: m[1], Raw: raw}
		if len(lines) > 1 {
			commit.Body, commit.Footers = parseBodyAndFooters(strings.TrimSpace(lines[1]))
		}
		return commit
	}

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
// A footer value may span multiple lines (git trailers and the conventional
// commits spec both allow it): non-footer lines in the trailing section are
// folded into the preceding footer's value.
func parseBodyAndFooters(text string) (string, []Footer) {
	if text == "" {
		return "", nil
	}

	lines := strings.Split(text, "\n")
	var footers []Footer

	// Scan from the end to find where footers begin. A non-footer line is a
	// continuation of a footer only when a footer-like line sits above it
	// before any blank line.
	footerStartIdx := len(lines)
	i := len(lines) - 1
	for i >= 0 {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			// Empty line breaks footer section from body
			break
		}
		if isFooterLine(line) {
			footerStartIdx = i
			i--
			continue
		}
		// Possible continuation line — look upward for its footer token.
		j := i - 1
		for j >= 0 && strings.TrimSpace(lines[j]) != "" && !isFooterLine(strings.TrimSpace(lines[j])) {
			j--
		}
		if j >= 0 && strings.TrimSpace(lines[j]) != "" && isFooterLine(strings.TrimSpace(lines[j])) {
			footerStartIdx = j
			i = j - 1
			continue
		}
		break
	}

	footers = parseFooterLines(lines[footerStartIdx:])

	// BREAKING CHANGE paragraphs are not always the last one: GitHub's
	// squash-merge appends "Co-authored-by:" as a separate paragraph after
	// them, and trailing-paragraph-only detection then demoted a major bump
	// to minor. Like conventional-commits-parser, honor a paragraph that
	// starts with the breaking token anywhere in the body.
	var bodyParagraphs []string
	var breakingFooters []Footer
	for _, paragraph := range splitParagraphs(lines[:footerStartIdx]) {
		if breakingTokenPattern.MatchString(strings.TrimSpace(paragraph[0])) {
			breakingFooters = append(breakingFooters, parseFooterLines(paragraph)...)
			continue
		}
		bodyParagraphs = append(bodyParagraphs, strings.Join(paragraph, "\n"))
	}
	footers = append(breakingFooters, footers...)

	body := strings.TrimSpace(strings.Join(bodyParagraphs, "\n\n"))
	return body, footers
}

// parseFooterLines turns a block of footer lines into footers. Non-footer
// lines are folded into the preceding footer's value.
func parseFooterLines(lines []string) []Footer {
	var footers []Footer
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if isFooterLine(line) {
			token, value := parseFooterLine(line)
			if token != "" {
				footers = append(footers, Footer{Token: token, Value: value})
			}
			continue
		}
		// Continuation of the previous footer's value
		if len(footers) > 0 {
			footers[len(footers)-1].Value += "\n" + line
		}
	}
	return footers
}

// splitParagraphs groups lines into blank-line-separated paragraphs,
// dropping the blank lines themselves.
func splitParagraphs(lines []string) [][]string {
	var paragraphs [][]string
	var current []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if len(current) > 0 {
				paragraphs = append(paragraphs, current)
				current = nil
			}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		paragraphs = append(paragraphs, current)
	}
	return paragraphs
}

// breakingTokenPattern matches the start of a BREAKING CHANGE footer line.
var breakingTokenPattern = regexp.MustCompile(`^BREAKING[ -]CHANGE\s*[:#]`)

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
