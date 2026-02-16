package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

// WriteConfigJSON writes the given config to a JSON file at the specified path.
// Fields that match the default config values are omitted to keep the output minimal.
func WriteConfigJSON(cfg *Config, path string) error {
	minimal := toMinimalMap(cfg)

	data, err := json.MarshalIndent(minimal, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config to JSON: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file %s: %w", path, err)
	}

	return nil
}

// toMinimalMap converts a Config to a map, omitting fields that match defaults.
func toMinimalMap(cfg *Config) map[string]interface{} {
	defaults := DefaultConfig()
	result := make(map[string]interface{})

	// Git section
	if gitMap := diffGit(&cfg.Git, &defaults.Git); len(gitMap) > 0 {
		result["git"] = gitMap
	}

	// GitHub section
	if ghMap := diffGitHub(&cfg.GitHub, &defaults.GitHub); len(ghMap) > 0 {
		result["github"] = ghMap
	}

	// GitLab section
	if glMap := diffGitLab(&cfg.GitLab, &defaults.GitLab); len(glMap) > 0 {
		result["gitlab"] = glMap
	}

	// Changelog section
	if clMap := diffChangelog(&cfg.Changelog, &defaults.Changelog); len(clMap) > 0 {
		result["changelog"] = clMap
	}

	return result
}

// diffStruct compares two structs using reflection and returns a map of differing fields.
// Uses the json tag name as the key.
func diffStruct(a, b interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	va := reflect.ValueOf(a).Elem()
	vb := reflect.ValueOf(b).Elem()
	t := va.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fa := va.Field(i).Interface()
		fb := vb.Field(i).Interface()

		if !reflect.DeepEqual(fa, fb) {
			result[jsonTag] = fa
		}
	}

	return result
}

func diffGit(a, b *GitConfig) map[string]interface{} {
	return diffStruct(a, b)
}

func diffGitHub(a, b *GitHubConfig) map[string]interface{} {
	return diffStruct(a, b)
}

func diffGitLab(a, b *GitLabConfig) map[string]interface{} {
	return diffStruct(a, b)
}

func diffChangelog(a, b *ChangelogConfig) map[string]interface{} {
	return diffStruct(a, b)
}

// WriteFullConfigJSON writes the entire config struct to a JSON file,
// including all fields regardless of whether they match defaults.
func WriteFullConfigJSON(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling full config to JSON: %w", err)
	}

	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file %s: %w", path, err)
	}

	return nil
}
