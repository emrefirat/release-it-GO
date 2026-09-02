package git

import (
	"fmt"
	"path"
	"strings"

	"github.com/Masterminds/semver/v3"

	"release-it-go/internal/version"
)

// CreateTag creates an annotated tag with the given name and message.
// Additional tag arguments from config are appended.
func (g *Git) CreateTag(tagName string, annotation string) error {
	exists, err := g.TagExists(tagName)
	if err != nil {
		return fmt.Errorf("checking tag existence: %w", err)
	}
	if exists {
		return fmt.Errorf("tag %s already exists", tagName)
	}

	args := []string{"tag", "--annotate", "--message", annotation, tagName}
	args = append(args, g.config.TagArgs...)
	_, err = g.run(args...)
	return err
}

// GetLatestTag returns the most recent tag, preferring tags that match the current
// tagName format. If no matching tag is found, falls back to any tag to preserve
// version continuity during format transitions (e.g., "v${version}" → "${version}").
func (g *Git) GetLatestTag() (string, error) {
	if g.config.GetLatestTagFromAllRefs {
		tag, err := g.getLatestTagFromAllRefs()
		if err == nil {
			return tag, nil
		}
		if g.config.TagMatch != "" {
			// An explicit filter is a user decision: no match means first release.
			return "", err
		}
		// The derived (tagName) filter matched nothing — a format transition
		// or a v-prefixed repo with the default template. Keep version
		// continuity, exactly like the describe path below.
		return g.highestRawTag()
	}

	out, err := g.runSilent("describe", "--tags", "--abbrev=0")
	if err != nil {
		return "", fmt.Errorf("no git tags found: %w", err)
	}

	tag := strings.TrimSpace(out)

	// If tag matches the current format, return it directly
	if g.matchesEffectiveFilter(tag) {
		return tag, nil
	}

	// Tag doesn't match current format (e.g., found "v1.0.0" but tagName is "${version}").
	// Search for a matching tag first, then fall back to any tag for version continuity.
	g.logger.Debug("tag %q does not match tagName format, searching for matching tag", tag)
	matchedTag, matchErr := g.getLatestTagFromAllRefs()
	if matchErr == nil {
		return matchedTag, nil
	}

	if g.config.TagMatch != "" {
		// Never hand another package's tag to an explicitly filtered release.
		return "", fmt.Errorf("no git tags match tagMatch %q", g.config.TagMatch)
	}

	// No matching tag found — this is a format transition scenario.
	// Return the original tag so version number is preserved.
	g.logger.Debug("no matching tags found for current format, using %q for version continuity", tag)
	return tag, nil
}

// highestRawTag returns the highest tag by semver comparison, ignoring the
// tag filters. Used only as the version-continuity fallback when a derived
// filter matches nothing.
func (g *Git) highestRawTag() (string, error) {
	out, err := g.runSilent("tag", "-l", "--sort=-v:refname")
	if err != nil {
		return "", fmt.Errorf("listing git tags: %w", err)
	}

	var bestTag string
	var bestVer *semver.Version
	first := ""
	for _, tag := range strings.Split(strings.TrimSpace(out), "\n") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if first == "" {
			first = tag
		}
		ver, parseErr := version.ParseVersion(tag)
		if parseErr != nil {
			continue
		}
		if bestVer == nil || ver.GreaterThan(bestVer) {
			bestVer, bestTag = ver, tag
		}
	}
	if bestTag != "" {
		return bestTag, nil
	}
	if first != "" {
		return first, nil
	}
	return "", fmt.Errorf("no git tags found")
}

// VersionFromTag strips the literal text around ${version} in the tagName
// template from a tag, e.g. "release-1.2.3" with "release-${version}" →
// "1.2.3". Tags the template does not apply to are returned unchanged
// (ParseVersion tolerates a bare "v" prefix on its own).
func VersionFromTag(tag string, template string) string {
	idx := strings.Index(template, "${version}")
	if idx < 0 {
		return tag
	}
	prefix := template[:idx]
	suffix := template[idx+len("${version}"):]
	if len(tag) < len(prefix)+len(suffix) || !strings.HasPrefix(tag, prefix) || !strings.HasSuffix(tag, suffix) {
		return tag
	}
	return tag[len(prefix) : len(tag)-len(suffix)]
}

// getLatestTagFromAllRefs lists all matching tags and returns the highest by
// semver comparison. Git's -v:refname sort is NOT semver: without a
// versionsort.suffix config it ranks 1.0.0-rc.1 above 1.0.0, so trusting the
// output order returned an older pre-release as "latest".
func (g *Git) getLatestTagFromAllRefs() (string, error) {
	out, err := g.runSilent("tag", "-l", "--sort=-v:refname")
	if err != nil {
		return "", fmt.Errorf("listing git tags: %w", err)
	}

	var bestTag string
	var bestVer *semver.Version
	firstMatch := ""
	for _, tag := range strings.Split(strings.TrimSpace(out), "\n") {
		tag = strings.TrimSpace(tag)
		if tag == "" || !g.matchesEffectiveFilter(tag) {
			continue
		}
		if firstMatch == "" {
			firstMatch = tag
		}
		ver, parseErr := version.ParseVersion(VersionFromTag(tag, g.config.TagName))
		if parseErr != nil {
			continue
		}
		if bestVer == nil || ver.GreaterThan(bestVer) {
			bestVer = ver
			bestTag = tag
		}
	}

	if bestTag != "" {
		return bestTag, nil
	}
	if firstMatch != "" {
		// Matching tags exist but none parse as semver — keep git's order.
		return firstMatch, nil
	}
	return "", fmt.Errorf("no matching git tags found")
}

