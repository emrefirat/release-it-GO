# release-it-go

Automate versioning and release workflows — **without npm or Node.js**.

`release-it-go` is a Go rewrite of [release-it](https://github.com/release-it/release-it). It handles Git tagging, GitHub/GitLab releases, changelog generation, multi-file version bumping, and webhook notifications from a single binary.

## Features

- **Zero dependencies** — single Go binary, no npm/Node.js required
- **Git automation** — commit, tag, push with configurable templates
- **GitHub & GitLab releases** — create releases, upload assets
- **Conventional Commits** — parse, validate, and generate changelogs from commit history
- **Keep a Changelog** — alternative changelog format support
- **Multi-file version bumping** — update version in JSON, YAML, TOML, INI, or plain text files
- **Calendar Versioning (CalVer)** — `yy.mm.minor` and custom formats
- **Lifecycle hooks** — run shell commands before/after each step, with `RELEASE_*` environment variables
- **Git hooks** — install `pre-commit` / `commit-msg` / `pre-push` scripts from the same config (`hooks install`)
- **Strict configuration** — unknown keys and invalid values are rejected before anything runs, with `did you mean` hints
- **Webhook notifications** — Slack and Microsoft Teams
- **Interactive prompts** — colorful terminal UI with spinners
- **CI/CD ready** — auto-detects CI environments, non-interactive mode
- **Backward compatible** — reads and migrates npm release-it config files
- **Config formats** — JSON, YAML, and TOML

## Quick Start

### Install

```bash
go install github.com/emrefirat/release-it-GO/cmd/release-it-go@latest
```

Or download a binary from [GitHub Releases](https://github.com/emrefirat/release-it-GO/releases).

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
release-it-go [increment | version] [flags]
release-it-go [command]
```

The optional positional argument is either an increment keyword (`release-it-go minor`, same as `-i minor`) or an exact version (`release-it-go 2.0.0`), npm release-it style. An exact version must be greater than the latest released one.

### Commands

| Command | Description |
|---------|-------------|
| `init` | Interactive config setup wizard |
| `init --full-example` | Generate `.release-it-go-full.yaml` with all options documented |
| `hooks install [--force]` | Write the git hooks from the `hooks` section to `.hooks/` and set `core.hooksPath` (see [Git Hooks](#git-hooks)) |
| `hooks remove` | Delete the managed hook scripts and reset `core.hooksPath` |
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
| `--check-commits` | | Validate the commits since the latest tag (no release) |
| `--check-msg <msg\|file\|->` | | Validate a single commit message: a string, a file path (what the `commit-msg` hook passes), or `-` for stdin |
| `--ignore-commit-lint` | | Skip conventional commit validation |

### Examples

```bash
# Standard release (auto-detect increment from commits)
release-it-go

# Specific version bump (flag or positional)
release-it-go -i minor
release-it-go minor

# Exact version
release-it-go 2.0.0

# Validate a commit message (exit 1 when invalid)
release-it-go --check-msg "feat: add login"

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

### Validation

Config files are decoded strictly: a key the tool does not know is an error, with a suggestion when it looks like a typo:

```
unknown config key "github.relase" (did you mean "release"?); unknown config key "hooks.precommit" (did you mean "pre-commit"?)
```

Values are checked before the pipeline starts (`invalid configuration:` lists every problem): `git.tagName` must contain `${version}`, `increment` must be a keyword or a semver version, `preReleaseId` must be a valid identifier, `calver.format` must be a supported format, `github.host` takes no scheme while `gitlab.origin` requires one, timeouts cannot be negative, webhook and bumper `type`s must be known, `changelog.preset` must be `angular` or `conventionalcommits`. Keys from npm release-it that have no counterpart (`npm`, `plugins`, `versionFile`, `changelogFile`) are translated or ignored; keys this tool once accepted but never acted on print a `config: ignored "..."` warning and keep loading.

### General

Top-level keys mirror the global flags. A flag given on the command line always wins; a key in the file is used when the flag is absent.

```yaml
increment: ""        # "major" | "minor" | "patch" | "pre*" | exact version | "" = auto-detect from commits
preReleaseId: ""     # e.g. "beta" — pre-release identifier
ci: false            # non-interactive mode
dry-run: false       # log every action, change nothing
verbose: 0           # 1 = verbose, 2 = debug
```

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
  tagName: "${version}"                 # Tag name template; only ${version} is allowed here
  tagMatch: ""                          # Glob to match tags for version detection
  tagExclude: ""                        # Glob to exclude tags (e.g. "*-rc.*")
  tagAnnotation: "Release ${version}"   # Annotation for annotated tags
  tagArgs: []                           # Extra git tag arguments
  push: true                            # Push to remote (default: true)
  pushArgs: ["--follow-tags", "--atomic"]  # Extra push arguments (atomic prevents orphan tags)
  pushRepo: "origin"                    # Remote name (default: "origin")
  requireBranch: ""                     # Required branch: glob or comma list ("main,release/*"); empty = any
  requireCleanWorkingDir: true          # Abort if working directory is dirty
  requireUpstream: true                 # Require upstream tracking branch
  requireCommits: true                  # Require new commits since last tag
  commitsPath: ""                       # Only count commits touching this path (monorepo)
  requireConventionalCommits: true      # Require conventional commit format
  getLatestTagFromAllRefs: false        # Search all refs for latest tag
  addUntrackedFiles: false              # false = git add . --update (tracked changes); true = git add . (new files too)
```

**Template variables:** `commitMessage` and `tagAnnotation` accept every variable from [Template Variables](#template-variables). `tagName` accepts only `${version}` — the template must be reversible so the previous release can be found again (`release-${version}` works, `${version}-${branchName}` does not).

**Tag prefix inference:** with the default `tagName`, a repository whose latest tag is `v1.4.0` keeps the `v` (the next tag is `v1.5.0`) even though the template says `${version}`. Writing `tagName` in the config file — either form — disables the inference.

**Staging:** the release commit stages every tracked modification made during the run (bumper outputs, changelog, files rewritten by hooks), i.e. `git add . --update`. With `addUntrackedFiles: true` new files are included as well (`git add .`).

### GitHub

Create GitHub releases and upload assets.

```yaml
github:
  release: false                        # Create a GitHub release
  releaseName: "Release ${version}"     # Release title
  draft: false                          # Create as draft
  preRelease: false                     # Mark as pre-release
  makeLatest: true                      # Mark as latest release (default: true)
  autoGenerate: false                   # Auto-generate notes via GitHub API
  assets: []                            # Glob patterns for assets to upload
  host: "github.com"                   # GitHub host (change for Enterprise)
  tokenRef: "GITHUB_TOKEN"             # Env var with GitHub token
  timeout: 0                            # API timeout in seconds (0 = default)
  proxy: ""                             # HTTP proxy URL
  skipChecks: false                     # Skip API pre-flight checks
  discussionCategoryName: ""            # GitHub Discussions category
```

**Authentication:** Set `GITHUB_TOKEN` environment variable (or use a custom name via `tokenRef`).

### GitLab

Create GitLab releases with milestone and asset support.

```yaml
gitlab:
  release: false                        # Create a GitLab release
  releaseName: "Release ${version}"     # Release title
  milestones: []                        # Associated milestone titles
  assets: []                            # Release asset links
  useGenericPackageRepositoryForAssets: true  # false = project uploads API (no Package Registry)
  tokenRef: "GITLAB_TOKEN"             # Env var with GitLab token
  tokenHeader: "Private-Token"          # Auth header name
  origin: ""                            # GitLab URL (for self-hosted)
  skipChecks: false                     # Skip API pre-flight checks
  certificateAuthorityFile: ""          # CA certificate file path (self-signed instances)
  certificateAuthorityFileRef: ""       # Env var holding that path instead
  secure: true                          # Verify TLS certificates (false = explicit opt-out)
```

**Authentication:** Set `GITLAB_TOKEN` environment variable (or use a custom name via `tokenRef`).

### Network behavior (GitHub, GitLab, webhooks)

- `HTTPS_PROXY` / `NO_PROXY` are honored; `github.proxy` overrides them for GitHub.
- Transient responses (`429`, `502`, `503`, `504`) are retried up to 3 attempts with exponential backoff, honoring `Retry-After`. A request that creates something (POST) is replayed only after `429`/`503` — statuses that mean "not processed" — never after a gateway `502`/`504`, so a release cannot be created twice. Connection errors are retried only for GET requests.
- TLS verification is on by default for GitLab; `secure: false` is an explicit opt-out for self-signed instances without a CA file.

### Changelog

Generate changelogs from conventional commits.

```yaml
changelog:
  enabled: true                         # Enable changelog generation (default: true)
  preset: "angular"                     # Format preset: "angular"
  infile: "CHANGELOG.md"               # Output file (empty string = don't write file)
  header: "# Changelog"                # File header
  keepAChangelog: false                 # Use Keep a Changelog format
  addVersionUrl: true                   # Compare links on version headings
```

**Conventional Changelog** groups commits by type:

```
## [2.0.0](https://github.com/owner/repo/compare/v1.1.0...v2.0.0) (2026-09-02)

### Features

* **api:** remove deprecated /v1 endpoints ([38059b7](https://github.com/owner/repo/commit/38059b7))
* **auth:** implement JWT authentication ([4ebc075](https://github.com/owner/repo/commit/4ebc075))

### Bug Fixes

* **api:** fix timeout handling ([f9ac79f](https://github.com/owner/repo/commit/f9ac79f))

### BREAKING CHANGES

* **api:** the /v1 endpoints are gone, use /v2
```

The compare link on the heading is controlled by `addVersionUrl`; commit links need a recognized `origin` remote. Sections are only written when there are new commits — a run without commits leaves the file untouched.

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

Run shell commands at specific points in the release lifecycle. Every pipeline step has a `before:` and an `after:` hook, plus two release-spanning events:

```yaml
hooks:
  "before:init": []
  "after:init": []
  "before:prerequisites": []
  "after:prerequisites": []
  "before:commitlint": []
  "after:commitlint": []
  "before:version": []
  "after:version": []
  "before:bump": []
  "after:bump": ["echo 'Bumped to v${version}'"]
  "before:changelog": []
  "after:changelog": []
  "before:release": []                   # once, right before git:release
  "before:git:release": []
  "after:git:release": []
  "before:github:release": []
  "after:github:release": []
  "before:gitlab:release": []
  "after:gitlab:release": []
  "before:notification": []
  "after:notification": []
  "after:release": ["echo 'Released v${version}'"]   # once, after every step
```

Each hook is an array of shell commands, run sequentially; a failing command aborts the release. On Unix the command runs through `sh -c`, on Windows through `%COMSPEC% /C` (`cmd.exe`). When the run stops because there are no new commits, the remaining hooks — including `after:release` — do not fire.

Commands may use `${variable}` placeholders, and every variable is also exported as an environment variable — `${version}` → `RELEASE_VERSION`, `${repo.owner}` → `RELEASE_REPO_OWNER` — which avoids quoting problems in scripts:

```yaml
hooks:
  "after:git:release": ["./scripts/publish.sh"]   # reads $RELEASE_VERSION, $RELEASE_TAG_NAME, ...
```

See [Template Variables](#template-variables) for the full list.

**Key naming:** lifecycle hooks are `before:<step>` / `after:<step>` (quote them in YAML); git hooks use the git names in kebab-case (`pre-commit`, `commit-msg`, `pre-push`, `post-commit`, `post-merge`, `prepare-commit-msg`). Unknown or wrong-case keys (`preCommit`) are rejected at load time.

### Git Hooks

The same `hooks` section can hold git hooks. `release-it-go hooks install` writes them as scripts to `.hooks/` in the project root and sets `core.hooksPath` to that directory, so the hooks can be committed and shared with the team (husky style):

```yaml
hooks:
  "pre-commit": ["go fmt ./...", "go vet ./..."]
  "commit-msg": ['release-it-go --check-msg "$1"']   # $1 = the message file git passes in
  "pre-push": ["go test ./..."]
```

```bash
release-it-go hooks install           # writes .hooks/<name>, sets core.hooksPath=.hooks
release-it-go hooks install --force   # overwrite hooks in .hooks/ that were not written by release-it-go
release-it-go hooks remove            # deletes the managed scripts, resets core.hooksPath when none remain
release-it-go hooks install --dry-run # report only
```

`install` reconciles: a hook removed from the config is deleted on the next install, scripts that were not written by release-it-go (no managed header) are never touched, and `.hooks/` is not created when nothing is configured.

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
| `consumeWholeFile` | Treat entire file content as the version string |

Structured files keep their formatting: only the version value is replaced, so key order, indentation, comments and a same-looking dependency version elsewhere in the file survive. Plain-text targets have the current version replaced in place (with `prefix` when given); if the current version cannot be found, the release fails before anything is committed instead of overwriting the file. `in` reads the current version from a file; it is the fallback when the repository has no tag yet.

### CalVer

Use Calendar Versioning instead of Semantic Versioning.

```yaml
calver:
  enabled: false                        # Enable CalVer (disables SemVer)
  format: "yy.mm.minor"               # "yy.mm.minor" | "yyyy.mm.minor" | "yyyy.mm.dd"
```

When enabled, versions follow the calendar (in September 2026: `26.9.0`, then `26.9.1`; the month is not zero-padded). The minor component resets when the year or month changes. CalVer cannot be combined with SemVer pre-releases.

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

Notification failures are non-fatal — they log a warning but don't stop the release. Transient webhook errors are retried (see [Network behavior](#network-behavior-github-gitlab-webhooks)); error messages never include the webhook URL, since it embeds a secret.

## Docker

The image is built locally (`make docker-build`); it contains the static binary, git and CA certificates and runs as a non-root user. The entrypoint requires a git identity for commits:

```bash
docker run --rm \
  -v "$(pwd):/workspace" \
  -e GIT_USER_NAME="Release Bot" -e GIT_USER_EMAIL="bot@example.com" \
  -e GITHUB_TOKEN \
  release-it-go:latest --ci
```

Only `version`, `completion` and `--help` skip the identity check; every other invocation (including `--dry-run` and `--changelog`) needs the two variables.

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

Each step fires `before:<step>` and `after:<step>` hooks; `before:release` fires once ahead of `git:release` and `after:release` once at the end. In dry-run mode, all actions are logged but not executed.

In interactive mode the commit, tag and push of step 7 are confirmed independently — declining the commit does not cancel the tag or the push prompts. `--only-version` asks for the version and runs everything else non-interactively. When there are no commits since the latest tag the run stops after the prerequisites with an informational message (`git.requireCommits`).

## Version Detection

Versions are detected in this order:

1. **Git tags** (primary) — the highest version among tags reachable from `HEAD` that match `tagMatch` (glob, e.g. `v1.*`) and not `tagExclude`; chosen by semver order, so `v1.2.0-rc.1` never shadows `v1.2.0`. `getLatestTagFromAllRefs: true` looks at every ref instead of only merged tags.
2. **`bumper.in` file** (secondary) — used when no tag exists
3. **0.0.0** — fallback for brand-new repositories

The next version is then, in order of precedence: an exact version given on the command line (`release-it-go 2.0.0`, must be greater than the latest), an increment keyword (`-i minor`, positional, or the `increment` config key), or the increment auto-detected from conventional commits.

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

The command inspects all commits since the latest tag and reports the ones that do not follow the format (exit code 1):

```
Error: Commit lint failed:
  9728b9b    Fixed the thing                          ← not in conventional commit format

  1 of 2 commits are not conventional.
  Use --ignore-commit-lint to bypass.
```

With `-V` (verbose), every commit is listed with its pass/fail status:

```
  ✓ 47adb0b feat: add login
  ✗ 9728b9b Fixed the thing ← not in conventional commit format
```

`fixup!`, `squash!`, `amend!`, merge and revert commits are always accepted. To skip commit linting during a release, use `--ignore-commit-lint`.

### Checking a Single Message

`--check-msg` validates one message — a string, a file path (what git passes to a `commit-msg` hook), or `-` for stdin — and explains what is wrong:

```
$ release-it-go --check-msg "fic: deneme"

✗ Invalid commit message

  message:   fic: deneme
  problem:   unknown type "fic" — did you mean "fix"?

  Expected:  <type>(<scope>): <description>   scope is optional
  Example:   fix: deneme
  Types:     feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert

  Run with -V for type descriptions and rules.
```

Wire it as a git hook with `hooks install` (see [Git Hooks](#git-hooks)).

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
      - run: go install github.com/emrefirat/release-it-GO/cmd/release-it-go@latest
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
    - go install github.com/emrefirat/release-it-GO/cmd/release-it-go@latest
    - release-it-go --ci
  variables:
    GITLAB_TOKEN: $GITLAB_TOKEN
```

## Template Variables

The following variables can be used in `git.commitMessage`, `git.tagAnnotation`, `github.releaseName`, `gitlab.releaseName`, hook commands and notification templates (`git.tagName` accepts only `${version}`). Hooks additionally receive each variable as an environment variable:

| Variable | Environment variable | Description | Example |
|----------|----------------------|-------------|---------|
| `${version}` | `RELEASE_VERSION` | New version number | `1.2.0` |
| `${latestVersion}` | `RELEASE_LATEST_VERSION` | Current/previous version | `1.1.0` |
| `${tagName}` | `RELEASE_TAG_NAME` | Full tag name | `v1.2.0` |
| `${branchName}` | `RELEASE_BRANCH_NAME` | Current branch | `main` |
| `${changelog}` | `RELEASE_CHANGELOG` | Generated changelog text (available from the changelog step on) | |
| `${releaseUrl}` | `RELEASE_RELEASE_URL` | URL of the created release (after the platform step) | |
| `${repo.remote}` | `RELEASE_REPO_REMOTE` | Remote URL | `git@github.com:octocat/my-project.git` |
| `${repo.protocol}` | `RELEASE_REPO_PROTOCOL` | `https` or `ssh` | `ssh` |
| `${repo.host}` | `RELEASE_REPO_HOST` | Remote host | `github.com` |
| `${repo.owner}` | `RELEASE_REPO_OWNER` | Repository owner | `octocat` |
| `${repo.repository}` | `RELEASE_REPO_REPOSITORY` | Repository name | `my-project` |

Variables are filled in as the pipeline progresses: `${version}` and `${tagName}` are known from the version step on, `${changelog}` after the changelog step, `${releaseUrl}` after the GitHub/GitLab step.

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
| `HTTPS_PROXY`, `NO_PROXY` | Proxy for GitHub/GitLab/webhook requests |
| `RELEASE_*` | Set *by* release-it-go for hook commands (see [Template Variables](#template-variables)) |

## License

MIT
