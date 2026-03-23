package config

import (
	"encoding/json"
	"strings"
)

// normalizeJSON pre-processes JSON config data to fix type mismatches between
// the original npm release-it format and the Go struct types.
// This runs BEFORE Viper unmarshal to prevent type errors.
func normalizeJSON(data []byte) []byte {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Return original data; Viper will report the parse error with context
		return data
	}

	changed := false

	// Remove unknown top-level keys that don't map to Config struct
	for _, key := range []string{"npm", "plugins", "versionFile"} {
		if _, ok := raw[key]; ok {
			delete(raw, key)
			changed = true
		}
	}

	// Fix git.requireBranch: []string → join to string
	if gitRaw, ok := raw["git"].(map[string]interface{}); ok {
		if rb, ok := gitRaw["requireBranch"]; ok {
			switch v := rb.(type) {
			case []interface{}:
				parts := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok {
						parts = append(parts, s)
					}
				}
				gitRaw["requireBranch"] = strings.Join(parts, ",")
				changed = true
			}
		}
		// Remove changelogFile - not in our struct (we use changelog.infile)
		if _, ok := gitRaw["changelogFile"]; ok {
			delete(gitRaw, "changelogFile")
			changed = true
		}
	}

	// Fix gitlab.assets: {links:[]} → [] (empty string array)
	if glRaw, ok := raw["gitlab"].(map[string]interface{}); ok {
		if assets, ok := glRaw["assets"]; ok {
			if _, isMap := assets.(map[string]interface{}); isMap {
				glRaw["assets"] = []string{}
				changed = true
			}
		}
	}

	// Fix github.assets: same as gitlab
	if ghRaw, ok := raw["github"].(map[string]interface{}); ok {
		if assets, ok := ghRaw["assets"]; ok {
			if _, isMap := assets.(map[string]interface{}); isMap {
				ghRaw["assets"] = []string{}
				changed = true
			}
		}
	}

	if !changed {
		return data
	}

	normalized, err := json.Marshal(raw)
	if err != nil {
		return data
	}
	return normalized
}

// applyPluginCompat reads the original JSON config to detect release-it npm
// plugin settings and maps them to the Go built-in equivalents.
// This provides backward compatibility with existing .release-it.json files
// that use the @release-it/conventional-changelog or
// @release-it/keep-a-changelog plugins.
func applyPluginCompat(cfg *Config, data []byte, format string) {
	if format != "json" {
		return
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	pluginsRaw, ok := raw["plugins"]
	if !ok {
		return
	}

	var plugins map[string]json.RawMessage
	if err := json.Unmarshal(pluginsRaw, &plugins); err != nil {
		return
	}

	// Map @release-it/conventional-changelog plugin settings
	for key, val := range plugins {
		if strings.Contains(key, "conventional-changelog") {
			applyConventionalChangelogPlugin(cfg, val)
		}
		if strings.Contains(key, "keep-a-changelog") {
			cfg.Changelog.KeepAChangelog = true
			applyKeepAChangelogPlugin(cfg, val)
		}
	}
}

// applyConventionalChangelogPlugin maps conventional-changelog plugin settings.
func applyConventionalChangelogPlugin(cfg *Config, data json.RawMessage) {
	// Use a raw map to detect which fields are explicitly set in the plugin config.
	// This prevents bool zero-value (false) from overriding user's explicit config.
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawFields); err != nil {
		return
	}

	var pluginCfg struct {
		Preset    string `json:"preset"`
		Infile    string `json:"infile"`
		Header    string `json:"header"`
		Changelog bool   `json:"changelog"`
	}
	if err := json.Unmarshal(data, &pluginCfg); err != nil {
		return
	}

	if pluginCfg.Preset != "" {
		cfg.Changelog.Preset = pluginCfg.Preset
	}
	if pluginCfg.Infile != "" {
		cfg.Changelog.Infile = pluginCfg.Infile
	}
	if pluginCfg.Header != "" {
		cfg.Changelog.Header = pluginCfg.Header
	}
	// Only override Enabled if the plugin explicitly sets "changelog" field
	if _, ok := rawFields["changelog"]; ok {
		cfg.Changelog.Enabled = pluginCfg.Changelog
	}
}

// applyKeepAChangelogPlugin maps keep-a-changelog plugin settings.
func applyKeepAChangelogPlugin(cfg *Config, data json.RawMessage) {
	var pluginCfg struct {
		Filename string `json:"filename"`
		Head     string `json:"head"`
	}
	if err := json.Unmarshal(data, &pluginCfg); err != nil {
		return
	}

	if pluginCfg.Filename != "" {
		cfg.Changelog.Infile = pluginCfg.Filename
	}
	if pluginCfg.Head != "" {
		cfg.Changelog.Header = pluginCfg.Head
	}
}
