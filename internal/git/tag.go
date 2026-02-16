package git

import (
	"fmt"
	"strings"
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

// GetLatestTag returns the most recent tag.
func (g *Git) GetLatestTag() (string, error) {
	if g.config.GetLatestTagFromAllRefs {
		return g.getLatestTagFromAllRefs()
	}

	out, err := g.runSilent("describe", "--tags", "--abbrev=0")
	if err != nil {
		return "", fmt.Errorf("no git tags found")
	}
	return strings.TrimSpace(out), nil
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

		if g.config.TagMatch != "" && !matchGlob(g.config.TagMatch, tag) {
			continue
		}
		if g.config.TagExclude != "" && matchGlob(g.config.TagExclude, tag) {
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
