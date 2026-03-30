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

// CommitInfo holds a commit's short hash and subject line.
type CommitInfo struct {
	Hash    string // short hash (8 char)
	Subject string // first line of commit message
}

// GetCommitsWithHashSinceTag returns commits with hash and subject since the given tag.
// If tag is empty, returns all commits.
func (g *Git) GetCommitsWithHashSinceTag(tag string) ([]CommitInfo, error) {
	var args []string
	if tag == "" {
		args = []string{"log", "--pretty=format:%h||%s"}
	} else {
		args = []string{"log", tag + "..HEAD", "--pretty=format:%h||%s"}
	}

	out, err := g.runSilent(args...)
	if err != nil {
		return nil, fmt.Errorf("getting commits with hash since %s: %w", tag, err)
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, nil
	}

	lines := strings.Split(trimmed, "\n")
	commits := make([]CommitInfo, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "||", 2)
		if len(parts) == 2 {
			commits = append(commits, CommitInfo{
				Hash:    parts[0],
				Subject: parts[1],
			})
		}
	}

	return commits, nil
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

// GetCommitCountSinceTag returns the number of commits since the given tag.
func (g *Git) GetCommitCountSinceTag(tag string) (int, error) {
	var args []string
	if tag == "" {
		args = []string{"rev-list", "--count", "HEAD"}
	} else {
		args = []string{"rev-list", "--count", tag + "..HEAD"}
	}

	out, err := g.runSilent(args...)
	if err != nil {
		return 0, err
	}

	trimmed := strings.TrimSpace(out)
	count := 0
	for _, ch := range trimmed {
		if ch >= '0' && ch <= '9' {
			count = count*10 + int(ch-'0')
		}
	}
	return count, nil
}

// GetContributorsSinceTag returns unique contributor names since the given tag.
func (g *Git) GetContributorsSinceTag(tag string) ([]string, error) {
	var args []string
	if tag == "" {
		args = []string{"log", "--pretty=format:%cn", "HEAD"}
	} else {
		args = []string{"log", "--pretty=format:%cn", tag + "..HEAD"}
	}

	out, err := g.runSilent(args...)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return nil, nil
	}

	seen := make(map[string]bool)
	var unique []string
	for _, name := range strings.Split(trimmed, "\n") {
		name = strings.TrimSpace(name)
		if name != "" && !seen[name] {
			seen[name] = true
			unique = append(unique, name)
		}
	}
	return unique, nil
}
