package config

import (
	"fmt"
	"regexp"
	"strings"

	"release-it-go/internal/version"
)

// preReleaseIDPattern: dot-separated semver identifiers (beta, rc.1, alpha-2).
var preReleaseIDPattern = regexp.MustCompile(`^[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*$`)

// calverFormats lists the CalVer formats the version package implements.
var calverFormats = map[string]bool{"yy.mm.minor": true, "yyyy.mm.minor": true, "yyyy.mm.dd": true}

// changelogPresets lists accepted conventional-changelog presets (both render
// with the built-in Angular-style renderer).
var changelogPresets = map[string]bool{"angular": true, "conventionalcommits": true}

// webhookTypes lists supported notification targets.
var webhookTypes = map[string]bool{"slack": true, "teams": true}

// bumperTypes lists supported bumper file formats ("" = detect by extension).
var bumperTypes = map[string]bool{"": true, "json": true, "yaml": true, "toml": true, "ini": true, "text": true}

// Validate rejects values the pipeline would otherwise accept silently and
// choke on later — after prerequisites, or after commit/tag/push. Every
// message names the config field.
func (c *Config) Validate() error {
	var problems []string
	add := func(format string, args ...interface{}) {
		problems = append(problems, fmt.Sprintf(format, args...))
	}

	if !strings.Contains(c.Git.TagName, "${version}") {
		add("git.tagName %q must contain ${version} (every release would produce the same tag)", c.Git.TagName)
	}

	if inc := c.Increment; inc != "" && inc != "no-increment" && !version.IsIncrementType(inc) {
		if _, err := version.ParseVersion(inc); err != nil {
			add("increment %q must be major|minor|patch|premajor|preminor|prepatch|prerelease or an explicit version like 1.5.0", inc)
		}
	}

	if id := c.PreReleaseID; id != "" && !preReleaseIDPattern.MatchString(id) {
		add("preReleaseId %q may only contain letters, digits, hyphens and dots", id)
	}

	if c.CalVer.Enabled && !calverFormats[c.CalVer.Format] {
		add("calver.format %q is not supported (yy.mm.minor, yyyy.mm.minor, yyyy.mm.dd)", c.CalVer.Format)
	}

	if strings.Contains(c.GitHub.Host, "://") {
		add("github.host %q must be a host name without a scheme (e.g. github.mycompany.com)", c.GitHub.Host)
	}
	if o := c.GitLab.Origin; o != "" && !strings.HasPrefix(o, "http://") && !strings.HasPrefix(o, "https://") {
		add("gitlab.origin %q must include the scheme (e.g. https://gitlab.mycompany.com)", o)
	}

	if c.GitHub.Timeout < 0 {
		add("github.timeout must be >= 0 (0 = default 30s), got %d", c.GitHub.Timeout)
	}

	for i, wh := range c.Notification.Webhooks {
		if !webhookTypes[strings.ToLower(wh.Type)] {
			add("notification.webhooks[%d].type %q is not supported (slack, teams)", i, wh.Type)
		}
		if wh.Timeout < 0 {
			add("notification.webhooks[%d].timeout must be >= 0, got %d", i, wh.Timeout)
		}
	}

	if c.Bumper.In != nil && !bumperTypes[strings.ToLower(c.Bumper.In.Type)] {
		add("bumper.in.type %q is not supported (json, yaml, toml, ini, text)", c.Bumper.In.Type)
	}
	for i, out := range c.Bumper.Out {
		if !bumperTypes[strings.ToLower(out.Type)] {
			add("bumper.out[%d].type %q is not supported (json, yaml, toml, ini, text)", i, out.Type)
		}
	}

	if p := c.Changelog.Preset; p != "" && !changelogPresets[p] {
		add("changelog.preset %q is not supported (angular, conventionalcommits)", p)
	}

	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
}
