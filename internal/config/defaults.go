package config

// DefaultConfig returns a Config with all default values set.
func DefaultConfig() *Config {
	return &Config{
		Git: GitConfig{
			Commit:                     true,
			CommitMessage:              "chore: release v${version}",
			Tag:                        true,
			TagName:                    "${version}",
			TagAnnotation:              "Release ${version}",
			Push:                       true,
			PushArgs:                   []string{"--follow-tags"},
			PushRepo:                   "origin",
			RequireCleanWorkingDir:     true,
			RequireUpstream:            true,
			RequireCommits:             true,
			RequireConventionalCommits: true,
			Changelog:                  `git log --pretty=format:"* %s (%h)" ${latestTag}...HEAD`,
		},
		GitHub: GitHubConfig{
			ReleaseName: "Release ${version}",
			MakeLatest:  true,
			Host:        "api.github.com",
			TokenRef:    "GITHUB_TOKEN",
			Comments: GitHubCommentsConfig{
				Issue: ":rocket: _This issue has been resolved in v${version}._",
				PR:    ":rocket: _This pull request is included in v${version}._",
			},
		},
		GitLab: GitLabConfig{
			ReleaseName:                 "Release ${version}",
			TokenRef:                    "GITLAB_TOKEN",
			TokenHeader:                 "Private-Token",
			CertificateAuthorityFileRef: "CI_SERVER_TLS_CA_FILE",
		},
		Changelog: ChangelogConfig{
			Enabled: true,
			Preset:  "angular",
			Infile:  "CHANGELOG.md",
			Header:  "# Changelog",
		},
		CalVer: CalVerConfig{
			Format:            "yy.mm.minor",
			Increment:         "calendar",
			FallbackIncrement: "minor",
		},
		Notification: NotificationConfig{
			Enabled:  false,
			Webhooks: []WebhookConfig{},
		},
	}
}
