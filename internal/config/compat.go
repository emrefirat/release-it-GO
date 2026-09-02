package config

import (
	"fmt"
	"strings"
)

// legacyKey describes a config key that is accepted for backward
// compatibility but has no effect. It is removed from the raw map with a
// warning instead of failing the unknown-key check.
type legacyKey struct {
	canonical string // original spelling for messages (viper lowercases keys)
	reason    string
}

// legacyTopLevelKeys are npm release-it keys with no Go equivalent.
var legacyTopLevelKeys = map[string]legacyKey{
	"npm":         {"npm", "npm publishing is not supported (npm-only)"},
	"versionfile": {"versionFile", "use bumper.in / bumper.out instead"},
}

// legacySectionKeys are per-section keys that are npm-only or were removed
// because they never had an effect.
var legacySectionKeys = map[string]map[string]legacyKey{
	"git": {
		"changelogfile": {"changelogFile", "use changelog.infile instead"},
		"changelog":     {"changelog", "removed: the conventional-commit renderer generates the changelog"},
	},
	"github": {
		"releasenotes": {"releaseNotes", "removed: release notes come from the generated changelog"},
		"web":          {"web", "removed: no web-based release flow"},
	},
	"gitlab": {
		"releasenotes": {"releaseNotes", "removed: release notes come from the generated changelog"},
		"prerelease":   {"preRelease", "removed: GitLab releases have no pre-release flag"},
	},
	"changelog": {
		"addunreleased":  {"addUnreleased", "removed: never implemented"},
		"keepunreleased": {"keepUnreleased", "removed: never implemented"},
	},
	"calver": {
		"increment":         {"increment", "removed: CalVer increments by calendar change, otherwise minor"},
		"fallbackincrement": {"fallbackIncrement", "removed: CalVer increments by calendar change, otherwise minor"},
	},
}

// normalizeRaw adapts a parsed config map (any format; keys lowercased by
// viper) to the Go struct shape: npm-only and removed keys are dropped with a
// warning, npm array/object shapes are converted, and the plugins section is
// returned for applyPluginCompat. Previously only JSON files got this
// treatment, so a legacy .release-it.yaml hard-failed on requireBranch arrays
// and silently ignored plugin settings.
func normalizeRaw(raw map[string]interface{}, cfg *Config) map[string]interface{} {
	var plugins map[string]interface{}
	if p, ok := raw["plugins"].(map[string]interface{}); ok {
		plugins = p
	}
	if _, ok := raw["plugins"]; ok {
		delete(raw, "plugins")
	}

	for key, lk := range legacyTopLevelKeys {
		if _, ok := raw[key]; ok {
			delete(raw, key)
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("ignored %q: %s", lk.canonical, lk.reason))
		}
	}

	for section, keys := range legacySectionKeys {
		sec, ok := raw[section].(map[string]interface{})
		if !ok {
			continue
		}
		for key, lk := range keys {
			if _, present := sec[key]; present {
				delete(sec, key)
				cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("ignored %q: %s", section+"."+lk.canonical, lk.reason))
			}
		}
	}

	if gitRaw, ok := raw["git"].(map[string]interface{}); ok {
		// npm: requireBranch may be an array meaning "any of these"
		if list, isList := gitRaw["requirebranch"].([]interface{}); isList {
			parts := make([]string, 0, len(list))
			for _, item := range list {
				if s, ok := item.(string); ok {
					parts = append(parts, s)
				}
			}
			gitRaw["requirebranch"] = strings.Join(parts, ",")
		}
	}

	// npm: assets may be {links: [...]} objects; only glob lists are supported
	for _, section := range []string{"github", "gitlab"} {
		sec, ok := raw[section].(map[string]interface{})
		if !ok {
			continue
		}
		if _, isMap := sec["assets"].(map[string]interface{}); isMap {
			sec["assets"] = []string{}
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("ignored %s.assets object form: only a list of file globs is supported", section))
		}
	}

	// bumper.out[].versionPrefix was never read; prefix is the real field
	if bumperRaw, ok := raw["bumper"].(map[string]interface{}); ok {
		if outs, ok := bumperRaw["out"].([]interface{}); ok {
			for i, o := range outs {
				if m, ok := o.(map[string]interface{}); ok {
					if _, present := m["versionprefix"]; present {
						delete(m, "versionprefix")
						cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("ignored bumper.out[%d].versionPrefix: use prefix instead", i))
					}
				}
			}
		}
	}

	return plugins
}

// applyPluginCompat maps npm release-it plugin settings
// (@release-it/conventional-changelog, @release-it/keep-a-changelog) onto the
// built-in equivalents. Works for every config format.
func applyPluginCompat(cfg *Config, plugins map[string]interface{}) {
	for key, val := range plugins {
		settings, _ := val.(map[string]interface{})
		if strings.Contains(key, "conventional-changelog") {
			applyConventionalChangelogPlugin(cfg, settings)
		}
		if strings.Contains(key, "keep-a-changelog") {
			cfg.Changelog.KeepAChangelog = true
			applyKeepAChangelogPlugin(cfg, settings)
		}
	}
}

// applyConventionalChangelogPlugin maps conventional-changelog plugin settings.
// Only fields present in the plugin config override the user's explicit
// values (a missing "changelog" bool must not disable the changelog).
func applyConventionalChangelogPlugin(cfg *Config, settings map[string]interface{}) {
	if s, ok := settings["preset"].(string); ok && s != "" {
		cfg.Changelog.Preset = s
	}
	if s, ok := settings["infile"].(string); ok && s != "" {
		cfg.Changelog.Infile = s
	}
	if s, ok := settings["header"].(string); ok && s != "" {
		cfg.Changelog.Header = s
	}
	if b, ok := settings["changelog"].(bool); ok {
		cfg.Changelog.Enabled = b
	}
}

// applyKeepAChangelogPlugin maps keep-a-changelog plugin settings.
func applyKeepAChangelogPlugin(cfg *Config, settings map[string]interface{}) {
	if s, ok := settings["filename"].(string); ok && s != "" {
		cfg.Changelog.Infile = s
	}
	if s, ok := settings["head"].(string); ok && s != "" {
		cfg.Changelog.Header = s
	}
}