// TagExists checks if a tag with the given name exists.
func (g *Git) TagExists(tagName string) (bool, error) {
	// In dry-run mode, we still need to check if the tag exists
	out, err := commandExecutor("git", "tag", "-l", tagName)
	if err != nil {
		return false, fmt.Errorf("checking tag %s: %w", tagName, err)
	}
	return strings.TrimSpace(out) == tagName, nil
}

// TagPointsAtHead reports whether the given tag references the current HEAD
// commit. Read-only, so it runs the real command even in dry-run mode.
func (g *Git) TagPointsAtHead(tagName string) (bool, error) {
	out, err := commandExecutor("git", "tag", "-l", tagName, "--points-at", "HEAD")
	if err != nil {
		return false, fmt.Errorf("checking tag %s against HEAD: %w", tagName, err)
	}
	return strings.TrimSpace(out) == tagName, nil
}

// GetLatestPreReleaseTagMerged returns the latest pre-release tag merged into HEAD
// that matches the given preReleaseID. This ensures only tags reachable from the
// current branch are considered, preventing cross-branch tag pollution.
// Returns ("", nil) if no matching tag is found.
func (g *Git) GetLatestPreReleaseTagMerged(preReleaseID string) (string, error) {
	if preReleaseID == "" {
		return "", nil
	}

	out, err := g.runSilent("tag", "-l", "--merged", "HEAD", "--sort=-v:refname")
	if err != nil {
		return "", fmt.Errorf("listing merged tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(out), "\n")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		if !g.matchesEffectiveFilter(tag) {
			continue
		}

		// Match exact preReleaseID on the PARSED pre-release component: the
		// template's own hyphens (release-${version}) must not be mistaken
		// for the semver separator. "beta." prevents "beta" matching "betafix".
		parsed, parseErr := version.ParseVersion(VersionFromTag(tag, g.config.TagName))
		if parseErr != nil {
			continue
		}
		if strings.HasPrefix(parsed.Prerelease(), preReleaseID+".") {
			return tag, nil
		}
	}

	return "", nil
}

// GetLatestStableTagMerged returns the latest stable (non-pre-release) tag merged
// into HEAD. A stable tag is one whose parsed version has no pre-release component.
// Returns ("", nil) if no stable tag is found.
func (g *Git) GetLatestStableTagMerged() (string, error) {
	out, err := g.runSilent("tag", "-l", "--merged", "HEAD", "--sort=-v:refname")
	if err != nil {
		return "", fmt.Errorf("listing merged tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(out), "\n")
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}

		if !g.matchesEffectiveFilter(tag) {
			continue
		}

		parsed, parseErr := version.ParseVersion(VersionFromTag(tag, g.config.TagName))
		if parseErr != nil {
			continue
		}
		if parsed.Prerelease() == "" {
			return tag, nil
		}
	}

	return "", nil
}

// matchesTagNameFormat checks if a tag matches the expected format derived from tagName.
// For "${version}" (bare format), tags must start with a digit.
// For other templates, uses glob matching.
func matchesTagNameFormat(tagName, tag string) bool {
	if tagName == "" {
		return true
	}
	if tagName == "${version}" {
		// Bare version: must start with a digit (rejects "v1.0.0", "release-1.0.0")
		return len(tag) > 0 && tag[0] >= '0' && tag[0] <= '9'
	}
	pattern := strings.ReplaceAll(tagName, "${version}", "*")
	return matchGlob(pattern, tag)
}

// matchesEffectiveFilter checks if a tag matches the effective tag filters.
// If the user has explicitly set TagMatch, that takes priority.
// Otherwise, TagName template is used to derive the expected format.
func (g *Git) matchesEffectiveFilter(tag string) bool {
	if g.config.TagMatch != "" {
		// Explicit TagMatch takes priority
		if !matchGlob(g.config.TagMatch, tag) {
			return false
		}
	} else {
		// Derive format from TagName template
		if !matchesTagNameFormat(g.config.TagName, tag) {
			return false
		}
	}
	if g.config.TagExclude != "" && matchGlob(g.config.TagExclude, tag) {
		return false
	}
	return true
}

// matchGlob performs glob pattern matching (*, ?, character classes) via
// path.Match. The previous hand-rolled version only understood leading and
// trailing *, so npm-style tagMatch patterns like "[0-9]*" or mid-pattern
// wildcards silently matched nothing.
func matchGlob(pattern, s string) bool {
	if pattern == "*" {
		// Bare * keeps its legacy meaning: match everything, including tags
		// containing slashes (path.Match's * stops at separators).
		return true
	}
	ok, err := path.Match(pattern, s)
	if err != nil {
		// Invalid pattern — fall back to exact comparison
		return s == pattern
	}
	return ok
}
