// Package config provides configuration management for release-it-go.
// It supports loading configuration from JSON, YAML, and TOML files,
// environment variables, and CLI flags.
package config

// Config is the root configuration struct for release-it-go.
type Config struct {
	Git          GitConfig          `json:"git" yaml:"git" toml:"git" mapstructure:"git"`
	GitHub       GitHubConfig       `json:"github" yaml:"github" toml:"github" mapstructure:"github"`
	GitLab       GitLabConfig       `json:"gitlab" yaml:"gitlab" toml:"gitlab" mapstructure:"gitlab"`
	Hooks        HooksConfig        `json:"hooks" yaml:"hooks" toml:"hooks" mapstructure:"hooks"`
	Changelog    ChangelogConfig    `json:"changelog" yaml:"changelog" toml:"changelog" mapstructure:"changelog"`
	Bumper       BumperConfig       `json:"bumper" yaml:"bumper" toml:"bumper" mapstructure:"bumper"`
	CalVer       CalVerConfig       `json:"calver" yaml:"calver" toml:"calver" mapstructure:"calver"`
	Notification NotificationConfig `json:"notification" yaml:"notification" toml:"notification" mapstructure:"notification"`
	CI           bool               `json:"ci" yaml:"ci" toml:"ci" mapstructure:"ci"`
	DryRun       bool               `json:"dry-run" yaml:"dry-run" toml:"dry-run" mapstructure:"dry-run"`
	Verbose      int                `json:"verbose" yaml:"verbose" toml:"verbose" mapstructure:"verbose"`
	Increment    string             `json:"increment" yaml:"increment" toml:"increment" mapstructure:"increment"`
	PreReleaseID string             `json:"preReleaseId" yaml:"preReleaseId" toml:"preReleaseId" mapstructure:"preReleaseId"`
	ConfigFile   string             `json:"-" yaml:"-" toml:"-" mapstructure:"-"` // loaded config file path (not serialized)
}

// GitConfig holds all git-related configuration.
type GitConfig struct {
	Commit                     bool     `json:"commit" yaml:"commit" toml:"commit" mapstructure:"commit"`
	CommitMessage              string   `json:"commitMessage" yaml:"commitMessage" toml:"commitMessage" mapstructure:"commitMessage"`
	CommitArgs                 []string `json:"commitArgs" yaml:"commitArgs" toml:"commitArgs" mapstructure:"commitArgs"`
	Tag                        bool     `json:"tag" yaml:"tag" toml:"tag" mapstructure:"tag"`
	TagName                    string   `json:"tagName" yaml:"tagName" toml:"tagName" mapstructure:"tagName"`
	TagMatch                   string   `json:"tagMatch" yaml:"tagMatch" toml:"tagMatch" mapstructure:"tagMatch"`
	TagExclude                 string   `json:"tagExclude" yaml:"tagExclude" toml:"tagExclude" mapstructure:"tagExclude"`
	TagAnnotation              string   `json:"tagAnnotation" yaml:"tagAnnotation" toml:"tagAnnotation" mapstructure:"tagAnnotation"`
	TagArgs                    []string `json:"tagArgs" yaml:"tagArgs" toml:"tagArgs" mapstructure:"tagArgs"`
	Push                       bool     `json:"push" yaml:"push" toml:"push" mapstructure:"push"`
	PushArgs                   []string `json:"pushArgs" yaml:"pushArgs" toml:"pushArgs" mapstructure:"pushArgs"`
	PushRepo                   string   `json:"pushRepo" yaml:"pushRepo" toml:"pushRepo" mapstructure:"pushRepo"`
	RequireBranch              string   `json:"requireBranch" yaml:"requireBranch" toml:"requireBranch" mapstructure:"requireBranch"`
	RequireCleanWorkingDir     bool     `json:"requireCleanWorkingDir" yaml:"requireCleanWorkingDir" toml:"requireCleanWorkingDir" mapstructure:"requireCleanWorkingDir"`
	RequireUpstream            bool     `json:"requireUpstream" yaml:"requireUpstream" toml:"requireUpstream" mapstructure:"requireUpstream"`
	RequireCommits             bool     `json:"requireCommits" yaml:"requireCommits" toml:"requireCommits" mapstructure:"requireCommits"`
	Changelog                  string   `json:"changelog" yaml:"changelog" toml:"changelog" mapstructure:"changelog"`
	GetLatestTagFromAllRefs    bool     `json:"getLatestTagFromAllRefs" yaml:"getLatestTagFromAllRefs" toml:"getLatestTagFromAllRefs" mapstructure:"getLatestTagFromAllRefs"`
	CommitsPath                string   `json:"commitsPath" yaml:"commitsPath" toml:"commitsPath" mapstructure:"commitsPath"`
	AddUntrackedFiles          bool     `json:"addUntrackedFiles" yaml:"addUntrackedFiles" toml:"addUntrackedFiles" mapstructure:"addUntrackedFiles"`
	RequireConventionalCommits bool     `json:"requireConventionalCommits" yaml:"requireConventionalCommits" toml:"requireConventionalCommits" mapstructure:"requireConventionalCommits"`
}

