package cli

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
)

// BuildInfo describes the running binary as printed by `release-it-go version`.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// Sentinel values of the ldflags variables when no -X flag was given.
const (
	unsetVersion = "dev"
	unsetCommit  = "none"
	unsetDate    = "unknown"
)

// readBuildInfo wraps debug.ReadBuildInfo; tests replace it.
var readBuildInfo = func() *debug.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return info
}

// pseudoVersionHash captures the commit hash of a Go pseudo-version; the
// timestamp always follows a dot (v0.4.1-0.20260903103000-b2afd41c0ffe,
// v0.5.0-beta.0.20260903103000-b2afd41c0ffe).
var pseudoVersionHash = regexp.MustCompile(`\.\d{14}-([0-9a-f]{12})$`)

// shortHashLen is the abbreviated commit length used everywhere in the output.
const shortHashLen = 7

// resolveBuildInfo fills in whatever the ldflags did not set from the Go
// toolchain's own stamps: GoReleaser and `make build` inject -X values and
// win; `go install pkg@version` has none, but the module version (and the
// commit inside a pseudo-version) is recorded; a plain `go build` in the
// repository records the VCS revision, time and dirty state.
func resolveBuildInfo(ld BuildInfo, info *debug.BuildInfo) BuildInfo {
	out := ld
	if info == nil {
		return out
	}

	moduleVersion := info.Main.Version
	if out.Version == unsetVersion && moduleVersion != "" && moduleVersion != "(devel)" {
		out.Version = strings.TrimPrefix(moduleVersion, "v")
	}

	settings := make(map[string]string, len(info.Settings))
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}

	if out.Commit == unsetCommit {
		switch {
		case settings["vcs.revision"] != "":
			out.Commit = shortHash(settings["vcs.revision"])
			if settings["vcs.modified"] == "true" {
				out.Commit += "-dirty"
			}
		default:
			if m := pseudoVersionHash.FindStringSubmatch(moduleVersion); m != nil {
				out.Commit = shortHash(m[1])
			}
		}
	}

	if out.Date == unsetDate && settings["vcs.time"] != "" {
		out.Date = settings["vcs.time"]
	}
	return out
}

func shortHash(rev string) string {
	if len(rev) > shortHashLen {
		return rev[:shortHashLen]
	}
	return rev
}

// formatBuildInfo renders the version line.
func formatBuildInfo(bi BuildInfo) string {
	return fmt.Sprintf("release-it-go %s (commit: %s, built: %s)\n", bi.Version, bi.Commit, bi.Date)
}
