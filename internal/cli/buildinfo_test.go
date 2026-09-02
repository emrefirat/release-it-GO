package cli

import (
	"runtime/debug"
	"strings"
	"testing"
)

func fakeBuildInfo(mainVersion string, settings map[string]string) *debug.BuildInfo {
	info := &debug.BuildInfo{}
	info.Main.Version = mainVersion
	for k, v := range settings {
		info.Settings = append(info.Settings, debug.BuildSetting{Key: k, Value: v})
	}
	return info
}

func TestResolveBuildInfo(t *testing.T) {
	ldDefaults := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	tests := []struct {
		name string
		ld   BuildInfo
		info *debug.BuildInfo
		want BuildInfo
	}{
		{
			name: "ldflags win over build info (GoReleaser, make build)",
			ld:   BuildInfo{Version: "0.4.0", Commit: "b2afd41", Date: "2026-09-02T21:33:47Z"},
			info: fakeBuildInfo("v9.9.9", map[string]string{"vcs.revision": "ffffffffffff", "vcs.time": "2000-01-01T00:00:00Z"}),
			want: BuildInfo{Version: "0.4.0", Commit: "b2afd41", Date: "2026-09-02T21:33:47Z"},
		},
		{
			name: "go install pkg@v0.4.0: module version, no VCS stamps",
			ld:   ldDefaults,
			info: fakeBuildInfo("v0.4.0", nil),
			want: BuildInfo{Version: "0.4.0", Commit: "none", Date: "unknown"},
		},
		{
			name: "go install pkg@main: pseudo-version carries the commit",
			ld:   ldDefaults,
			info: fakeBuildInfo("v0.4.1-0.20260903103000-b2afd41c0ffe", nil),
			want: BuildInfo{Version: "0.4.1-0.20260903103000-b2afd41c0ffe", Commit: "b2afd41", Date: "unknown"},
		},
		{
			name: "go build in the repo: (devel) with VCS stamps",
			ld:   ldDefaults,
			info: fakeBuildInfo("(devel)", map[string]string{
				"vcs.revision": "b2afd41c0ffee1234567890abcdef1234567890a",
				"vcs.time":     "2026-09-02T21:33:47Z",
				"vcs.modified": "false",
			}),
			want: BuildInfo{Version: "dev", Commit: "b2afd41", Date: "2026-09-02T21:33:47Z"},
		},
		{
			name: "go build with uncommitted changes is marked dirty",
			ld:   ldDefaults,
			info: fakeBuildInfo("(devel)", map[string]string{"vcs.revision": "b2afd41c0ffee", "vcs.modified": "true"}),
			want: BuildInfo{Version: "dev", Commit: "b2afd41-dirty", Date: "unknown"},
		},
		{
			// Observed on Go 1.24+: a go build inside the repository stamps a
			// pseudo-version with +dirty instead of "(devel)"; VCS wins for commit.
			name: "go build in the repo (Go 1.24+): pseudo-version +dirty with VCS stamps",
			ld:   ldDefaults,
			info: fakeBuildInfo("v0.4.1-0.20260902214320-0ca1a8725a67+dirty", map[string]string{
				"vcs.revision": "0ca1a8725a67deadbeef",
				"vcs.time":     "2026-09-02T21:43:20Z",
				"vcs.modified": "true",
			}),
			want: BuildInfo{Version: "0.4.1-0.20260902214320-0ca1a8725a67+dirty", Commit: "0ca1a87-dirty", Date: "2026-09-02T21:43:20Z"},
		},
		{
			name: "no build info at all keeps the defaults",
			ld:   ldDefaults,
			info: nil,
			want: ldDefaults,
		},
		{
			name: "partial ldflags: only the missing fields are filled in",
			ld:   BuildInfo{Version: "0.4.0", Commit: "none", Date: "unknown"},
			info: fakeBuildInfo("(devel)", map[string]string{"vcs.revision": "abcdef0123456", "vcs.time": "2026-09-03T00:00:00Z"}),
			want: BuildInfo{Version: "0.4.0", Commit: "abcdef0", Date: "2026-09-03T00:00:00Z"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveBuildInfo(tt.ld, tt.info); got != tt.want {
				t.Errorf("resolveBuildInfo() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// The version subcommand must print the resolved values, not the raw ldflags
// variables — the go install case prints "dev" otherwise.
func TestVersionCommand_UsesResolvedBuildInfo(t *testing.T) {
	origRead := readBuildInfo
	origV, origC, origD := Version, Commit, Date
	t.Cleanup(func() { readBuildInfo = origRead; Version, Commit, Date = origV, origC, origD })

	Version, Commit, Date = "dev", "none", "unknown"
	readBuildInfo = func() *debug.BuildInfo { return fakeBuildInfo("v0.4.0", nil) }

	var out strings.Builder
	cmd := newVersionCommand()
	cmd.SetOut(&out)
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version: %v", err)
	}
	want := "release-it-go 0.4.0 (commit: none, built: unknown)\n"
	if out.String() != want {
		t.Errorf("version output = %q, want %q", out.String(), want)
	}
}
