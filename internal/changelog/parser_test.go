package changelog

import "testing"

func TestParseCommit(t *testing.T) {
	tests := []struct {
		name            string
		raw             string
		hash            string
		wantType        string
		wantScope       string
		wantDescription string
		wantBreaking    bool
		wantNil         bool
	}{
		{
			name:            "simple feat",
			raw:             "feat: add login",
			hash:            "abc1234",
			wantType:        "feat",
			wantDescription: "add login",
		},
		{
			name:            "fix with scope",
			raw:             "fix(api): rate limit issue",
			hash:            "def5678",
			wantType:        "fix",
			wantScope:       "api",
			wantDescription: "rate limit issue",
		},
		{
			name:            "breaking change with bang",
			raw:             "feat!: new API",
			hash:            "bbb1111",
			wantType:        "feat",
			wantDescription: "new API",
			wantBreaking:    true,
		},
		{
			name:            "breaking with scope and bang",
			raw:             "refactor(core)!: rewrite engine",
			hash:            "ccc2222",
			wantType:        "refactor",
			wantScope:       "core",
			wantDescription: "rewrite engine",
			wantBreaking:    true,
		},
		{
			name:         "breaking change in footer",
			raw:          "feat: add oauth\n\nSome body text\n\nBREAKING CHANGE: removed basic auth",
			hash:         "ddd3333",
			wantType:     "feat",
			wantBreaking: true,
		},
		{
			name:    "non-conventional commit",
			raw:     "random message without type",
			hash:    "eee4444",
			wantNil: true,
		},
		{
			name:    "empty message",
			raw:     "",
			hash:    "fff5555",
			wantNil: true,
		},
		{
			name:    "whitespace only",
			raw:     "   \n  ",
			hash:    "ggg6666",
			wantNil: true,
		},
		{
			name:            "docs type",
			raw:             "docs: update readme",
			hash:            "hhh7777",
			wantType:        "docs",
			wantDescription: "update readme",
		},
		{
			name:            "chore type",
			raw:             "chore: update deps",
			hash:            "iii8888",
			wantType:        "chore",
			wantDescription: "update deps",
		},
		{
			name:            "commit with body",
			raw:             "feat: add feature\n\nThis is the body of the commit.\nIt has multiple lines.",
			hash:            "jjj9999",
			wantType:        "feat",
			wantDescription: "add feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCommit(tt.raw, tt.hash)

			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseCommit() = %+v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("ParseCommit() = nil, want non-nil")
			}

			if got.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tt.wantType)
			}
			if got.Scope != tt.wantScope {
				t.Errorf("Scope = %q, want %q", got.Scope, tt.wantScope)
			}
			if got.Description != tt.wantDescription && tt.wantDescription != "" {
				t.Errorf("Description = %q, want %q", got.Description, tt.wantDescription)
			}
			if got.BreakingChange != tt.wantBreaking {
				t.Errorf("BreakingChange = %v, want %v", got.BreakingChange, tt.wantBreaking)
			}
			if got.Hash != tt.hash {
				t.Errorf("Hash = %q, want %q", got.Hash, tt.hash)
			}
		})
	}
}

func TestParseCommit_BodyAndFooters(t *testing.T) {
	raw := "feat(auth): add OAuth2\n\nImplements OAuth2 flow.\n\nCloses: #123\nRefs: #456"
	c := ParseCommit(raw, "abc1234")

	if c == nil {
		t.Fatal("ParseCommit() = nil, want non-nil")
	}

	if c.Body != "Implements OAuth2 flow." {
		t.Errorf("Body = %q, want 'Implements OAuth2 flow.'", c.Body)
	}

	if len(c.Footers) != 2 {
		t.Fatalf("len(Footers) = %d, want 2", len(c.Footers))
	}

	if c.Footers[0].Token != "Closes" {
		t.Errorf("Footers[0].Token = %q, want 'Closes'", c.Footers[0].Token)
	}
	if c.Footers[0].Value != "#123" {
		t.Errorf("Footers[0].Value = %q, want '#123'", c.Footers[0].Value)
	}
	if c.Footers[1].Token != "Refs" {
		t.Errorf("Footers[1].Token = %q, want 'Refs'", c.Footers[1].Token)
	}
}

func TestParseCommit_BreakingChangeFooter(t *testing.T) {
	raw := "feat: add new API\n\nBREAKING CHANGE: removed old endpoints"
	c := ParseCommit(raw, "abc1234")

	if c == nil {
		t.Fatal("ParseCommit() = nil")
	}
	if !c.BreakingChange {
		t.Error("BreakingChange = false, want true")
	}
	if c.BreakingMessage != "removed old endpoints" {
		t.Errorf("BreakingMessage = %q, want 'removed old endpoints'", c.BreakingMessage)
	}
}

func TestParseCommits(t *testing.T) {
	rawCommits := []RawCommit{
		{Hash: "aaa", Message: "feat: add feature"},
		{Hash: "bbb", Message: "not a conventional commit"},
		{Hash: "ccc", Message: "fix(api): fix bug"},
		{Hash: "ddd", Message: ""},
		{Hash: "eee", Message: "chore: update deps"},
	}

	commits := ParseCommits(rawCommits)

	if len(commits) != 3 {
		t.Fatalf("len(commits) = %d, want 3 (non-conventional and empty should be skipped)", len(commits))
	}

	if commits[0].Type != "feat" {
		t.Errorf("commits[0].Type = %q, want 'feat'", commits[0].Type)
	}
	if commits[1].Type != "fix" {
		t.Errorf("commits[1].Type = %q, want 'fix'", commits[1].Type)
	}
	if commits[2].Type != "chore" {
		t.Errorf("commits[2].Type = %q, want 'chore'", commits[2].Type)
	}
}

func TestParseCommits_Empty(t *testing.T) {
	commits := ParseCommits(nil)
	if len(commits) != 0 {
		t.Errorf("len(commits) = %d, want 0", len(commits))
	}
}
