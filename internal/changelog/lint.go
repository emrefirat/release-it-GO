package changelog

import "strings"

// allowedTypes defines the valid conventional commit types (Angular preset).
var allowedTypes = map[string]bool{
	"feat":     true,
	"fix":      true,
	"docs":     true,
	"style":    true,
	"refactor": true,
	"perf":     true,
	"test":     true,
	"build":    true,
	"ci":       true,
	"chore":    true,
	"revert":   true,
}

// LintInput represents a commit to be linted.
type LintInput struct {
	Hash    string
	Subject string
}

// LintResult represents the lint result for a single commit.
type LintResult struct {
	Hash    string
	Subject string
	Valid   bool
	Reason  string
}

// LintCommits checks whether each commit follows the conventional commit format.
// Validates both the format and the commit type against the Angular preset.
// Merge commits and revert commits are automatically passed.
// Returns lists of passed and failed results.
func LintCommits(commits []LintInput) (passed, failed []LintResult) {
	passed = make([]LintResult, 0, len(commits))
	failed = make([]LintResult, 0)

	for _, c := range commits {
		subject := strings.TrimSpace(c.Subject)

		// Auto-pass merge commits
		if strings.HasPrefix(subject, "Merge ") {
			passed = append(passed, LintResult{
				Hash:    c.Hash,
				Subject: subject,
				Valid:   true,
				Reason:  "merge commit",
			})
			continue
		}

		// Auto-pass revert commits
		if strings.HasPrefix(subject, "Revert ") {
			passed = append(passed, LintResult{
				Hash:    c.Hash,
				Subject: subject,
				Valid:   true,
				Reason:  "revert commit",
			})
			continue
		}

		// Check against conventional commit pattern
		matches := commitPattern.FindStringSubmatch(subject)
		if matches == nil {
			failed = append(failed, LintResult{
				Hash:    c.Hash,
				Subject: subject,
				Valid:   false,
				Reason:  "not in conventional commit format",
			})
			continue
		}

		// Validate the commit type against allowed types
		commitType := matches[commitPattern.SubexpIndex("type")]
		if !allowedTypes[commitType] {
			failed = append(failed, LintResult{
				Hash:    c.Hash,
				Subject: subject,
				Valid:   false,
				Reason:  "unknown type: " + commitType,
			})
			continue
		}

		passed = append(passed, LintResult{
			Hash:    c.Hash,
			Subject: subject,
			Valid:   true,
			Reason:  "conventional commit",
		})
	}

	return passed, failed
}
