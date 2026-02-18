# release-it-go

Automate versioning and release workflows — **without npm or Node.js**.

`release-it-go` is a Go rewrite of [release-it](https://github.com/release-it/release-it). It handles Git tagging, GitHub/GitLab releases, changelog generation, multi-file version bumping, and webhook notifications from a single binary.

## Features

- **Zero dependencies** — single Go binary, no npm/Node.js required
- **Git automation** — commit, tag, push with configurable templates
- **GitHub & GitLab releases** — create releases, upload assets, post comments
- **Conventional Commits** — parse, validate, and generate changelogs from commit history
- **Keep a Changelog** — alternative changelog format support
- **Multi-file version bumping** — update version in JSON, YAML, TOML, INI, or plain text files
- **Calendar Versioning (CalVer)** — `yy.mm.minor` and custom formats
- **Lifecycle hooks** — run shell commands before/after each step
- **Webhook notifications** — Slack and Microsoft Teams
- **Interactive prompts** — colorful terminal UI with spinners
- **CI/CD ready** — auto-detects CI environments, non-interactive mode
- **Backward compatible** — reads and migrates npm release-it config files
- **Config formats** — JSON, YAML, and TOML

## Quick Start

### Install

```bash
go install release-it-go/cmd/release-it-go@latest
```

Or download a binary from [GitHub Releases](https://github.com/user/release-it-go/releases).

### Initialize

```bash
release-it-go init
```

This starts an interactive wizard that creates a `.release-it-go.json` or `.release-it-go.yaml` config file.

To generate a full reference config with all options documented:

```bash
release-it-go init --full-example
```

### Release

```bash
release-it-go
```

That's it. The tool detects the latest version from Git tags, determines the next version from your commits, and runs the full release pipeline.

## CLI Usage

```
release-it-go [flags]
release-it-go [command]
```

### Commands

| Command | Description |
|---------|-------------|
| `init` | Interactive config setup wizard |
| `init --full-example` | Generate `.release-it-go-full.yaml` with all options documented |
| `version` | Print version, commit hash, and build date |
| `completion <shell>` | Generate shell completions (`bash`, `zsh`, `fish`, `powershell`) |

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--config <path>` | `-c` | Path to config file |
| `--dry-run` | `-d` | Preview all actions without making changes |
| `--ci` | | Non-interactive mode (auto-confirms all prompts) |
| `--verbose` | `-V` | Verbose output (`-V` verbose, `-VV` debug) |
| `--increment <type>` | `-i` | Version increment: `major`, `minor`, `patch`, `premajor`, `preminor`, `prepatch`, `prerelease` |
| `--preReleaseId <id>` | | Pre-release identifier (e.g. `beta`, `alpha`, `rc`) |
| `--preRelease <id>` | | Shorthand: sets pre-release ID and marks releases accordingly |
| `--changelog` | | Generate and print changelog only |
| `--release-version` | | Print next version only |
| `--only-version` | | Prompt for version, then run remaining steps non-interactively |
| `--no-increment` | | Run pipeline without incrementing version |
| `--no-git.commit` | | Skip git commit step |
| `--no-git.tag` | | Skip git tag step |
| `--no-git.push` | | Skip git push step |
| `--check-commits` | | Validate conventional commits only (no release) |
| `--ignore-commit-lint` | | Skip conventional commit validation |

### Examples

```bash
# Standard release (auto-detect increment from commits)
release-it-go

# Specific version bump
release-it-go -i minor

# Pre-release
release-it-go --preRelease beta

# Dry run to preview
release-it-go --dry-run

# CI/CD pipeline
release-it-go --ci

# Only generate changelog
release-it-go --changelog

# Check what version would be next
release-it-go --release-version
```

## Configuration

### Config Files

`release-it-go` searches for config files in the following order:

| Priority | File | Format |
|----------|------|--------|
| 1 | `.release-it-go.json` | JSON |
| 2 | `.release-it-go.yaml` | YAML |
| 3 | `.release-it-go.yml` | YAML |
| 4 | `.release-it-go.toml` | TOML |
| 5 | `.release-it.json` | JSON (legacy) |
| 6 | `.release-it.yaml` | YAML (legacy) |
| 7 | `.release-it.yml` | YAML (legacy) |
| 8 | `.release-it.toml` | TOML (legacy) |

Legacy `.release-it.*` files are auto-detected and migration is offered during `init`.

### Minimal Config

The config file only needs to contain values that differ from the defaults. Here's a minimal example:

```yaml
# .release-it-go.yaml
github:
  release: true
```

```json
{
  "github": {
    "release": true
  }
}
```

---

## Configuration Reference

### Git

Controls commit, tag, and push behavior.

```yaml
git:
  commit: true                          # Create a git commit (default: true)
  commitMessage: "chore: release v${version}"  # Commit message template
  commitArgs: []                        # Extra git commit arguments
  tag: true                             # Create a git tag (default: true)
  tagName: "${version}"                 # Tag name template (default: "${version}")
  tagMatch: ""                          # Glob to match tags for version detection
  tagExclude: ""                        # Glob to exclude tags (e.g. "*-rc.*")
  tagAnnotation: "Release ${version}"   # Annotation for annotated tags
  tagArgs: []                           # Extra git tag arguments
  push: true                            # Push to remote (default: true)
  pushArgs: ["--follow-tags"]           # Extra push arguments
  pushRepo: "origin"                    # Remote name (default: "origin")
  requireBranch: ""                     # Required branch (empty = any)
  requireCleanWorkingDir: true          # Abort if working directory is dirty
  requireUpstream: true                 # Require upstream tracking branch
  requireCommits: true                  # Require new commits since last tag
  requireConventionalCommits: true      # Require conventional commit format
  getLatestTagFromAllRefs: false        # Search all refs for latest tag
  addUntrackedFiles: false              # Stage untracked files before commit
```

**Template variables:** `${version}`, `${latestTag}`, `${tagName}`

### GitHub

Create GitHub releases, upload assets, and comment on issues/PRs.

```yaml
github:
  release: false                        # Create a GitHub release
  releaseName: "Release ${version}"     # Release title
  releaseNotes: ""                      # Custom release notes template
  draft: false                          # Create as draft
  preRelease: false                     # Mark as pre-release
  makeLatest: true                      # Mark as latest release (default: true)
  autoGenerate: false                   # Auto-generate notes via GitHub API
  assets: []                            # Glob patterns for assets to upload
  host: "api.github.com"               # API host (change for Enterprise)
  tokenRef: "GITHUB_TOKEN"             # Env var with GitHub token
  timeout: 0                            # API timeout in seconds (0 = default)
  proxy: ""                             # HTTP proxy URL
  skipChecks: false                     # Skip API pre-flight checks
  web: false                            # Open release URL in browser
  comments:
    submit: false                       # Post comments on resolved issues/PRs
    issue: ":rocket: _This issue has been resolved in v${version}._"
    pr: ":rocket: _This pull request is included in v${version}._"
  discussionCategoryName: ""            # GitHub Discussions category
```

**Authentication:** Set `GITHUB_TOKEN` environment variable (or use a custom name via `tokenRef`).

### GitLab

Create GitLab releases with milestone and asset support.

```yaml
gitlab:
  release: false                        # Create a GitLab release
  releaseName: "Release ${version}"     # Release title
  releaseNotes: ""                      # Custom release notes template
  preRelease: false                     # Mark as upcoming release
  milestones: []                        # Associated milestone titles
  assets: []                            # Release asset links
  tokenRef: "GITLAB_TOKEN"             # Env var with GitLab token
  tokenHeader: "Private-Token"          # Auth header name
  origin: ""                            # GitLab URL (for self-hosted)
  skipChecks: false                     # Skip API pre-flight checks
  certificateAuthorityFile: ""          # CA certificate file path
  secure: false                         # Use HTTPS for API calls
```

**Authentication:** Set `GITLAB_TOKEN` environment variable (or use a custom name via `tokenRef`).

### Changelog

Generate changelogs from conventional commits.

```yaml
changelog:
  enabled: true                         # Enable changelog generation (default: true)
  preset: "angular"                     # Format preset: "angular"
  infile: "CHANGELOG.md"               # Output file (empty string = don't write file)
  header: "# Changelog"                # File header
  keepAChangelog: false                 # Use Keep a Changelog format
  addUnreleased: false                  # Add [Unreleased] section
  keepUnreleased: false                 # Keep [Unreleased] after release
  addVersionUrl: false                  # Add compare URLs for versions
```

**Conventional Changelog** groups commits by type:

```
## [1.2.0](https://github.com/owner/repo/compare/v1.1.0...v1.2.0) (2026-02-17)

### Features

* **auth:** implement JWT authentication (abc1234)

### Bug Fixes

* **api:** fix timeout handling (def5678)

### BREAKING CHANGES

* **api:** removed deprecated /v1 endpoints
```

**Keep a Changelog** uses semantic sections:

```
## [1.2.0] - 2026-02-17

### Added
- JWT authentication

### Fixed
- API timeout handling
```

To disable changelog file writing but still generate release notes, set `infile: ""`.

### Hooks

Run shell commands at specific points in the release lifecycle.

```yaml
hooks:
  "before:init": []
  "after:init": []
  "before:bump": []
  "after:bump": ["echo 'Bumped to v${version}'"]
  "before:release": []
  "after:release": ["echo 'Released v${version}'"]
  "before:git:release": []
  "after:git:release": []
  "before:github:release": []
  "after:github:release": []
  "before:gitlab:release": []
  "after:gitlab:release": []
```

Each hook accepts an array of shell commands. Commands are executed sequentially via `sh -c` and support template variables.

**Available variables:** `${version}`, `${latestVersion}`, `${tagName}`, `${changelog}`, `${releaseUrl}`, `${repo.owner}`, `${repo.repository}`, `${repo.remote}`

### Bumper

Update version strings across multiple files simultaneously.

```yaml
bumper:
  enabled: false
  in:                                   # Source file to read current version
    file: "VERSION"
    consumeWholeFile: true
  out:                                  # Target files to write new version
    - file: "VERSION"
      consumeWholeFile: true
    - file: "package.json"
      path: "version"
    - file: "Cargo.toml"
      path: "package.version"
```

**BumperFile options:**

| Field | Description |
|-------|-------------|
| `file` | File path (supports glob patterns in `out`) |
| `path` | Dot-separated path for JSON/YAML/TOML files (e.g. `package.version`) |
| `type` | File format: `json`, `yaml`, `toml`, `ini`, `text` (auto-detected from extension) |
| `prefix` | Text prefix before version string |
| `versionPrefix` | Version prefix (e.g. `v`) |
| `consumeWholeFile` | Treat entire file content as the version string |

### CalVer

Use Calendar Versioning instead of Semantic Versioning.

```yaml
calver:
  enabled: false                        # Enable CalVer (disables SemVer)
  format: "yy.mm.minor"               # Version format
  increment: "calendar"                 # Increment strategy
  fallbackIncrement: "minor"            # Fallback when calendar hasn't changed
```

When enabled, versions follow the calendar format (e.g. `26.02.0`, `26.02.1`). The minor component resets when the year or month changes.

### Notifications

Send webhook notifications after a successful release.

```yaml
notification:
  enabled: false
  webhooks:
    - type: "slack"                     # "slack" or "teams"
      urlRef: "SLACK_WEBHOOK_URL"       # Env var containing the webhook URL
      messageTemplate: ""               # Custom message (empty = default)
      timeout: 0                        # Timeout in seconds (0 = 30s default)
    - type: "teams"
      urlRef: "TEAMS_WEBHOOK_URL"
```

**Default messages:**

- **Slack:** `🚀 *${repo.repository}* v${version} released!\n${releaseUrl}`
- **Teams:** `🚀 ${repo.repository} v${version} released!\n${releaseUrl}`

Notification failures are non-fatal — they log a warning but don't stop the release.

---

## Release Pipeline

When you run `release-it-go`, the following steps execute in order:

```
1. init             → Detect repo, resolve variables
2. prerequisites    → Check branch, clean dir, upstream, commits
3. commitlint       → Validate conventional commits
4. version          → Determine next version (prompt or auto)
5. bump             → Update version in configured files
6. changelog        → Generate/update CHANGELOG.md
7. git:release      → Commit, tag, push
8. github:release   → Create GitHub release, upload assets
9. gitlab:release   → Create GitLab release
10. notification    → Send Slack/Teams webhooks
```

Each step fires `before:<step>` and `after:<step>` hooks. In dry-run mode, all actions are logged but not executed.

## Version Detection

Versions are detected in this order:

1. **Git tags** (primary) — latest tag matching `tagMatch` pattern
2. **VERSION file** (secondary) — fallback if no tags found
3. **0.0.0** — fallback for brand-new repositories

### Increment Types

| Type | Example (from 1.2.3) | Description |
|------|----------------------|-------------|
| `patch` | 1.2.4 | Bug fixes |
| `minor` | 1.3.0 | New features |
| `major` | 2.0.0 | Breaking changes |
| `prepatch` | 1.2.4-beta.0 | Pre-release patch |
| `preminor` | 1.3.0-beta.0 | Pre-release minor |
| `premajor` | 2.0.0-beta.0 | Pre-release major |
| `prerelease` | 1.2.4-beta.1 | Increment pre-release number |

When using conventional commits, the increment is auto-detected:

- `feat:` → minor
- `fix:`, `perf:`, `revert:` → patch
- `BREAKING CHANGE` footer or `!` suffix → major

## Conventional Commits

`release-it-go` parses commits following the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
type(scope): description

optional body

optional footer(s)
```

**Supported types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`

**Breaking changes** are detected from:
- `!` after type/scope: `feat!: remove deprecated API`
- `BREAKING CHANGE:` footer in commit body

By default, `requireConventionalCommits` is `true`. This can be disabled in config or bypassed with `--ignore-commit-lint`.

### Commit Linting

Use `--check-commits` to validate commits without running a release. This is useful in CI pipelines or pre-merge checks:

```bash
release-it-go --check-commits
```

The command inspects all commits since the latest tag and reports which ones follow the conventional format:

```
# All valid
All 5 commits are conventional. ✓

# Some invalid (exit code 1)
3 of 5 commits are not conventional:
  ✗ abc1234 oops forgot the type ← missing type prefix
  ✗ def5678 Update readme ← missing type prefix
  ✗ ghi9012 wip ← missing type prefix
```

With `-V` (verbose), every commit is listed with its pass/fail status:

```bash
release-it-go --check-commits -V
  ✓ 1a2b3c4 feat: add user authentication
  ✓ 5e6f7g8 fix: resolve login timeout
  ✗ 9h0i1j2 update docs ← missing type prefix
```

To skip commit linting during a release, use `--ignore-commit-lint`.

## Migration from npm release-it

If you have an existing `.release-it.json` config from the npm package, `release-it-go` automatically detects it during `init` and offers migration:

```bash
release-it-go init
# → Detected legacy .release-it.json. Migrate to release-it-go format? (Y/n)
```

Migration:
1. Creates a `.release-it.json.bak` backup
2. Normalizes npm-specific fields and plugin settings
3. Maps `@release-it/conventional-changelog` and `@release-it/keep-a-changelog` plugin configs
4. Writes a clean `.release-it-go.json` or `.release-it-go.yaml`

## CI/CD Integration

### Auto-Detection

`release-it-go` detects CI environments automatically from environment variables: `CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, `CIRCLECI`, `TRAVIS`, `JENKINS_URL`, `BITBUCKET_BUILD_NUMBER`, and more. When no TTY is available, CI mode is also auto-enabled.

### GitHub Actions

```yaml
name: Release
on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22+"
      - run: go install release-it-go/cmd/release-it-go@latest
      - run: release-it-go --ci
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### GitLab CI

```yaml
release:
  stage: deploy
  only:
    - main
  script:
    - go install release-it-go/cmd/release-it-go@latest
    - release-it-go --ci
  variables:
    GITLAB_TOKEN: $GITLAB_TOKEN
```

## Template Variables

The following variables can be used in commit messages, tag names, release names, hooks, and notification templates:

| Variable | Description | Example |
|----------|-------------|---------|
| `${version}` | New version number | `1.2.0` |
| `${latestVersion}` | Current/previous version | `1.1.0` |
| `${tagName}` | Full tag name | `v1.2.0` |
| `${latestTag}` | Previous tag name | `v1.1.0` |
| `${changelog}` | Generated changelog text | |
| `${releaseUrl}` | URL of the created release | |
| `${repo.owner}` | Repository owner | `octocat` |
| `${repo.repository}` | Repository name | `my-project` |
| `${repo.remote}` | Remote URL | `origin` |

## Shell Completions

```bash
# Bash
release-it-go completion bash > /etc/bash_completion.d/release-it-go

# Zsh
release-it-go completion zsh > "${fpath[1]}/_release-it-go"

# Fish
release-it-go completion fish > ~/.config/fish/completions/release-it-go.fish

# PowerShell
release-it-go completion powershell | Out-String | Invoke-Expression
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | GitHub API token (configurable via `github.tokenRef`) |
| `GITLAB_TOKEN` | GitLab API token (configurable via `gitlab.tokenRef`) |
| `SLACK_WEBHOOK_URL` | Slack webhook URL (configurable via `notification.webhooks[].urlRef`) |
| `TEAMS_WEBHOOK_URL` | Teams webhook URL (configurable via `notification.webhooks[].urlRef`) |
| `CI` | Enables CI mode when set |
| `NO_COLOR` | Disables colored output |

## License

MIT
