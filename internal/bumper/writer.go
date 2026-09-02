package bumper

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
	"release-it-go/internal/config"
)

// WriteVersionToFile updates the version in the specified file. Text targets
// need the current version to edit in place; use WriteVersionToFileFrom.
func WriteVersionToFile(file config.BumperFile, version string) error {
	return WriteVersionToFileFrom(file, "", version)
}

// WriteVersionToFileFrom updates the version in the specified file, given the
// current version (from) and the new one (to). Structured formats locate the
// value by path; plain-text targets are edited in place by replacing every
// occurrence of the current version — they are NEVER truncated to the bare
// version unless ConsumeWholeFile is set explicitly.
func WriteVersionToFileFrom(file config.BumperFile, from string, to string) error {
	finalVersion := file.Prefix + to

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
		updated, err = replaceVersionInText(data, from, to, file.File)
	default:
		return fmt.Errorf("unsupported format for %s", file.File)
	}

	if err != nil {
		return fmt.Errorf("updating version in %s: %w", file.File, err)
	}

	return os.WriteFile(file.File, updated, 0644)
}

// updateVersionTargeted attempts a formatting-preserving update: it replaces
// only the old value at the path's final key in the raw bytes, then PROVES
// the edit correct by re-parsing and comparing the whole tree against the
// expected result. Ambiguous or unprovable edits return ok=false and the
// caller falls back to the full re-marshal (which loses comments, key order,
// and custom indentation — the very diff noise this path avoids).
func updateVersionTargeted(data []byte, path string, version string, parse func([]byte, interface{}) error) ([]byte, bool) {
	var current map[string]interface{}
	if err := parse(data, &current); err != nil {
		return nil, false
	}

	oldValue, isString := getNestedString(current, path)
	if !isString {
		return nil, false
	}
	if oldValue == version {
		return data, true // already up to date; keep the file byte-identical
	}

	var expected map[string]interface{}
	if err := parse(data, &expected); err != nil {
		return nil, false
	}
	if err := setNestedValue(expected, path, version); err != nil {
		return nil, false
	}

	keys := strings.Split(path, ".")
	lastKey := keys[len(keys)-1]
	// Matches `"key": "old`, `key: old`, `key = "old` across JSON/YAML/TOML.
	pattern := regexp.MustCompile(`("?` + regexp.QuoteMeta(lastKey) + `"?\s*[:=]\s*["']?)` + regexp.QuoteMeta(oldValue))

	for _, loc := range pattern.FindAllSubmatchIndex(data, -1) {
		prefixEnd := loc[3] // end of the key+separator group; old value follows
		candidate := make([]byte, 0, len(data)+len(version))
		candidate = append(candidate, data[:prefixEnd]...)
		candidate = append(candidate, version...)
		candidate = append(candidate, data[loc[1]:]...)

		var got map[string]interface{}
		if err := parse(candidate, &got); err != nil {
			continue
		}
		if reflect.DeepEqual(got, expected) {
			return candidate, true
		}
	}

	return nil, false
}

// getNestedString reads a string value at a dot-separated path.
func getNestedString(obj map[string]interface{}, path string) (string, bool) {
	current := interface{}(obj)
	for _, key := range strings.Split(path, ".") {
		m, ok := current.(map[string]interface{})
		if !ok {
			return "", false
		}
		current, ok = m[key]
		if !ok {
			return "", false
		}
	}
	s, ok := current.(string)
	return s, ok
}

// writeJSON updates a version in JSON data at the given dot-separated path.
func writeJSON(data []byte, path string, version string) ([]byte, error) {
	if updated, ok := updateVersionTargeted(data, path, version, jsonParse); ok {
		return updated, nil
	}

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
	if updated, ok := updateVersionTargeted(data, path, version, yamlParse); ok {
		return updated, nil
	}

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
	if updated, ok := updateVersionTargeted(data, path, version, tomlParse); ok {
		return updated, nil
	}

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

func jsonParse(data []byte, v interface{}) error { return json.Unmarshal(data, v) }
func yamlParse(data []byte, v interface{}) error { return yaml.Unmarshal(data, v) }
func tomlParse(data []byte, v interface{}) error { return toml.Unmarshal(data, v) }

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

// replaceVersionInText replaces every occurrence of the current version in a
// plain-text file. Any file whose extension is not a structured format lands
// here (README.md, install.sh, ...), so overwriting the whole file — the old
// behavior — silently destroyed documentation. Without a known current
// version, or when it does not occur in the file, refuse and point at the
// explicit consumeWholeFile opt-in instead of guessing.
func replaceVersionInText(data []byte, from string, to string, name string) ([]byte, error) {
	if from == "" {
		return nil, fmt.Errorf("%s: current version is unknown, cannot replace it in a text file (set consumeWholeFile: true to overwrite the whole file, or configure bumper.in)", name)
	}
	if !strings.Contains(string(data), from) {
		return nil, fmt.Errorf("%s: current version %q not found in file (set consumeWholeFile: true to overwrite the whole file)", name, from)
	}
	return []byte(strings.ReplaceAll(string(data), from, to)), nil
}