// GitHubConfig holds GitHub release configuration.
type GitHubConfig struct {
	Release                bool                 `json:"release" yaml:"release" toml:"release" mapstructure:"release"`
	ReleaseName            string               `json:"releaseName" yaml:"releaseName" toml:"releaseName" mapstructure:"releaseName"`
	ReleaseNotes           string               `json:"releaseNotes" yaml:"releaseNotes" toml:"releaseNotes" mapstructure:"releaseNotes"`
	Draft                  bool                 `json:"draft" yaml:"draft" toml:"draft" mapstructure:"draft"`
	PreRelease             bool                 `json:"preRelease" yaml:"preRelease" toml:"preRelease" mapstructure:"preRelease"`
	MakeLatest             bool                 `json:"makeLatest" yaml:"makeLatest" toml:"makeLatest" mapstructure:"makeLatest"`
	AutoGenerate           bool                 `json:"autoGenerate" yaml:"autoGenerate" toml:"autoGenerate" mapstructure:"autoGenerate"`
	Assets                 []string             `json:"assets" yaml:"assets" toml:"assets" mapstructure:"assets"`
	Host                   string               `json:"host" yaml:"host" toml:"host" mapstructure:"host"`
	TokenRef               string               `json:"tokenRef" yaml:"tokenRef" toml:"tokenRef" mapstructure:"tokenRef"`
	Timeout                int                  `json:"timeout" yaml:"timeout" toml:"timeout" mapstructure:"timeout"`
	Proxy                  string               `json:"proxy" yaml:"proxy" toml:"proxy" mapstructure:"proxy"`
	SkipChecks             bool                 `json:"skipChecks" yaml:"skipChecks" toml:"skipChecks" mapstructure:"skipChecks"`
	Web                    bool                 `json:"web" yaml:"web" toml:"web" mapstructure:"web"`
	Comments               GitHubCommentsConfig `json:"comments" yaml:"comments" toml:"comments" mapstructure:"comments"`
	DiscussionCategoryName string               `json:"discussionCategoryName" yaml:"discussionCategoryName" toml:"discussionCategoryName" mapstructure:"discussionCategoryName"`
}

// GitHubCommentsConfig holds GitHub comment templates.
type GitHubCommentsConfig struct {
	Submit bool   `json:"submit" yaml:"submit" toml:"submit" mapstructure:"submit"`
	Issue  string `json:"issue" yaml:"issue" toml:"issue" mapstructure:"issue"`
	PR     string `json:"pr" yaml:"pr" toml:"pr" mapstructure:"pr"`
}

