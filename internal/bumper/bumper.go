// Package bumper provides multi-file version reading and writing.
// It supports JSON, YAML, TOML, INI, and plain text formats,
// enabling version updates across multiple project files simultaneously.
package bumper

import (
	"fmt"
	"path/filepath"

	"release-it-go/internal/config"
	applog "release-it-go/internal/log"
)

// Bumper reads and writes version strings across multiple files.
type Bumper struct {
	config *config.BumperConfig
	logger *applog.Logger
	dryRun bool
}

// NewBumper creates a new Bumper instance.
func NewBumper(cfg *config.BumperConfig, logger *applog.Logger, dryRun bool) *Bumper {
	return &Bumper{
		config: cfg,
		logger: logger,
		dryRun: dryRun,
	}
}

// ReadVersion reads the current version from the configured input file.
// Returns empty string if no input file is configured.
func (b *Bumper) ReadVersion() (string, error) {
	if b.config == nil || b.config.In == nil {
		return "", nil
	}

	return ReadVersionFromFile(*b.config.In)
}

// WriteVersion writes the new version to all configured output files.
func (b *Bumper) WriteVersion(version string) error {
	if b.config == nil || len(b.config.Out) == 0 {
		return nil
	}

	for _, outFile := range b.config.Out {
		files, err := resolveGlob(outFile.File)
		if err != nil {
			return fmt.Errorf("resolving glob %q: %w", outFile.File, err)
		}

		for _, f := range files {
			fileCopy := outFile
			fileCopy.File = f

			if b.dryRun {
				b.logger.DryRun("Would update version in %s to %s", f, version)
				continue
			}

			b.logger.Verbose("Updating version in %s to %s", f, version)
			if err := WriteVersionToFile(fileCopy, version); err != nil {
				return err
			}
		}
	}

	return nil
}

// resolveGlob expands glob patterns in file paths.
// If the pattern contains no glob characters, returns the path as-is.
func resolveGlob(pattern string) ([]string, error) {
	if !containsGlobChar(pattern) {
		return []string{pattern}, nil
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern %q: %w", pattern, err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no files matched pattern %q", pattern)
	}

	return matches, nil
}

// containsGlobChar checks if a string contains glob metacharacters.
func containsGlobChar(s string) bool {
	for _, c := range s {
		if c == '*' || c == '?' || c == '[' {
			return true
		}
	}
	return false
}
