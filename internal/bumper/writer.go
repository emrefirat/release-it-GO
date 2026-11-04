package bumper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/emrefirat/release-it-GO/internal/config"
	toml "github.com/pelletier/go-toml/v2"
	yaml "go.yaml.in/yaml/v3"
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
	path := effectivePath(file, format)

	var updated []byte
	switch format {
	case FormatJSON:
		updated, err = writeJSON(data, path, finalVersion)
	case FormatYAML:
		updated, err = writeYAML(data, path, finalVersion)
	case FormatTOML:
		updated, err = writeTOML(data, path, finalVersion)
	case FormatINI:
		updated, err = writeINI(data, path, finalVersion)
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

	// Fallback re-marshal: no HTML escaping ("&" must stay "&", not
	// "\u0026" — package.json descriptions and URLs are prose, not HTML).
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(obj); err != nil {
		return nil, fmt.Errorf("encoding JSON: %w", err)
	}

	return buf.Bytes(), nil // Encode appends the trailing newline
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
//
// Only the value span is replaced: the key's spacing, an inline comment
// (`version = 1.0.0 ; current`) and the line's CR are kept, so setup.cfg-style
// files do not get rewritten around the version.
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

		if !inSection {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 || strings.TrimSpace(line[:eq]) != key {
			continue
		}

		cr := ""
		body := line
		if strings.HasSuffix(body, "\r") {
			cr = "\r"
			body = body[:len(body)-1]
		}
		leading, _, trailing := splitINIValue(body[eq+1:])
		lines[i] = body[:eq+1] + leading + version + trailing + cr
		found = true
		break
	}

	if !found {
		return nil, fmt.Errorf("key %q not found in INI", path)
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// splitINIValue splits the text after "=" into the leading whitespace, the
// value and the trailing part (whitespace + inline comment, if any). An
// inline comment starts at ";" or "#" preceded by whitespace.
func splitINIValue(rest string) (leading string, value string, trailing string) {
	valueStart := len(rest) - len(strings.TrimLeft(rest, " \t"))
	leading = rest[:valueStart]
	content := rest[valueStart:]

	commentStart := len(content)
	for i := 1; i < len(content); i++ {
		if (content[i] == ';' || content[i] == '#') && (content[i-1] == ' ' || content[i-1] == '\t') {
			commentStart = i
			break
		}
	}
	valueWithSpace := content[:commentStart]
	value = strings.TrimRight(valueWithSpace, " \t")
	trailing = valueWithSpace[len(value):] + content[commentStart:]
	return leading, value, trailing
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
//
// Only whole-version occurrences are replaced: "1.0.0" must not turn
// "v11.0.0" into "v12.0.0" or "1.0.0.1" into "2.0.0.1" (compat matrices and
// changelogs list other versions). A match may be preceded by anything but a
// digit or a dot (so "v1.0.0" and "^1.0.0" count) and followed by anything
// but a digit, letter or dot.
func replaceVersionInText(data []byte, from string, to string, name string) ([]byte, error) {
	if from == "" {
		return nil, fmt.Errorf("%s: current version is unknown, cannot replace it in a text file (set consumeWholeFile: true to overwrite the whole file, or configure bumper.in)", name)
	}

	text := string(data)
	var b strings.Builder
	b.Grow(len(text))
	replaced := 0
	for i := 0; i < len(text); {
		idx := strings.Index(text[i:], from)
		if idx < 0 {
			b.WriteString(text[i:])
			break
		}
		start := i + idx
		end := start + len(from)
		b.WriteString(text[i:start])
		if isWholeVersionMatch(text, start, end) {
			b.WriteString(to)
			replaced++
		} else {
			b.WriteString(from)
		}
		i = end
	}

	if replaced == 0 {
		return nil, fmt.Errorf("%s: current version %q not found in file (set consumeWholeFile: true to overwrite the whole file)", name, from)
	}
	return []byte(b.String()), nil
}

// isWholeVersionMatch reports whether text[start:end] is a standalone version
// token rather than part of a longer version-like string.
func isWholeVersionMatch(text string, start int, end int) bool {
	if start > 0 {
		c := text[start-1]
		if isDigit(c) || c == '.' {
			return false
		}
	}
	if end < len(text) {
		c := text[end]
		if isDigit(c) || isLetter(c) || c == '.' {
			return false
		}
	}
	return true
}

func isDigit(c byte) bool  { return c >= '0' && c <= '9' }
func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
