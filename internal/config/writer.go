package config

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"go.yaml.in/yaml/v3"
)

// ForceFields maps section names (e.g. "git", "changelog") to field names
// (json tag names) that should always be included in the output, even if they
// match the default values. This is used by the init wizard so that every
// explicitly answered question appears in the generated config file.
type ForceFields map[string]map[string]bool

// WriteConfigJSON writes the given config to a JSON file at the specified path.
// Fields that match the default config values are omitted to keep the output minimal.
func WriteConfigJSON(cfg *Config, path string) error {
	return WriteConfigJSONWith(cfg, path, nil)
}

// WriteConfigJSONWith writes the config as JSON, always including the force fields.
func WriteConfigJSONWith(cfg *Config, path string, force ForceFields) error {
	m := toConfigMap(cfg, force)

	data, err := json.MarshalIndent(m, "", "  ")
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
	return toConfigMap(cfg, nil)
}

// toConfigMap converts a Config to a map. Fields matching defaults are omitted
// unless they appear in the force set for their section.
func toConfigMap(cfg *Config, force ForceFields) map[string]interface{} {
	defaults := DefaultConfig()
	result := make(map[string]interface{})

	if gitMap := diffStructForce(&cfg.Git, &defaults.Git, force["git"]); len(gitMap) > 0 {
		result["git"] = gitMap
	}

	if ghMap := diffStructForce(&cfg.GitHub, &defaults.GitHub, force["github"]); len(ghMap) > 0 {
		result["github"] = ghMap
	}

	if glMap := diffStructForce(&cfg.GitLab, &defaults.GitLab, force["gitlab"]); len(glMap) > 0 {
		result["gitlab"] = glMap
	}

	if clMap := diffStructForce(&cfg.Changelog, &defaults.Changelog, force["changelog"]); len(clMap) > 0 {
		result["changelog"] = clMap
	}

	return result
}

// diffStructForce compares two structs using reflection and returns a map of
// differing fields. Fields listed in forceKeys are always included regardless
// of whether they match defaults. Uses the json tag name as the key.
func diffStructForce(a, b interface{}, forceKeys map[string]bool) map[string]interface{} {
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

		if !reflect.DeepEqual(fa, fb) || forceKeys[jsonTag] {
			result[jsonTag] = fa
		}
	}

	return result
}

// WriteConfigYAML writes the given config to a YAML file at the specified path.
// Fields that match the default config values are omitted to keep the output minimal.
func WriteConfigYAML(cfg *Config, path string) error {
	return WriteConfigYAMLWith(cfg, path, nil)
}

// WriteConfigYAMLWith writes the config as YAML, always including the force fields.
func WriteConfigYAMLWith(cfg *Config, path string, force ForceFields) error {
	m := toConfigMap(cfg, force)

	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshaling config to YAML: %w", err)
	}

	if len(m) == 0 {
		data = []byte("{}\n")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file %s: %w", path, err)
	}

	return nil
}

