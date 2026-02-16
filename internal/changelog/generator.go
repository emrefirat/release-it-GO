package changelog

import (
	"fmt"
	"os"
	"strings"

	"github.com/emfi/release-it-go/internal/git"
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
	AddUnreleased  bool          // add [Unreleased] section
	KeepUnreleased bool          // keep existing [Unreleased] section
	AddVersionURL  bool          // add compare URLs
	RepoInfo       *git.RepoInfo // repository info for URLs
}

// GenerateChangelog creates changelog content from parsed commits.
// Returns only the new version section (does not write to file).
func GenerateChangelog(commits []*Commit, version string, prevVersion string, opts Options) string {
	if opts.KeepAChangelog {
		return RenderKeepAChangelog(commits, version, "")
	}
	return RenderConventional(commits, version, prevVersion, opts.RepoInfo)
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

// insertAfterHeader inserts new content after the header in existing text.
func insertAfterHeader(existing string, header string, newSection string) string {
	idx := strings.Index(existing, header)
	if idx == -1 {
		// Header not found, prepend header + new content
		return header + "\n\n" + newSection + "\n" + existing
	}

	headerEnd := idx + len(header)

	// Find the end of the header line (skip any trailing newlines)
	rest := existing[headerEnd:]

	return existing[:headerEnd] + "\n\n" + newSection + "\n" + strings.TrimLeft(rest, "\n")
}
