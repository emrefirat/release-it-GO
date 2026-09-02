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
			PushArgs:                   []string{"--follow-tags", "--atomic"},
			PushRepo:                   "origin",
			RequireCleanWorkingDir:     true,
			RequireUpstream:            true,
			RequireCommits:             true,
			RequireConventionalCommits: true,
		},
		GitHub: GitHubConfig{
			ReleaseName: "Release ${version}",
			MakeLatest:  true,
			Host:        "github.com",
			TokenRef:    "GITHUB_TOKEN",
		},
		GitLab: GitLabConfig{
			ReleaseName:                 "Release ${version}",
			TokenRef:                    "GITLAB_TOKEN",
			TokenHeader:                 "Private-Token",
			CertificateAuthorityFileRef: "CI_SERVER_TLS_CA_FILE",
			// TLS verification is on by default; secure: false is an explicit
			// opt-out for self-signed instances without a CA file.
			Secure: true,
			// Generic Package Registry keeps the historical upload behavior;
			// false switches to the project uploads API (npm's default).
			UseGenericPackageRepositoryForAssets: true,
		},
		Changelog: ChangelogConfig{
			Enabled: true,
			Preset:  "angular",
			Infile:  "CHANGELOG.md",
			Header:  "# Changelog",
			// compare links were always emitted before; now honored as a switch
			AddVersionUrl: true,
		},
		CalVer: CalVerConfig{
			Format: "yy.mm.minor",
		},
		Notification: NotificationConfig{
			Enabled:  false,
			Webhooks: []WebhookConfig{},
		},
	}
}
