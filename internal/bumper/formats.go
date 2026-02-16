package bumper

import (
	"path/filepath"
	"strings"

	"release-it-go/internal/config"
)

// Supported file format identifiers.
const (
	FormatJSON = "json"
	FormatYAML = "yaml"
	FormatTOML = "toml"
	FormatINI  = "ini"
	FormatText = "text"
)

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
