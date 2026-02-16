package bumper

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/emfi/release-it-go/internal/config"
	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
)

// ReadVersionFromFile reads a version string from the specified file.
func ReadVersionFromFile(file config.BumperFile) (string, error) {
	data, err := os.ReadFile(file.File)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", file.File, err)
	}

	format := detectFormat(file)

	var version string
	switch format {
	case FormatJSON:
		version, err = readJSON(data, file.Path)
	case FormatYAML:
		version, err = readYAML(data, file.Path)
	case FormatTOML:
		version, err = readTOML(data, file.Path)
	case FormatINI:
		version, err = readINI(data, file.Path)
	case FormatText:
		version, err = readText(data)
	default:
		return "", fmt.Errorf("unsupported format for %s", file.File)
	}

	if err != nil {
		return "", fmt.Errorf("reading version from %s: %w", file.File, err)
	}

	return version, nil
}

// readJSON reads a version from JSON data using a dot-separated path.
func readJSON(data []byte, path string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("parsing JSON: %w", err)
	}

	return extractNestedValue(obj, path)
}

// readYAML reads a version from YAML data using a dot-separated path.
func readYAML(data []byte, path string) (string, error) {
	var obj interface{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("parsing YAML: %w", err)
	}

	return extractNestedValue(obj, path)
}

// readTOML reads a version from TOML data using a dot-separated path.
func readTOML(data []byte, path string) (string, error) {
	var obj map[string]interface{}
	if err := toml.Unmarshal(data, &obj); err != nil {
		return "", fmt.Errorf("parsing TOML: %w", err)
	}

	return extractNestedValue(obj, path)
}

// readINI reads a version from INI data using [section].key format.
func readINI(data []byte, path string) (string, error) {
	section, key := parseINIPath(path)
	lines := strings.Split(string(data), "\n")

	inSection := section == ""
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			inSection = sectionName == section
			continue
		}

		if inSection {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.TrimSpace(parts[1])
				if k == key {
					return v, nil
				}
			}
		}
	}

	return "", fmt.Errorf("key %q not found in INI", path)
}

// readText reads the entire file content as a version string.
func readText(data []byte) (string, error) {
	return strings.TrimSpace(string(data)), nil
}

// extractNestedValue extracts a value from a nested map using dot-separated path.
func extractNestedValue(obj interface{}, path string) (string, error) {
	if path == "" {
		return fmt.Sprintf("%v", obj), nil
	}

	keys := strings.Split(path, ".")
	current := obj

	for _, key := range keys {
		switch m := current.(type) {
		case map[string]interface{}:
			val, ok := m[key]
			if !ok {
				return "", fmt.Errorf("key %q not found at path %q", key, path)
			}
			current = val
		case map[interface{}]interface{}:
			// YAML may produce this type
			val, ok := m[key]
			if !ok {
				return "", fmt.Errorf("key %q not found at path %q", key, path)
			}
			current = val
		default:
			return "", fmt.Errorf("cannot traverse into %T at key %q", current, key)
		}
	}

	return fmt.Sprintf("%v", current), nil
}

// parseINIPath splits an INI path like "[section].key" into section and key.
func parseINIPath(path string) (string, string) {
	if strings.HasPrefix(path, "[") {
		end := strings.Index(path, "]")
		if end > 0 {
			section := path[1:end]
			key := strings.TrimPrefix(path[end+1:], ".")
			return section, key
		}
	}
	return "", path
}
