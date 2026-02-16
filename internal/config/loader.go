package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// configSearchFiles lists config file names in search priority order.
var configSearchFiles = []string{
	".release-it.json",
	".release-it.yaml",
	".release-it.yml",
	".release-it.toml",
}

// LoadConfig loads configuration from the given path or searches for a config file
// in the current directory. Returns defaults if no config file is found.
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath != "" {
		return loadFromFile(cfg, configPath)
	}

	for _, f := range configSearchFiles {
		if fileExists(f) {
			return loadFromFile(cfg, f)
		}
	}

	return cfg, nil
}

// loadFromFile reads and merges a config file into the given default config.
func loadFromFile(cfg *Config, path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	format := extToViperType(ext)

	// Keep original data for plugin compat processing
	originalData := data

	// Pre-process JSON to fix type mismatches from npm release-it format
	if format == "json" {
		data = normalizeJSON(data)
	}

	v := viper.New()
	v.SetConfigType(format)

	if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", path, err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config file %s: %w", path, err)
	}

	// Apply backward compatibility for release-it npm plugin settings
	applyPluginCompat(cfg, originalData, format)

	return cfg, nil
}

// LoadConfigFromBytes parses config from raw bytes with the specified format.
// Supported formats: "json", "yaml", "toml".
func LoadConfigFromBytes(data []byte, format string) (*Config, error) {
	cfg := DefaultConfig()

	if format == "json" {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing JSON config: %w", err)
		}
		return cfg, nil
	}

	v := viper.New()
	v.SetConfigType(format)

	if err := v.ReadConfig(strings.NewReader(string(data))); err != nil {
		return nil, fmt.Errorf("parsing %s config: %w", format, err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling %s config: %w", format, err)
	}

	return cfg, nil
}

// extToViperType converts a file extension to a viper config type.
func extToViperType(ext string) string {
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	default:
		return "json"
	}
}

// fileExists checks if a file exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
