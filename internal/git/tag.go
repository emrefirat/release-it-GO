package git

import (
	"fmt"
	"strings"

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
		return g.getLatestTagFromAllRefs()
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

	// No matching tag found — this is a format transition scenario.
	// Return the original tag so version number is preserved.
	g.logger.Debug("no matching tags found for current format, using %q for version continuity", tag)
	return tag, nil
}

// getLatestTagFromAllRefs lists all tags sorted by version and returns the first.
func (g *Git) getLatestTagFromAllRefs() (string, error) {
	out, err := g.runSilent("tag", "-l", "--sort=-v:refname")
	if err != nil {
		return "", fmt.Errorf("listing git tags: %w", err)
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

		return tag, nil
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

		// Match exact preReleaseID: find first "-" (semver pre-release separator),
		// then check that the pre-release section starts with "preReleaseID."
		// This prevents "beta" from matching "betafix" tags.
		if idx := strings.Index(tag, "-"); idx >= 0 {
			preReleasePart := tag[idx+1:]
			if strings.HasPrefix(preReleasePart, preReleaseID+".") {
				return tag, nil
			}
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

		parsed, parseErr := version.ParseVersion(tag)
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

// matchGlob performs simple glob-like pattern matching.
func matchGlob(pattern, s string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(s, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(s, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(s, pattern[:len(pattern)-1])
	}
	return s == pattern
}
