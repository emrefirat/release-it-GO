package changelog

import "testing"

func TestAnalyzeBump(t *testing.T) {
	tests := []struct {
		name    string
		commits []*Commit
		want    BumpType
	}{
		{
			name:    "no commits",
			commits: nil,
			want:    BumpNone,
		},
		{
			name: "only feat",
			commits: []*Commit{
				{Type: "feat", Description: "add feature"},
			},
			want: BumpMinor,
		},
		{
			name: "only fix",
			commits: []*Commit{
				{Type: "fix", Description: "fix bug"},
			},
			want: BumpPatch,
		},
		{
			name: "feat and fix",
			commits: []*Commit{
				{Type: "fix", Description: "fix bug"},
				{Type: "feat", Description: "add feature"},
			},
			want: BumpMinor,
		},
		{
			name: "breaking change overrides all",
			commits: []*Commit{
				{Type: "feat", Description: "add feature"},
				{Type: "fix", Description: "fix bug", BreakingChange: true, BreakingMessage: "API changed"},
			},
			want: BumpMajor,
		},
		{
			name: "only docs and chore",
			commits: []*Commit{
				{Type: "docs", Description: "update readme"},
				{Type: "chore", Description: "update deps"},
			},
			want: BumpNone,
		},
		{
			name: "perf is patch",
			commits: []*Commit{
				{Type: "perf", Description: "optimize query"},
			},
			want: BumpPatch,
		},
		{
			name: "revert is patch",
			commits: []*Commit{
				{Type: "revert", Description: "revert: add feature"},
			},
			want: BumpPatch,
		},
		{
			name: "mixed with breaking",
			commits: []*Commit{
				{Type: "docs", Description: "update docs"},
				{Type: "feat", Description: "add feature", BreakingChange: true, BreakingMessage: "removed old API"},
				{Type: "fix", Description: "fix bug"},
			},
			want: BumpMajor,
		},
		{
			name: "only style and test",
			commits: []*Commit{
				{Type: "style", Description: "format code"},
				{Type: "test", Description: "add tests"},
				{Type: "ci", Description: "update ci"},
			},
			want: BumpNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnalyzeBump(tt.commits)
			if got != tt.want {
				t.Errorf("AnalyzeBump() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalyzeBumpWithConfig(t *testing.T) {
	commits := []*Commit{
		{Type: "feat", Description: "add feature"},
	}

	got := AnalyzeBumpWithConfig(commits, "angular")
	if got != BumpMinor {
		t.Errorf("AnalyzeBumpWithConfig() = %v, want minor", got)
	}
}
