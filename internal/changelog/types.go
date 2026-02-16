// Package changelog provides conventional commit parsing, bump analysis,
// and changelog generation in conventional-changelog and keep-a-changelog formats.
package changelog

// Commit represents a parsed conventional commit.
type Commit struct {
	Hash            string
	Type            string   // feat, fix, docs, etc.
	Scope           string   // optional scope
	Description     string   // commit description
	Body            string   // optional body
	Footers         []Footer // parsed footers
	BreakingChange  bool
	BreakingMessage string // BREAKING CHANGE description
	Raw             string // original commit message
}

// Footer represents a git trailer / commit footer.
type Footer struct {
	Token string // "BREAKING CHANGE", "Closes", "Refs", etc.
	Value string
}

// RawCommit holds the hash and full message of a git commit before parsing.
type RawCommit struct {
	Hash    string
	Message string
}

// BumpType represents the semantic version increment type.
type BumpType int

const (
	// BumpNone indicates no version change is needed.
	BumpNone BumpType = iota
	// BumpPatch indicates a patch version increment.
	BumpPatch
	// BumpMinor indicates a minor version increment.
	BumpMinor
	// BumpMajor indicates a major version increment.
	BumpMajor
)

// String returns the string representation of a BumpType.
func (b BumpType) String() string {
	switch b {
	case BumpPatch:
		return "patch"
	case BumpMinor:
		return "minor"
	case BumpMajor:
		return "major"
	default:
		return ""
	}
}

// commitTypeSection maps conventional commit types to changelog section headings
// for the conventional-changelog format.
var commitTypeSection = map[string]string{
	"feat":   "Features",
	"fix":    "Bug Fixes",
	"perf":   "Performance Improvements",
	"revert": "Reverts",
}

// commitTypeKeepAChangelog maps conventional commit types to keep-a-changelog sections.
var commitTypeKeepAChangelog = map[string]string{
	"feat":     "Added",
	"fix":      "Fixed",
	"perf":     "Changed",
	"refactor": "Changed",
	"revert":   "Removed",
}
