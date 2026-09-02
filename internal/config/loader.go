package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// configSearchFiles lists config file names in search priority order.
// Native .release-it-go.* files take priority over legacy .release-it.* files.
var configSearchFiles = []string{
	".release-it-go.json",
	".release-it-go.yaml",
	".release-it-go.yml",
	".release-it-go.toml",
	".release-it.json",
	".release-it.yaml",
	".release-it.yml",
	".release-it.toml",
}

// LoadConfig loads configuration from the given path or searches for a config file
// in the current directory. Returns defaults if no config file is found.
// The loaded file path is stored in Config.ConfigFile (empty if using defaults).
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath != "" {
		loaded, err := loadFromFile(cfg, configPath)
		if err != nil {
			return nil, err
		}
		loaded.ConfigFile = configPath
		return loaded, nil
	}

	for _, f := range configSearchFiles {
		if fileExists(f) {
			loaded, err := loadFromFile(cfg, f)
			if err != nil {
				return nil, err
			}
			loaded.ConfigFile = f
			return loaded, nil
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

	format := extToViperType(strings.ToLower(filepath.Ext(path)))
	loaded, err := decodeConfig(cfg, data, format)
	if err != nil {
		return nil, fmt.Errorf("config file %s: %w", path, err)
	}
	return loaded, nil
}

// LoadConfigFromBytes parses config from raw bytes with the specified format
// ("json", "yaml", "toml"). Same pipeline as file loading: npm compat,
// unknown-key detection, validation.
func LoadConfigFromBytes(data []byte, format string) (*Config, error) {
	return decodeConfig(DefaultConfig(), data, format)
}

// decodeConfig is the single loading pipeline for every format:
// parse → normalize (npm compat, legacy keys → warnings) → strict decode
// (unknown keys are errors with suggestions) → plugin compat → validate.
func decodeConfig(cfg *Config, data []byte, format string) (*Config, error) {
	v := viper.New()
	v.SetConfigType(format)
	if err := v.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", format, err)
	}
	raw := v.AllSettings()

	plugins := normalizeRaw(raw, cfg)

	if err := decodeStrict(cfg, raw); err != nil {
		return nil, err
	}

	applyPluginCompat(cfg, plugins)

	// Record whether tagName was written by the user — the runner's v-prefix
	// inference only applies to the shipped default template.
	if gitRaw, ok := raw["git"].(map[string]interface{}); ok {
		_, cfg.Git.TagNameExplicit = gitRaw["tagname"]
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// decodeStrict maps the raw config onto the struct. Unknown keys are errors:
// a misspelled key (github.relase, hooks.preCommit) used to be dropped
// silently while the user believed the setting was active. A bare string
// where a list is expected becomes a one-element list — viper's default hook
// split "echo a, b" into two commands.
func decodeStrict(cfg *Config, raw map[string]interface{}) error {
	dc := &mapstructure.DecoderConfig{
		Result:           cfg,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		ErrorUnused:      true,
		DecodeHook:       stringToSliceHook,
	}
	dec, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return fmt.Errorf("preparing config decoder: %w", err)
	}
	if err := dec.Decode(raw); err != nil {
		return describeDecodeError(err)
	}
	return nil
}

// stringToSliceHook wraps a single string into []string without splitting.
func stringToSliceHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	if from.Kind() == reflect.String && to.Kind() == reflect.Slice && to.Elem().Kind() == reflect.String {
		return []string{data.(string)}, nil
	}
	return data, nil
}

// invalidKeysPattern matches mapstructure's unused-key diagnostics, e.g.
// "* 'github' has invalid keys: relase" or "* ” has invalid keys: relase".
var invalidKeysPattern = regexp.MustCompile(`^\*?\s*'([^']*)' has invalid keys: (.*)$`)

// describeDecodeError turns mapstructure's aggregated diagnostics ("N
// error(s) decoding:\n\n* ...") into actionable messages with a closest-key
// suggestion.
func describeDecodeError(err error) error {
	var msgs []string
	for _, line := range strings.Split(err.Error(), "\n") {
		line = strings.TrimSpace(line)
		// mapstructure prefixes its list with a header line ("decoding failed
		// due to the following error(s):"); only the entries carry information.
		if line == "" || strings.HasSuffix(line, ":") {
			continue
		}
		m := invalidKeysPattern.FindStringSubmatch(line)
		if m == nil {
			msgs = append(msgs, strings.TrimPrefix(line, "* "))
			continue
		}
		section := m[1]
		for _, key := range strings.Split(m[2], ",") {
			key = strings.TrimSpace(key)
			full := key
			if section != "" {
				full = section + "." + key
			}
			msg := fmt.Sprintf("unknown config key %q", full)
			if suggestion := suggestKey(section, key); suggestion != "" {
				msg += fmt.Sprintf(" (did you mean %q?)", suggestion)
			}
			msgs = append(msgs, msg)
		}
	}
	return fmt.Errorf("%s", strings.Join(msgs, "; "))
}

// knownKeys lists the mapstructure keys of the struct that section refers to
// ("" = top level, "github", "hooks", ...).
func knownKeys(section string) []string {
	t := reflect.TypeOf(Config{})
	if section != "" {
		found := false
		for _, part := range strings.Split(section, ".") {
			next, ok := fieldByKey(t, part)
			if !ok {
				return nil
			}
			t = next
			found = true
		}
		if !found {
			return nil
		}
	}
	var keys []string
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("mapstructure")
		if tag != "" && tag != "-" {
			keys = append(keys, tag)
		}
	}
	return keys
}

// fieldByKey resolves a nested struct type by mapstructure tag (case-insensitive).
func fieldByKey(t reflect.Type, key string) (reflect.Type, bool) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if strings.EqualFold(f.Tag.Get("mapstructure"), key) {
			ft := f.Type
			for ft.Kind() == reflect.Ptr || ft.Kind() == reflect.Slice {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				return ft, true
			}
			return nil, false
		}
	}
	return nil, false
}

// suggestKey returns the closest known key in the section: a case/hyphen
// variant first (preCommit → pre-commit), then edit distance ≤ 2.
func suggestKey(section string, unknown string) string {
	norm := func(s string) string {
		return strings.ToLower(strings.NewReplacer("-", "", "_", "", ":", "").Replace(s))
	}
	best, bestDist := "", 3
	for _, k := range knownKeys(section) {
		if norm(k) == norm(unknown) {
			return k
		}
		if d := levenshtein(strings.ToLower(unknown), strings.ToLower(k)); d < bestDist {
			best, bestDist = k, d
		}
	}
	return best
}

// levenshtein computes the edit distance between two short strings.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
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
