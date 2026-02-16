package version

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FallbackVersion is used when no version can be detected.
const FallbackVersion = "0.0.0"

// VersionOptions configures how the latest version is detected from git tags.
type VersionOptions struct {
	TagMatch       string
	TagExclude     string
	GetFromAllRefs bool
}

// GetLatestTagVersion returns the latest version from git tags.
// Uses tagMatch/tagExclude patterns to filter tags.
func GetLatestTagVersion(opts VersionOptions) (string, error) {
	if opts.GetFromAllRefs {
		return getLatestTagFromAllRefs(opts)
	}

	out, err := runGit("describe", "--tags", "--abbrev=0")
	if err != nil {
		return "", fmt.Errorf("no git tags found: %w", err)
	}

	tag := strings.TrimSpace(out)
	if tag == "" {
		return "", fmt.Errorf("no git tags found")
	}

	if opts.TagMatch != "" && !matchPattern(tag, opts.TagMatch) {
		return "", fmt.Errorf("latest tag %q does not match pattern %q", tag, opts.TagMatch)
	}

	if opts.TagExclude != "" && matchPattern(tag, opts.TagExclude) {
		return "", fmt.Errorf("latest tag %q is excluded by pattern %q", tag, opts.TagExclude)
	}

	return strings.TrimPrefix(tag, "v"), nil
}

// getLatestTagFromAllRefs gets the latest tag from all refs, sorted by semver.
func getLatestTagFromAllRefs(opts VersionOptions) (string, error) {
	out, err := runGit("tag", "-l", "--sort=-v:refname")
	if err != nil {
		return "", fmt.Errorf("listing git tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(out), "\n")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		if opts.TagMatch != "" && !matchPattern(tag, opts.TagMatch) {
			continue
		}

		if opts.TagExclude != "" && matchPattern(tag, opts.TagExclude) {
			continue
		}

		return strings.TrimPrefix(tag, "v"), nil
	}

	return "", fmt.Errorf("no matching git tags found")
}

// GetVersionFromFile reads a version string from the given file path.
func GetVersionFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("reading version file %s: %w", filePath, err)
	}

	version := strings.TrimSpace(string(data))
	if version == "" {
		return "", fmt.Errorf("version file %s is empty", filePath)
	}

	return strings.TrimPrefix(version, "v"), nil
}

// DetectVersion tries multiple sources to determine the current version.
// Priority: git tag > VERSION file > fallback (0.0.0).
func DetectVersion(opts VersionOptions) string {
	if v, err := GetLatestTagVersion(opts); err == nil && v != "" {
		return v
	}

	if v, err := GetVersionFromFile("VERSION"); err == nil && v != "" {
		return v
	}

	return FallbackVersion
}

// runGit executes a git command and returns its stdout.
var runGit = func(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// matchPattern performs simple glob-like matching using filepath.Match semantics.
// For basic use: "*" matches anything, "?" matches single char.
func matchPattern(s, pattern string) bool {
	// Simple glob: use strings for basic matching
	if pattern == "*" {
		return true
	}

	// Handle prefix/suffix wildcards
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(s, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(s, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(s, pattern[:len(pattern)-1])
	}

	// Exact match fallback
	return s == pattern
}
