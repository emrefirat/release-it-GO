package changelog

import "strings"

// TypeInfo describes an accepted conventional commit type.
type TypeInfo struct {
	Name        string
	Description string
}

// validTypes is the single source of truth for accepted commit types (Angular
// preset) in display order. The linter's allow-list and every help text are
// derived from it so they can never disagree again.
var validTypes = []TypeInfo{
	{"feat", "new feature"},
	{"fix", "bug fix"},
	{"docs", "documentation only"},
	{"style", "formatting / whitespace"},
	{"refactor", "code restructuring"},
	{"perf", "performance improvement"},
	{"test", "add or fix tests"},
	{"build", "build system / dependencies"},
	{"ci", "CI/CD changes"},
	{"chore", "maintenance tasks"},
	{"revert", "revert a commit"},
}

// allowedTypes is the lookup set derived from validTypes.
var allowedTypes = func() map[string]bool {
	m := make(map[string]bool, len(validTypes))
	for _, t := range validTypes {
		m[t.Name] = true
	}
	return m
}()

// ValidTypes returns the accepted commit types with descriptions, in
// display order.
func ValidTypes() []TypeInfo {
	out := make([]TypeInfo, len(validTypes))
	copy(out, validTypes)
	return out
}

// ValidTypeNames returns the accepted commit type names in display order.
func ValidTypeNames() []string {
	names := make([]string, len(validTypes))
	for i, t := range validTypes {
		names[i] = t.Name
	}
	return names
}

// autoPassPrefixes are subjects the linter always accepts: merge and revert
// commits (git-generated) and fixup!/squash!/amend! commits, which exist to
// be squashed before release — commitlint ignores them by default, and the
// commit-msg hook must not reject `git commit --fixup`.
var autoPassPrefixes = []struct {
	prefix string
	reason string
}{
	{"Merge ", "merge commit"},
	{"Revert ", "revert commit"},
	{"fixup! ", "fixup commit (squash before release)"},
	{"squash! ", "squash commit (squash before release)"},
	{"amend! ", "amend commit (squash before release)"},
}

// SuggestType returns the accepted type closest to an unknown one: a
// case-only mismatch ("Feat"), or a typo within two edits ("fic" → fix,
// "chroe" → chore). Empty when the type is already valid or nothing is close.
func SuggestType(unknown string) string {
	if unknown == "" || allowedTypes[unknown] {
		return ""
	}
	lower := strings.ToLower(unknown)
	if allowedTypes[lower] {
		return lower
	}
	best, bestDist := "", 3
	for _, t := range validTypes {
		if d := levenshtein(lower, t.Name); d < bestDist {
			best, bestDist = t.Name, d
		}
	}
	return best
}

// levenshtein computes the edit distance between two short ASCII strings.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
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
	// Suggestion is the closest accepted type when Reason is an unknown type.
	Suggestion string
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

		if reason, ok := autoPassReason(subject); ok {
			passed = append(passed, LintResult{
				Hash:    c.Hash,
				Subject: subject,
				Valid:   true,
				Reason:  reason,
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
				Hash:       c.Hash,
				Subject:    subject,
				Valid:      false,
				Reason:     "unknown type: " + commitType,
				Suggestion: SuggestType(commitType),
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

// autoPassReason reports whether the subject is always accepted, with why.
func autoPassReason(subject string) (string, bool) {
	for _, p := range autoPassPrefixes {
		if strings.HasPrefix(subject, p.prefix) {
			return p.reason, true
		}
	}
	return "", false
}
