package config

import (
	"fmt"
	"os"
)

const (
	// LegacyConfigFile is the npm release-it config file name.
	LegacyConfigFile = ".release-it.json"
	// NativeConfigFile is the release-it-go native config file name.
	NativeConfigFile = ".release-it-go.json"
)

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

	// Write as native config
	if err := WriteConfigJSON(cfg, NativeConfigFile); err != nil {
		return fmt.Errorf("writing native config: %w", err)
	}

	return nil
}
