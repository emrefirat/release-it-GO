package bumper

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"release-it-go/internal/config"
	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
)

// WriteVersionToFile updates the version in the specified file.
func WriteVersionToFile(file config.BumperFile, version string) error {
	finalVersion := file.Prefix + version

	if file.ConsumeWholeFile {
		return os.WriteFile(file.File, []byte(finalVersion+"\n"), 0644)
	}

	data, err := os.ReadFile(file.File)
	if err != nil {
		return fmt.Errorf("reading %s: %w", file.File, err)
	}

	format := detectFormat(file)

	var updated []byte
	switch format {
	case FormatJSON:
		updated, err = writeJSON(data, file.Path, finalVersion)
	case FormatYAML:
		updated, err = writeYAML(data, file.Path, finalVersion)
	case FormatTOML:
		updated, err = writeTOML(data, file.Path, finalVersion)
	case FormatINI:
		updated, err = writeINI(data, file.Path, finalVersion)
	case FormatText:
		updated = []byte(finalVersion + "\n")
		err = nil
	default:
		return fmt.Errorf("unsupported format for %s", file.File)
	}

	if err != nil {
		return fmt.Errorf("updating version in %s: %w", file.File, err)
	}

	return os.WriteFile(file.File, updated, 0644)
}

// writeJSON updates a version in JSON data at the given dot-separated path.
func writeJSON(data []byte, path string, version string) ([]byte, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	if err := setNestedValue(obj, path, version); err != nil {
		return nil, err
	}

	result, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encoding JSON: %w", err)
	}

	return append(result, '\n'), nil
}

// writeYAML updates a version in YAML data at the given dot-separated path.
func writeYAML(data []byte, path string, version string) ([]byte, error) {
	var obj map[string]interface{}
	if err := yaml.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if err := setNestedValue(obj, path, version); err != nil {
		return nil, err
	}

	result, err := yaml.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("encoding YAML: %w", err)
	}

	return result, nil
}

// writeTOML updates a version in TOML data at the given dot-separated path.
func writeTOML(data []byte, path string, version string) ([]byte, error) {
	var obj map[string]interface{}
	if err := toml.Unmarshal(data, &obj); err != nil {
		return nil, fmt.Errorf("parsing TOML: %w", err)
	}

	if err := setNestedValue(obj, path, version); err != nil {
		return nil, err
	}

	result, err := toml.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("encoding TOML: %w", err)
	}

	return result, nil
}

// writeINI updates a version in INI data at the given [section].key path.
func writeINI(data []byte, path string, version string) ([]byte, error) {
	section, key := parseINIPath(path)
	lines := strings.Split(string(data), "\n")

	inSection := section == ""
	found := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, ";") {
			continue
		}

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			sectionName := strings.TrimSpace(trimmed[1 : len(trimmed)-1])
			inSection = sectionName == section
			continue
		}

		if inSection {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				if k == key {
					lines[i] = fmt.Sprintf("%s = %s", key, version)
					found = true
					break
				}
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("key %q not found in INI", path)
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// setNestedValue sets a value in a nested map using a dot-separated path.
func setNestedValue(obj map[string]interface{}, path string, value string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	keys := strings.Split(path, ".")
	current := interface{}(obj)

	for i, key := range keys {
		if i == len(keys)-1 {
			// Set the final value
			switch m := current.(type) {
			case map[string]interface{}:
				m[key] = value
				return nil
			default:
				return fmt.Errorf("cannot set value at path %q: parent is %T", path, current)
			}
		}

		// Traverse deeper
		switch m := current.(type) {
		case map[string]interface{}:
			next, ok := m[key]
			if !ok {
				return fmt.Errorf("key %q not found at path %q", key, path)
			}
			current = next
		default:
			return fmt.Errorf("cannot traverse into %T at key %q", current, key)
		}
	}

	return nil
}