// GitLabConfig holds GitLab release configuration.
type GitLabConfig struct {
	Release                              bool     `json:"release" yaml:"release" toml:"release" mapstructure:"release"`
	ReleaseName                          string   `json:"releaseName" yaml:"releaseName" toml:"releaseName" mapstructure:"releaseName"`
	ReleaseNotes                         string   `json:"releaseNotes" yaml:"releaseNotes" toml:"releaseNotes" mapstructure:"releaseNotes"`
	PreRelease                           bool     `json:"preRelease" yaml:"preRelease" toml:"preRelease" mapstructure:"preRelease"`
	Milestones                           []string `json:"milestones" yaml:"milestones" toml:"milestones" mapstructure:"milestones"`
	Assets                               []string `json:"assets" yaml:"assets" toml:"assets" mapstructure:"assets"`
	TokenRef                             string   `json:"tokenRef" yaml:"tokenRef" toml:"tokenRef" mapstructure:"tokenRef"`
	TokenHeader                          string   `json:"tokenHeader" yaml:"tokenHeader" toml:"tokenHeader" mapstructure:"tokenHeader"`
	Origin                               string   `json:"origin" yaml:"origin" toml:"origin" mapstructure:"origin"`
	SkipChecks                           bool     `json:"skipChecks" yaml:"skipChecks" toml:"skipChecks" mapstructure:"skipChecks"`
	CertificateAuthorityFile             string   `json:"certificateAuthorityFile" yaml:"certificateAuthorityFile" toml:"certificateAuthorityFile" mapstructure:"certificateAuthorityFile"`
	CertificateAuthorityFileRef          string   `json:"certificateAuthorityFileRef" yaml:"certificateAuthorityFileRef" toml:"certificateAuthorityFileRef" mapstructure:"certificateAuthorityFileRef"`
	Secure                               bool     `json:"secure" yaml:"secure" toml:"secure" mapstructure:"secure"`
	UseGenericPackageRepositoryForAssets bool     `json:"useGenericPackageRepositoryForAssets" yaml:"useGenericPackageRepositoryForAssets" toml:"useGenericPackageRepositoryForAssets" mapstructure:"useGenericPackageRepositoryForAssets"`
}

// HooksConfig holds lifecycle hook commands.
type HooksConfig struct {
	BeforeInit          []string `json:"before:init" yaml:"before:init" toml:"before:init" mapstructure:"before:init"`
	AfterInit           []string `json:"after:init" yaml:"after:init" toml:"after:init" mapstructure:"after:init"`
	BeforeBump          []string `json:"before:bump" yaml:"before:bump" toml:"before:bump" mapstructure:"before:bump"`
	AfterBump           []string `json:"after:bump" yaml:"after:bump" toml:"after:bump" mapstructure:"after:bump"`
	BeforeRelease       []string `json:"before:release" yaml:"before:release" toml:"before:release" mapstructure:"before:release"`
	AfterRelease        []string `json:"after:release" yaml:"after:release" toml:"after:release" mapstructure:"after:release"`
	BeforeGitRelease    []string `json:"before:git:release" yaml:"before:git:release" toml:"before:git:release" mapstructure:"before:git:release"`
	AfterGitRelease     []string `json:"after:git:release" yaml:"after:git:release" toml:"after:git:release" mapstructure:"after:git:release"`
	BeforeGitHubRelease []string `json:"before:github:release" yaml:"before:github:release" toml:"before:github:release" mapstructure:"before:github:release"`
	AfterGitHubRelease  []string `json:"after:github:release" yaml:"after:github:release" toml:"after:github:release" mapstructure:"after:github:release"`
	BeforeGitLabRelease []string `json:"before:gitlab:release" yaml:"before:gitlab:release" toml:"before:gitlab:release" mapstructure:"before:gitlab:release"`
	AfterGitLabRelease  []string `json:"after:gitlab:release" yaml:"after:gitlab:release" toml:"after:gitlab:release" mapstructure:"after:gitlab:release"`

	// Git hooks — installed to .git/hooks/ via `release-it-go install`
	PreCommit        []string `json:"pre-commit" yaml:"pre-commit" toml:"pre-commit" mapstructure:"pre-commit"`
	CommitMsg        []string `json:"commit-msg" yaml:"commit-msg" toml:"commit-msg" mapstructure:"commit-msg"`
	PrePush          []string `json:"pre-push" yaml:"pre-push" toml:"pre-push" mapstructure:"pre-push"`
	PostCommit       []string `json:"post-commit" yaml:"post-commit" toml:"post-commit" mapstructure:"post-commit"`
	PostMerge        []string `json:"post-merge" yaml:"post-merge" toml:"post-merge" mapstructure:"post-merge"`
	PrepareCommitMsg []string `json:"prepare-commit-msg" yaml:"prepare-commit-msg" toml:"prepare-commit-msg" mapstructure:"prepare-commit-msg"`
}

