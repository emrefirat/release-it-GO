package git

import (
	"fmt"
	"strings"
)

// GenerateChangelog generates a changelog from git log between two refs.
// If fromTag is empty, all commits up to toRef are included.
func (g *Git) GenerateChangelog(fromTag string, toRef string) (string, error) {
	if toRef == "" {
		toRef = "HEAD"
	}

	var rangeArg string
	if fromTag == "" {
		rangeArg = toRef
	} else {
		rangeArg = fromTag + ".." + toRef
	}

	format := g.config.Changelog
	if format == "" {
		format = "* %s (%h)"
	}

	out, err := g.runSilent("log", rangeArg, "--pretty=format:"+format)
	if err != nil {
		return "", fmt.Errorf("generating changelog: %w", err)
	}

	return strings.TrimSpace(out), nil
}

// GetCommitsSinceTag returns commit subject lines since the given tag.
// If tag is empty, returns all commits.
func (g *Git) GetCommitsSinceTag(tag string) ([]string, error) {
	var args []string
	if tag == "" {
		args = []string{"log", "--pretty=format:%s"}
	} else {
		args = []string{"log", tag + "..HEAD", "--pretty=format:%s"}
	}

	out, err := g.runSilent(args...)
	if err != nil {
		return nil, fmt.Errorf("getting commits since %s: %w", tag, err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, nil
	}

	lines := strings.Split(trimmed, "\n")
	commits := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			commits = append(commits, line)
		}
	}

	return commits, nil
}