// fullExampleYAML is a curated YAML string showing all available config options
// with documentation comments. This is the primary full-example format because
// YAML supports inline comments, making it much more useful as a reference.
const fullExampleYAML = `# ============================================================
# release-it-go — Full Configuration Reference
# Copy the options you need to .release-it-go.yaml
# Docs: https://github.com/user/release-it-go
# ============================================================

# Git commit, tag and push settings
git:
  commit: true
  # Commit message template. Use ${version} for the new version number
  commitMessage: "chore: release v${version}"
  # Extra arguments passed to git commit command
  commitArgs: []
  tag: true
  # Tag name format. Use ${version} for the new version number
  tagName: "v${version}"
  # Glob pattern to match existing tags for version detection
  tagMatch: "v*"
  # Glob pattern to exclude tags (e.g. release candidates)
  tagExclude: "*-rc.*"
  # Annotation message for annotated tags
  tagAnnotation: "Release ${version}"
  # Extra arguments passed to git tag command
  tagArgs: []
  push: true
  # Extra arguments passed to git push command
  pushArgs: ["--follow-tags"]
  # Remote repository name for push
  pushRepo: "origin"
  # Only allow releases from this branch (empty = any branch)
  requireBranch: "main"
  # Require clean working directory before release
  requireCleanWorkingDir: true
  # Require upstream branch to be configured
  requireUpstream: true
  # Require new commits since latest tag
  requireCommits: true
  # Require all commits to follow conventional commit format
  requireConventionalCommits: true
  # Search all refs (not just current branch) for latest tag
  getLatestTagFromAllRefs: false
  # Stage untracked files before commit
  addUntrackedFiles: false

# GitHub release settings
github:
  # Create a GitHub release
  release: true
  # Release title template
  releaseName: "Release ${version}"
  # Create as draft release
  draft: false
  # Mark as pre-release
  preRelease: false
  # Mark as latest release
  makeLatest: true
  # Auto-generate release notes via GitHub API
  autoGenerate: false
  # Glob patterns for release assets to upload
  assets: ["dist/*.tar.gz", "dist/*.zip"]
  # GitHub API host (change for GitHub Enterprise)
  host: "api.github.com"
  # Environment variable name containing GitHub token
  tokenRef: "GITHUB_TOKEN"
  # API request timeout in seconds
  timeout: 30
  # Skip GitHub API pre-flight checks
  skipChecks: false
  # Open release URL in browser after creation
  web: false
  # Automated comments on issues/PRs included in the release
  comments:
    # Enable automated comments
    submit: false
    # Comment template for resolved issues
    issue: ":rocket: _This issue has been resolved in v${version}._"
    # Comment template for included pull requests
    pr: ":rocket: _This pull request is included in v${version}._"

# GitLab release settings
gitlab:
  # Create a GitLab release
  release: false
  # Release title template
  releaseName: "Release ${version}"
  # Mark as pre-release (upcoming release)
  preRelease: false
  # Associate release with milestones
  milestones: []
  # Release asset links
  assets: []
  # Environment variable name containing GitLab token
  tokenRef: "GITLAB_TOKEN"
  # HTTP header name for authentication
  tokenHeader: "Private-Token"
  # GitLab instance URL (for self-hosted)
  origin: "https://gitlab.example.com"
  # Skip GitLab API pre-flight checks
  skipChecks: false
  # Use HTTPS for API calls
  secure: false

# Lifecycle hooks — run shell commands at specific points
hooks:
  "before:init": []
  "after:init": []
  "before:bump": []
  # Example: notify after version bump
  "after:bump": ["echo 'Bumped to v${version}'"]
  "before:release": []
  # Example: notify after release
  "after:release": ["echo 'Released v${version}'"]
  "before:git:release": []
  "after:git:release": []
  "before:github:release": []
  "after:github:release": []
  "before:gitlab:release": []
  "after:gitlab:release": []

# Changelog generation settings
changelog:
  # Enable changelog generation
  enabled: true
  # Changelog format preset (angular = conventional-changelog)
  preset: "angular"
  # File path for changelog output
  infile: "CHANGELOG.md"
  # Header text at the top of changelog file
  header: "# Changelog"
  # Use Keep a Changelog format instead of conventional
  keepAChangelog: false
  # Add Unreleased section to changelog
  addUnreleased: false
  # Keep Unreleased section after release
  keepUnreleased: false
  # Add compare URL for each version
  addVersionUrl: false

# Bumper — update version in multiple files
bumper:
  # Enable bumper
  enabled: false
  # Source file to read current version from
  in:
    file: "VERSION"
    consumeWholeFile: true
  # Target files to write new version to
  out:
    - file: "VERSION"
      consumeWholeFile: true
    - file: "package.json"
      path: "version"

# Calendar Versioning (CalVer) settings
calver:
  # Enable CalVer (disables SemVer)
  enabled: false
  # CalVer format pattern
  format: "yy.mm.minor"
  # How to increment calendar version
  increment: "calendar"
  # Fallback increment when no calendar change
  fallbackIncrement: "minor"

# Webhook notifications after release
notification:
  # Enable notifications
  enabled: false
  webhooks:
    # Slack webhook
    - type: "slack"
      # Environment variable name containing Slack webhook URL
      urlRef: "SLACK_WEBHOOK_URL"
    # Microsoft Teams webhook (sends rich MessageCard with facts and changelog)
    - type: "teams"
      # Environment variable name containing Teams webhook URL
      urlRef: "TEAMS_WEBHOOK_URL"
      # Custom message template (overrides rich card, supports ${version}, ${releaseUrl}, ${repo.repository})
      # messageTemplate: ""
      # Card theme color (hex, default: "0076D7")
      themeColor: "0076D7"
      # Card activity image URL
      # imageUrl: "https://example.com/logo.png"
      # Contributors to exclude from notifications (e.g., bot accounts)
      ignoredContributors: ["Jenkins", "GitLab Bot"]
      # Request timeout in seconds
      timeout: 30
`

// WriteFullExampleYAML writes the curated full example YAML config to a file.
func WriteFullExampleYAML(path string) error {
	if err := os.WriteFile(path, []byte(fullExampleYAML), 0644); err != nil {
		return fmt.Errorf("writing config file %s: %w", path, err)
	}
	return nil
}
