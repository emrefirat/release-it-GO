package config

import (
	"fmt"
	"regexp"
	"sort"
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
	"npm":            {"npm", "npm publishing is not supported (npm-only)"},
	"versionfile":    {"versionFile", "use bumper.in / bumper.out instead"},
	"prereleasebase": {"preReleaseBase", "not supported: pre-release numbering always starts at 0"},
}

// schemaKey is the JSON Schema pointer most npm configs carry
// ("$schema": "https://unpkg.com/release-it/schema/release-it.json"). It is
// editor metadata, not configuration: accepted silently.
const schemaKey = "$schema"

// pluginHookPattern matches npm's plugin-scoped lifecycle hooks
// (before:git:bump, after:npm:release, before:github:init, ...). Only the
// git/github/gitlab release stages exist here; the rest map onto the
// unscoped pipeline events.
var pluginHookPattern = regexp.MustCompile(`^(before|after):(git|npm|github|gitlab):(init|bump|release)$`)

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
		"comments":     {"comments", "not implemented yet: issue/PR comments are not posted"},
		"update":       {"update", "not supported: an existing release for the tag is not updated"},
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
	delete(raw, "plugins")
	delete(raw, schemaKey)

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

	if hooksRaw, ok := raw["hooks"].(map[string]interface{}); ok {
		dropPluginScopedHooks(hooksRaw, cfg)
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

// dropPluginScopedHooks removes npm's plugin-scoped hook keys that have no
// struct field (before:git:bump, after:npm:release, ...) with a warning that
// names the unscoped equivalent. They used to fail the unknown-key check,
// which made every npm config that had them unloadable.
func dropPluginScopedHooks(hooksRaw map[string]interface{}, cfg *Config) {
	known := make(map[string]bool)
	for _, k := range knownKeys("hooks") {
		known[strings.ToLower(k)] = true
	}
	keys := make([]string, 0, len(hooksRaw))
	for key := range hooksRaw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if known[key] {
			continue
		}
		m := pluginHookPattern.FindStringSubmatch(key)
		if m == nil {
			continue
		}
		delete(hooksRaw, key)
		if m[2] == "npm" {
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("ignored \"hooks.%s\": npm publishing is not supported", key))
			continue
		}
		cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("ignored \"hooks.%s\": plugin-scoped hooks are not supported, use \"%s:%s\"", key, m[1], m[3]))
	}
}

// applyPluginCompat maps npm release-it plugin settings
// (@release-it/conventional-changelog, @release-it/keep-a-changelog,
// @release-it/bumper) onto the built-in equivalents. Any other plugin is
// dropped with a warning instead of silently. Works for every config format.
func applyPluginCompat(cfg *Config, plugins map[string]interface{}) {
	keys := make([]string, 0, len(plugins))
	for key := range plugins {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		settings, _ := plugins[key].(map[string]interface{})
		switch {
		case strings.Contains(key, "conventional-changelog"):
			applyConventionalChangelogPlugin(cfg, settings)
		case strings.Contains(key, "keep-a-changelog"):
			cfg.Changelog.KeepAChangelog = true
			applyKeepAChangelogPlugin(cfg, settings)
		case strings.Contains(key, "bumper"):
			applyBumperPlugin(cfg, settings)
		default:
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("ignored plugins.%q: not supported (everything is built in; see the hooks and bumper sections)", key))
		}
	}
}

// applyBumperPlugin maps @release-it/bumper settings onto the built-in bumper
// section. npm accepts a file name, a {file, path, type} object, or (for
// out) a list of either.
func applyBumperPlugin(cfg *Config, settings map[string]interface{}) {
	if in, ok := parseBumperFileSpec(settings["in"]); ok {
		cfg.Bumper.In = &in
		cfg.Bumper.Enabled = true
	}
	var outs []BumperFile
	switch v := settings["out"].(type) {
	case []interface{}:
		for _, item := range v {
			if f, ok := parseBumperFileSpec(item); ok {
				outs = append(outs, f)
			}
		}
	default:
		if f, ok := parseBumperFileSpec(v); ok {
			outs = append(outs, f)
		}
	}
	if len(outs) > 0 {
		cfg.Bumper.Out = append(cfg.Bumper.Out, outs...)
		cfg.Bumper.Enabled = true
	}
}

// parseBumperFileSpec converts an npm bumper file spec ("VERSION" or
// {file, path, type, prefix, consumeWholeFile}) into a BumperFile.
func parseBumperFileSpec(spec interface{}) (BumperFile, bool) {
	switch v := spec.(type) {
	case string:
		if v == "" {
			return BumperFile{}, false
		}
		return BumperFile{File: v}, true
	case map[string]interface{}:
		f := BumperFile{}
		f.File, _ = v["file"].(string)
		if f.File == "" {
			return BumperFile{}, false
		}
		f.Path, _ = v["path"].(string)
		f.Type, _ = v["type"].(string)
		f.Prefix, _ = v["prefix"].(string)
		f.ConsumeWholeFile, _ = v["consumewholefile"].(bool)
		return f, true
	}
	return BumperFile{}, false
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
