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

// fullExampleJSON is a curated JSON string showing all available config options
// with sensible, usable example values. Only meaningful options are included —
// runtime flags (ci, dry-run, verbose) and zero-value noise are omitted.
const fullExampleJSON = `{
  "git": {
    "commit": true,
    "commitMessage": "chore: release v${version}",
    "commitArgs": [],
    "tag": true,
    "tagName": "v${version}",
    "tagMatch": "v*",
    "tagExclude": "*-rc.*",
    "tagAnnotation": "Release ${version}",
    "tagArgs": [],
    "push": true,
    "pushArgs": ["--follow-tags"],
    "pushRepo": "origin",
    "requireBranch": "main",
    "requireCleanWorkingDir": true,
    "requireUpstream": true,
    "requireCommits": true,
    "requireConventionalCommits": true,
    "getLatestTagFromAllRefs": false,
    "addUntrackedFiles": false
  },
  "github": {
    "release": true,
    "releaseName": "Release ${version}",
    "draft": false,
    "preRelease": false,
    "makeLatest": true,
    "autoGenerate": false,
    "assets": ["dist/*.tar.gz", "dist/*.zip"],
    "host": "api.github.com",
    "tokenRef": "GITHUB_TOKEN",
    "timeout": 30,
    "skipChecks": false,
    "web": false,
    "comments": {
      "submit": false,
      "issue": ":rocket: _This issue has been resolved in v${version}._",
      "pr": ":rocket: _This pull request is included in v${version}._"
    }
  },
  "gitlab": {
    "release": false,
    "releaseName": "Release ${version}",
    "preRelease": false,
    "milestones": [],
    "assets": [],
    "tokenRef": "GITLAB_TOKEN",
    "tokenHeader": "Private-Token",
    "origin": "https://gitlab.example.com",
    "skipChecks": false,
    "secure": false
  },
  "hooks": {
    "before:init": [],
    "after:init": [],
    "before:bump": [],
    "after:bump": ["echo 'Bumped to v${version}'"],
    "before:release": [],
    "after:release": ["echo 'Released v${version}'"],
    "before:git:release": [],
    "after:git:release": [],
    "before:github:release": [],
    "after:github:release": [],
    "before:gitlab:release": [],
    "after:gitlab:release": []
  },
  "changelog": {
    "enabled": true,
    "preset": "angular",
    "infile": "CHANGELOG.md",
    "header": "# Changelog",
    "keepAChangelog": false,
    "addUnreleased": false,
    "keepUnreleased": false,
    "addVersionUrl": false
  },
  "bumper": {
    "enabled": false,
    "in": {
      "file": "VERSION",
      "consumeWholeFile": true
    },
    "out": [
      { "file": "VERSION", "consumeWholeFile": true },
      { "file": "package.json", "path": "version" }
    ]
  },
  "calver": {
    "enabled": false,
    "format": "yy.mm.minor",
    "increment": "calendar",
    "fallbackIncrement": "minor"
  },
  "notification": {
    "enabled": false,
    "webhooks": [
      {
        "type": "slack",
        "urlRef": "SLACK_WEBHOOK_URL"
      },
      {
        "type": "teams",
        "urlRef": "TEAMS_WEBHOOK_URL",
        "messageTemplate": "🚀 ${repo.repository} v${version} released!\n${releaseUrl}",
        "timeout": 30
      }
    ]
  }
}
`

// WriteFullExampleJSON writes the curated full example config to a file.
func WriteFullExampleJSON(path string) error {
	if err := os.WriteFile(path, []byte(fullExampleJSON), 0644); err != nil {
		return fmt.Errorf("writing config file %s: %w", path, err)
	}
	return nil
}