// ChangelogConfig holds changelog generation configuration.
type ChangelogConfig struct {
	Enabled        bool   `json:"enabled" yaml:"enabled" toml:"enabled" mapstructure:"enabled"`
	Preset         string `json:"preset" yaml:"preset" toml:"preset" mapstructure:"preset"`
	Infile         string `json:"infile" yaml:"infile" toml:"infile" mapstructure:"infile"`
	Header         string `json:"header" yaml:"header" toml:"header" mapstructure:"header"`
	KeepAChangelog bool   `json:"keepAChangelog" yaml:"keepAChangelog" toml:"keepAChangelog" mapstructure:"keepAChangelog"`
	AddUnreleased  bool   `json:"addUnreleased" yaml:"addUnreleased" toml:"addUnreleased" mapstructure:"addUnreleased"`
	KeepUnreleased bool   `json:"keepUnreleased" yaml:"keepUnreleased" toml:"keepUnreleased" mapstructure:"keepUnreleased"`
	AddVersionUrl  bool   `json:"addVersionUrl" yaml:"addVersionUrl" toml:"addVersionUrl" mapstructure:"addVersionUrl"`
}

// BumperConfig holds multi-file version bumping configuration.
type BumperConfig struct {
	Enabled bool         `json:"enabled" yaml:"enabled" toml:"enabled" mapstructure:"enabled"`
	In      *BumperFile  `json:"in" yaml:"in" toml:"in" mapstructure:"in"`
	Out     []BumperFile `json:"out" yaml:"out" toml:"out" mapstructure:"out"`
}

// BumperFile represents a file for version reading/writing.
type BumperFile struct {
	File             string `json:"file" yaml:"file" toml:"file" mapstructure:"file"`
	Path             string `json:"path" yaml:"path" toml:"path" mapstructure:"path"`
	Type             string `json:"type" yaml:"type" toml:"type" mapstructure:"type"`
	Prefix           string `json:"prefix" yaml:"prefix" toml:"prefix" mapstructure:"prefix"`
	VersionPrefix    string `json:"versionPrefix" yaml:"versionPrefix" toml:"versionPrefix" mapstructure:"versionPrefix"`
	ConsumeWholeFile bool   `json:"consumeWholeFile" yaml:"consumeWholeFile" toml:"consumeWholeFile" mapstructure:"consumeWholeFile"`
}

// CalVerConfig holds calendar versioning configuration.
type CalVerConfig struct {
	Enabled           bool   `json:"enabled" yaml:"enabled" toml:"enabled" mapstructure:"enabled"`
	Format            string `json:"format" yaml:"format" toml:"format" mapstructure:"format"`
	Increment         string `json:"increment" yaml:"increment" toml:"increment" mapstructure:"increment"`
	FallbackIncrement string `json:"fallbackIncrement" yaml:"fallbackIncrement" toml:"fallbackIncrement" mapstructure:"fallbackIncrement"`
}

// NotificationConfig holds webhook notification configuration.
type NotificationConfig struct {
	Enabled  bool            `json:"enabled" yaml:"enabled" toml:"enabled" mapstructure:"enabled"`
	Webhooks []WebhookConfig `json:"webhooks" yaml:"webhooks" toml:"webhooks" mapstructure:"webhooks"`
}

// WebhookConfig holds a single webhook endpoint configuration.
type WebhookConfig struct {
	Type                string   `json:"type" yaml:"type" toml:"type" mapstructure:"type"`
	URLRef              string   `json:"urlRef" yaml:"urlRef" toml:"urlRef" mapstructure:"urlRef"`
	MessageTemplate     string   `json:"messageTemplate" yaml:"messageTemplate" toml:"messageTemplate" mapstructure:"messageTemplate"`
	Timeout             int      `json:"timeout" yaml:"timeout" toml:"timeout" mapstructure:"timeout"`
	ThemeColor          string   `json:"themeColor" yaml:"themeColor" toml:"themeColor" mapstructure:"themeColor"`
	ImageURL            string   `json:"imageUrl" yaml:"imageUrl" toml:"imageUrl" mapstructure:"imageUrl"`
	IgnoredContributors []string `json:"ignoredContributors" yaml:"ignoredContributors" toml:"ignoredContributors" mapstructure:"ignoredContributors"`
}
