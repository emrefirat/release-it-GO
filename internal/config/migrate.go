package config

import (
	"fmt"
	"os"
)

const (
	// LegacyConfigFile is the npm release-it config file name.
	LegacyConfigFile = ".release-it.json"
	// NativeConfigFile is the release-it-go native config file name (JSON, default).
	NativeConfigFile = ".release-it-go.json"
	// NativeConfigFileYAML is the release-it-go native config file name (YAML).
	NativeConfigFileYAML = ".release-it-go.yaml"
)

// NativeConfigFileForFormat returns the native config filename for the given format.
func NativeConfigFileForFormat(format string) string {
	switch format {
	case "yaml":
		return NativeConfigFileYAML
	default:
		return NativeConfigFile
	}
}

// DetectNativeConfigAny checks if any native config file (.json or .yaml) exists.
// Returns the path and true if found. Checks JSON first, then YAML.
func DetectNativeConfigAny() (string, bool) {
	if fileExists(NativeConfigFile) {
		return NativeConfigFile, true
	}
	if fileExists(NativeConfigFileYAML) {
		return NativeConfigFileYAML, true
	}
	return "", false
}

// DetectLegacyConfig checks if a legacy .release-it.json file exists
// in the current directory. Returns the path and true if found.
func DetectLegacyConfig() (string, bool) {
	if fileExists(LegacyConfigFile) {
		return LegacyConfigFile, true
	}
	return "", false
}

// DetectNativeConfig checks if a native .release-it-go.json file exists
// in the current directory.
func DetectNativeConfig() bool {
	return fileExists(NativeConfigFile)
}

// MigrateLegacyConfig reads a legacy .release-it.json file, creates a backup,
// normalizes the content, applies plugin compatibility mappings, and writes
// the result as .release-it-go.json.
func MigrateLegacyConfig(path string) error {
	// Read the legacy config
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading legacy config %s: %w", path, err)
	}

	// Create backup
	backupPath := path + ".bak"
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("creating backup %s: %w", backupPath, err)
	}

	// Load config through the normal pipeline (normalizeJSON + applyPluginCompat)
	cfg, err := loadFromFile(DefaultConfig(), path)
	if err != nil {
		return fmt.Errorf("parsing legacy config %s: %w", path, err)
	}

	// Write as native config (JSON)
	if err := WriteConfigJSON(cfg, NativeConfigFile); err != nil {
		return fmt.Errorf("writing native config: %w", err)
	}

	return nil
}

// MigrateLegacyConfigTo reads a legacy config, creates a backup, and writes
// the result in the specified format ("json" or "yaml").
func MigrateLegacyConfigTo(path string, format string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading legacy config %s: %w", path, err)
	}

	backupPath := path + ".bak"
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("creating backup %s: %w", backupPath, err)
	}

	cfg, err := loadFromFile(DefaultConfig(), path)
	if err != nil {
		return fmt.Errorf("parsing legacy config %s: %w", path, err)
	}

	outFile := NativeConfigFileForFormat(format)
	switch format {
	case "yaml":
		if err := WriteConfigYAML(cfg, outFile); err != nil {
			return fmt.Errorf("writing native config: %w", err)
		}
	default:
		if err := WriteConfigJSON(cfg, outFile); err != nil {
			return fmt.Errorf("writing native config: %w", err)
		}
	}

	return nil
}
