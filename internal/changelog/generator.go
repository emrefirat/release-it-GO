package changelog

import (
	"fmt"
	"os"
	"strings"

	"github.com/emrefirat/release-it-GO/internal/git"
)

// DefaultHeader is the default changelog file header.
const DefaultHeader = "# Changelog"

// KeepAChangelogHeader is the full keep-a-changelog header.
const KeepAChangelogHeader = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).`

// Options holds configuration for changelog generation.
type Options struct {
	Preset         string        // "angular" (default)
	Infile         string        // "CHANGELOG.md"
	Header         string        // "# Changelog"
	KeepAChangelog bool          // false = conventional format
	AddVersionURL  bool          // add compare URLs to version headings
	RepoInfo       *git.RepoInfo // repository info for URLs
	TagName        string        // git tag of the new version (compare URL ref)
	PrevTagName    string        // git tag of the previous version (compare URL ref)
}

// GenerateChangelog creates changelog content from parsed commits.
// Returns only the new version section (does not write to file).
func GenerateChangelog(commits []*Commit, version string, prevVersion string, opts Options) string {
	if opts.KeepAChangelog {
		return RenderKeepAChangelog(commits, version, "")
	}
	compareRepo := opts.RepoInfo
	if !opts.AddVersionURL {
		compareRepo = nil // heading compare link off; commit hash links stay
	}
	return renderConventional(commits, version, prevVersion, opts.TagName, opts.PrevTagName, compareRepo, opts.RepoInfo)
}

// UpdateChangelogFile prepends new changelog content to the existing file.
// Creates the file if it does not exist.
func UpdateChangelogFile(filePath string, newContent string, header string) error {
	if header == "" {
		header = DefaultHeader
	}

	existing, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading changelog file: %w", err)
		}
		// File doesn't exist, create new
		content := header + "\n\n" + newContent
		return os.WriteFile(filePath, []byte(content), 0644)
	}

	updated := insertAfterHeader(string(existing), header, newContent)
	return os.WriteFile(filePath, []byte(updated), 0644)
}

// unreleasedHeading is the keep-a-changelog section that always stays on top.
const unreleasedHeading = "## [Unreleased]"

// insertAfterHeader inserts the new section after the header in the existing
// text. The header must sit at the start of a line (so "# Changelog" does not
// match inside "## Changelog"). A keep-a-changelog "## [Unreleased]" block
// stays first and the new section goes after it. Line endings are preserved:
// CRLF files used to end up with mixed endings and extra blank lines.
func insertAfterHeader(existing string, header string, newSection string) string {
	crlf := strings.Contains(existing, "\r\n")
	text := strings.ReplaceAll(existing, "\r\n", "\n")
	header = strings.ReplaceAll(header, "\r\n", "\n")

	var result string
	idx := indexHeaderLine(text, header)
	if idx == -1 {
		// Header not found, prepend header + new content
		result = header + "\n\n" + newSection + "\n" + text
	} else {
		headerEnd := idx + len(header)
		rest := text[headerEnd:]
		if blockEnd := unreleasedBlockEnd(rest); blockEnd >= 0 {
			unreleased := strings.TrimRight(rest[:blockEnd], "\n")
			result = text[:headerEnd] + unreleased + "\n\n" + newSection + "\n" + strings.TrimLeft(rest[blockEnd:], "\n")
		} else {
			result = text[:headerEnd] + "\n\n" + newSection + "\n" + strings.TrimLeft(rest, "\n")
		}
	}

	if crlf {
		result = strings.ReplaceAll(result, "\n", "\r\n")
	}
	return result
}

// indexHeaderLine finds header at a line boundary: preceded by the start of
// the text or a newline, and followed by a newline or the end of the text.
func indexHeaderLine(text string, header string) int {
	offset := 0
	for {
		idx := strings.Index(text[offset:], header)
		if idx == -1 {
			return -1
		}
		idx += offset
		end := idx + len(header)
		startOK := idx == 0 || text[idx-1] == '\n'
		endOK := end == len(text) || text[end] == '\n'
		if startOK && endOK {
			return idx
		}
		offset = idx + 1
	}
}

// unreleasedBlockEnd returns the index in rest where the "## [Unreleased]"
// block ends (the start of the next "## " heading, or len(rest)), or -1 when
// the block is absent or not the first heading.
func unreleasedBlockEnd(rest string) int {
	idx := indexHeaderLine(rest, unreleasedHeading)
	if idx == -1 {
		return -1
	}
	// Only when it is the first section after the header.
	if first := strings.Index(rest, "\n## "); first != -1 && first+1 < idx {
		return -1
	}
	next := strings.Index(rest[idx+len(unreleasedHeading):], "\n## ")
	if next == -1 {
		return len(rest)
	}
	return idx + len(unreleasedHeading) + next + 1
}
