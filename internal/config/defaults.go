package config

// FullExampleConfig returns a Config with all fields populated as a
// comprehensive example. Used by `init --full-example` to show every
// available option so users can pick what they need.
func FullExampleConfig() *Config {
	cfg := DefaultConfig()

	// Git: show all options with sensible example values
	cfg.Git.CommitArgs = []string{}
	cfg.Git.TagMatch = ""
	cfg.Git.TagExclude = ""
	cfg.Git.TagArgs = []string{}
	cfg.Git.PushArgs = []string{"--follow-tags"}
	cfg.Git.RequireBranch = "main"
	cfg.Git.GetLatestTagFromAllRefs = false
	cfg.Git.CommitsPath = ""
	cfg.Git.AddUntrackedFiles = false

	// GitHub: show all options
	cfg.GitHub.Release = true
	cfg.GitHub.Draft = false
	cfg.GitHub.PreRelease = false
	cfg.GitHub.AutoGenerate = false
	cfg.GitHub.Assets = []string{"dist/*.tar.gz", "dist/*.zip"}
	cfg.GitHub.Timeout = 0
	cfg.GitHub.Proxy = ""
	cfg.GitHub.SkipChecks = false
	cfg.GitHub.Web = false
	cfg.GitHub.Comments.Submit = false
	cfg.GitHub.DiscussionCategoryName = ""

	// GitLab: show all options
	cfg.GitLab.Release = false
	cfg.GitLab.PreRelease = false
	cfg.GitLab.Milestones = []string{}
	cfg.GitLab.Assets = []string{}
	cfg.GitLab.Origin = ""
	cfg.GitLab.SkipChecks = false
	cfg.GitLab.CertificateAuthorityFile = ""
	cfg.GitLab.Secure = false
	cfg.GitLab.UseGenericPackageRepositoryForAssets = false

	// Hooks: show all lifecycle hooks with example commands
	cfg.Hooks.BeforeInit = []string{}
	cfg.Hooks.AfterInit = []string{}
	cfg.Hooks.BeforeBump = []string{}
	cfg.Hooks.AfterBump = []string{"echo 'Bumped to v${version}'"}
	cfg.Hooks.BeforeRelease = []string{}
	cfg.Hooks.AfterRelease = []string{}
	cfg.Hooks.BeforeGitRelease = []string{}
	cfg.Hooks.AfterGitRelease = []string{}
	cfg.Hooks.BeforeGitHubRelease = []string{}
	cfg.Hooks.AfterGitHubRelease = []string{}
	cfg.Hooks.BeforeGitLabRelease = []string{}
	cfg.Hooks.AfterGitLabRelease = []string{}

	// Changelog: show all options
	cfg.Changelog.AddUnreleased = false
	cfg.Changelog.KeepUnreleased = false
	cfg.Changelog.AddVersionUrl = false

	// Bumper: show example with VERSION file
	cfg.Bumper.Enabled = false
	cfg.Bumper.In = &BumperFile{
		File:             "VERSION",
		ConsumeWholeFile: true,
	}
	cfg.Bumper.Out = []BumperFile{
		{File: "VERSION", ConsumeWholeFile: true},
		{File: "package.json", Path: "version"},
	}

	// CalVer: show example
	cfg.CalVer.Enabled = false

	// Notification: show example with Slack and Teams
	cfg.Notification.Enabled = false
	cfg.Notification.Webhooks = []WebhookConfig{
		{
			Type:   "slack",
			URLRef: "SLACK_WEBHOOK_URL",
		},
		{
			Type:            "teams",
			URLRef:          "TEAMS_WEBHOOK_URL",
			MessageTemplate: "🚀 ${repo.repository} v${version} released!\n${releaseUrl}",
			Timeout:         30,
		},
	}

	return cfg
}

// DefaultConfig returns a Config with all default values set.
func DefaultConfig() *Config {
	return &Config{
		Git: GitConfig{
			Commit:                 true,
			CommitMessage:          "chore: release v${version}",
			Tag:                    true,
			TagName:                "${version}",
			TagAnnotation:          "Release ${version}",
			Push:                   true,
			PushArgs:               []string{"--follow-tags"},
			PushRepo:               "origin",
			RequireCleanWorkingDir:     true,
			RequireUpstream:            true,
			RequireCommits:             true,
			RequireConventionalCommits: true,
			Changelog:              `git log --pretty=format:"* %s (%h)" ${latestTag}...HEAD`,
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
