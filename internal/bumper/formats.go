package bumper

import (
	"path/filepath"
	"strings"

	"github.com/emrefirat/release-it-GO/internal/config"
)

// Supported file format identifiers.
const (
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatTOML = "toml"
	FormatINI  = "ini"
	FormatText = "text"
)

// defaultStructuredPath is the key used when a JSON/YAML/TOML target names
// no path — npm's @release-it/bumper default, so `in: {file: package.json}`
// reads the version instead of stringifying the whole document.
const defaultStructuredPath = "version"

// effectivePath returns the value path for the file, defaulting to "version"
// for structured formats. INI and text have no key-path default.
func effectivePath(file config.BumperFile, format string) string {
	if file.Path != "" {
		return file.Path
	}
	switch format {
	case FormatJSON, FormatYAML, FormatTOML:
		return defaultStructuredPath
	}
	return file.Path
}

// detectFormat determines the file format from config or file extension.
func detectFormat(file config.BumperFile) string {
	if file.Type != "" {
		return strings.ToLower(file.Type)
	}

	ext := strings.ToLower(filepath.Ext(file.File))
	switch ext {
	case ".json":
		return FormatJSON
	case ".yaml", ".yml":
		return FormatYAML
	case ".toml":
		return FormatTOML
	case ".ini", ".cfg":
		return FormatINI
	default:
		return FormatText
	}
}
