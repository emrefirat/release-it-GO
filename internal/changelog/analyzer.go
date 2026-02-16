package changelog

// bumpTypeForCommit maps commit types to their default bump types.
var bumpTypeForCommit = map[string]BumpType{
	"feat":   BumpMinor,
	"fix":    BumpPatch,
	"perf":   BumpPatch,
	"revert": BumpPatch,
}

// AnalyzeBump determines the recommended version bump from parsed commits.
// Returns the highest severity bump found:
//   - Breaking change -> major
//   - feat -> minor
//   - fix, perf, revert -> patch
//   - No matching types -> none
func AnalyzeBump(commits []*Commit) BumpType {
	return AnalyzeBumpWithConfig(commits, "angular")
}

// AnalyzeBumpWithConfig determines the bump type using the specified preset.
// Currently only "angular" preset is supported.
func AnalyzeBumpWithConfig(commits []*Commit, preset string) BumpType {
	result := BumpNone

	for _, c := range commits {
		if c.BreakingChange {
			return BumpMajor
		}

		if bump, ok := bumpTypeForCommit[c.Type]; ok {
			if bump > result {
				result = bump
			}
		}
	}

	return result
}
