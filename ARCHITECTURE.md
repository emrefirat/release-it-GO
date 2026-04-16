# ARCHITECTURE — release-it-go

> Explains how the system works. For a new developer: "which package do I start with, how does the flow work?"

Last updated: 2026-04-16 | See also: [DECISIONS.md](DECISIONS.md) (why), [CLAUDE.md](CLAUDE.md) (rules)

---

## 1. System Overview

**release-it-go** is a CLI tool delivered as a single static Go binary. It uses a pipeline orchestrator pattern to coordinate a **10-step release pipeline**. Each step is an independent function, runs sequentially, and a failure stops the pipeline (except `notification`).

```
┌────────────────────────────────────────────────────────────┐
│                      cmd/release-it-go                     │
│                         (main.go)                          │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│                       internal/cli                         │
│   root.go (release pipeline) | init.go | install.go        │
│              (Cobra command tree)                          │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│                     internal/runner                        │
│   Runner.Run() → 10-step pipeline + lifecycle hooks        │
│   ReleaseContext (shared state across steps)               │
└─┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬───────┘
  │      │      │      │      │      │      │      │
  ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼
config  git  version changelog bumper hook  release notification
                                            └──┬──┘
                                  ┌────────┬───┴───┐
                                  ▼        ▼       ▼
                                GitHub  GitLab  Slack/Teams
                                  REST    REST   webhook
```

---

## 2. Package Responsibilities

| Package | Responsibility | Public Types |
|---------|----------------|--------------|
| `cli` | Cobra command tree, flag parsing, runner invocation | `Execute()`, `NewRootCommand()` |
| `config` | Config loading (JSON/YAML/TOML), defaults, merge, write, npm compat, migration | `Config`, `LoadConfig`, `DefaultConfig`, `WriteConfigYAML` |
| `runner` | Pipeline orchestration, ReleaseContext (shared state), hook integration | `Runner`, `ReleaseContext`, `NewRunner` |
| `git` | Git CLI wrapper (commit, tag, push, log, repo info), prerequisite checks | `Git`, `RepoInfo`, `GetRepoInfo`, `commandExecutor` (var) |
| `githook` | Git hook installer (`.hooks/` + core.hooksPath, managed header) | `Installer`, `HooksFromConfig` |
| `version` | Semver parse/increment, CalVer, branch-aware pre-release | `ParseVersion`, `IncrementVersion`, `CalVer` |
| `changelog` | Conventional commit parser, lint, bump analyzer, renderer (conventional + keep-a-changelog) | `Commit`, `ParseCommits`, `LintCommits`, `AnalyzeBump`, `GenerateChangelog` |
| `release` | GitHub + GitLab REST client, asset upload, PR/MR/issue comment | `ReleaseProvider`, `GitHubClient`, `GitLabClient`, `ReleaseOptions` |
| `bumper` | Multi-file version writing (JSON/YAML/TOML/INI/text), reading from input file | `Bumper`, `BumperFile`, `ReadVersionFromFile`, `WriteVersionToFile` |
| `hook` | Lifecycle hook runner (before:/after: shell commands, template variable substitution) | `HookRunner`, `RunHooks`, `SetVars` |
| `notification` | Webhook notifications (Slack JSON, Teams MessageCard) | `Client`, `RichNotificationContext`, `SendAll` |
| `ui` | Prompter (interactive Bubbletea / non-interactive CI), spinner, colors, CI detect | `Prompter`, `InteractivePrompter`, `NonInteractivePrompter`, `Spinner`, `IsCI` |
| `log` | slog wrapper (4 verbose levels + dry-run formatting) | `Logger`, `NewLogger` |

### Package Dependency Graph (directional)

```
cli ─→ config, runner, ui, changelog, log, githook
runner ─→ config, git, version, changelog, bumper, hook, notification, release, ui, log
release ─→ config, git, log
notification ─→ config, log
git ─→ config, log
hook ─→ config, log
bumper ─→ config, log
config (leaf)
ui (leaf)
log (leaf)
version (leaf — only Masterminds/semver)
changelog ─→ git (only for the RepoInfo type)
githook ─→ config
```

**Critical rule**: `config`, `log`, `ui`, `version` are leaf packages — they don't import `runner`/`cli`. `runner` orchestrates everything. No cycles.

---

## 3. Pipeline Flow

### 3.1 Full Pipeline (`r.Run()`)

```
1. init                  → fetch repoInfo + branchName
2. prerequisites         → git installed?, in repo?, clean working dir?, upstream?, commits?, token validation
3. commitlint            → conventional commit format validation (if RequireConventionalCommits)
4. version               → latestTag → IncrementVersion (auto-detected bump type or CLI flag)
5. bump                  → write the new version into BumperConfig.Out files
6. changelog             → update CHANGELOG.md, store content in ReleaseContext.Changelog
7. git:release           → stage → commit → tag → push (each prompts separately, can be skipped with --no-X)
8. github:release        → GitHub REST: create release, upload assets
9. gitlab:release        → GitLab REST: create release, upload assets
10. notification         → Slack + Teams webhook (non-fatal: error → log + continue)
```

Before each step, a `before:STEP` hook runs; after each step, an `after:STEP` hook runs.

### 3.2 Alternate Modes

`internal/runner/runner.go` provides additional pipeline variants:
- `RunChangelogOnly()` — only compute the changelog, print to stdout
- `RunReleaseVersionOnly()` — only compute the next version, print to stdout
- `RunOnlyVersion()` — interactive version prompt, then run rest in CI mode
- `RunNoIncrement()` — release without incrementing the version (on existing tag)
- `RunCheckCommits()` — only commit lint, no release

CLI flags (`--changelog`, `--release-version`, `--only-version`, `--no-increment`, `--check-commits`, `--check-msg`) select these modes.

### 3.3 ReleaseContext (Shared State)

State shared across the pipeline. Steps communicate through this struct:

```go
type ReleaseContext struct {
    Config     *config.Config
    Logger     *applog.Logger
    Git        *git.Git
    Prompter   ui.Prompter
    HookRunner *hook.HookRunner
    Spinner    *ui.Spinner

    // Populated by steps
    LatestVersion string  // e.g., "1.2.3"
    Version       string  // e.g., "1.3.0" (new)
    TagName       string  // e.g., "v1.3.0" (template render)
    Changelog     string  // markdown
    ReleaseURL    string  // GitHub/GitLab release URL
    RepoInfo      *git.RepoInfo
    BranchName    string
    IsDryRun      bool
    IsCI          bool
    noCommits     bool

    // For hook and template substitution
    Vars map[string]string  // ${version}, ${tagName}, ${repo.host}, ...
}
```

`UpdateVars()` is called after every step so hooks see the current template values.

---

## 4. Critical Interfaces

### `ui.Prompter`
```go
type Prompter interface {
    SelectVersion(current, recommended string, options []VersionOption) (string, error)
    Confirm(message string, defaultYes bool) (bool, error)
    Input(message string, defaultValue string) (string, error)
    Select(question string, options []string, defaultIndex int) (int, error)
}
```
- `InteractivePrompter` (Bubbletea TUI) — when TTY is available
- `NonInteractivePrompter` (auto-default answers) — CI mode or `--ci` flag

CI detection: `IsCI()` checks env vars (CI, GITHUB_ACTIONS, GITLAB_CI, ...) or whether stdin is a TTY → CI mode if not.

### `release.ReleaseProvider`
```go
type ReleaseProvider interface {
    CreateRelease(opts ReleaseOptions) (*ReleaseResult, error)
    UploadAssets(releaseID string, assets []string) error
    PostComment(target CommentTarget, message string) error
    ValidateToken() error
}
```
- `GitHubClient` — `https://api.github.com` or GHE custom host
- `GitLabClient` — `https://gitlab.com/api/v4` or self-hosted

To add a new platform, implement this interface + add a new step in `runner.go`.

### `commandExecutor` (Function Variable Pattern)
For test mocking. Defined separately in `internal/git/git.go` and `internal/githook/githook.go` (because they are different packages):

```go
var commandExecutor = func(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
    out, err := cmd.CombinedOutput()
    return strings.TrimSpace(string(out)), err
}
```

---

## 5. External System Integration

| System | Protocol | Auth | Timeout | Failure Behavior |
|--------|----------|------|---------|------------------|
| Git CLI | exec.Command | Local git config | OS default | Pipeline stops |
| GitHub API | HTTPS REST | `GITHUB_TOKEN` env | 30s (default) | Pipeline stops |
| GitLab API | HTTPS REST | `GITLAB_TOKEN` env (`Private-Token` header) | 30s (default) | Pipeline stops |
| Slack Webhook | HTTPS POST | Webhook URL (from env, `urlRef`) | 30s | Non-fatal: log + continue |
| Teams Webhook | HTTPS POST (MessageCard JSON) | Webhook URL (from env, `urlRef`) | 30s | Non-fatal: log + continue |

### Token / Secret Management
**No token is stored in the config file.** Pattern:

```yaml
github:
  tokenRef: "GITHUB_TOKEN"  # env var name

notification:
  webhooks:
    - type: slack
      urlRef: "SLACK_WEBHOOK_URL"  # env var name
```

Read at runtime via `os.Getenv(tokenRef)`. `getToken()` (`internal/release/release.go`) is the helper.

### URL Credential Stripping
HTTPS git remotes can include credentials: `https://user:token@github.com/owner/repo.git`. `ParseRepoURL()` (`internal/git/repo.go`) strips them automatically — they don't leak into logs.

---

## 6. Config Loading Strategy

`internal/config/loader.go` — `LoadConfig(path string)`:

1. If path is given → load that file
2. If no path → search in this order:
   ```
   .release-it-go.json → .yaml → .yml → .toml
   .release-it.json    → .yaml → .yml → .toml  (legacy npm)
   ```
3. If nothing found → return `DefaultConfig()`

**Native (.release-it-go.\*) > Legacy (.release-it.\*)** priority.

### npm Compat (`internal/config/compat.go`)
The npm `release-it` config uses different types in some fields (string vs bool vs object). `normalizeJSON()` and `applyPluginCompat()` convert these. Goal: a user migrating from npm shouldn't have to change the config.

### Migration (`internal/config/migrate.go`)
The `init` command, on detecting a legacy `.release-it.json`, offers migration. `MigrateLegacyConfigTo(format)` produces JSON or YAML output.

---

## 7. Versioning Algorithm

```
1. Find latest tag (from Git):
   - Default: GetLatestTag() (most recent tag)
   - If tagName template changed: filter via matchesTagNameFormat()
   - If PreReleaseID is set: GetLatestPreReleaseTagMerged() + GetLatestStableTagMerged() (branch-aware)

2. Determine bump type:
   - If --increment flag is set, use it
   - Otherwise autoDetectIncrement(): compute from conventional commits
     - feat → minor
     - fix → patch
     - feat! or BREAKING CHANGE → major

3. Increment the version:
   - SemVer: IncrementVersion(current, type, preReleaseID)
   - CalVer: cv.NextVersion(latest) — yy.mm.minor format
   - Pre-release: prerelease (continue same ID: .0 → .1) or pre+major/minor/patch (new series)

4. If interactive: Prompter.SelectVersion(current, recommended, options)
```

### Branch-Aware Pre-Release (Phase 15)
Problem: When working on `feature-x` branch with `--preRelease=deneme`, tags from `feature-y` branch (e.g., `v2.0.0-deneme.5`) were being picked up and breaking series.

Solution: Use `git tag -l --merged HEAD --sort=-v:refname` — only look at tags merged into HEAD. The `resolvePreReleaseBaseTag()` algorithm:
1. Find the latest pre-release tag (matching ID) merged into HEAD
2. Find the latest stable tag merged into HEAD
3. Pre-release base version >= stable → continue series
4. Otherwise → start a new series from the stable tag

---

## 8. Concurrency Model

**The pipeline runs on a single goroutine.** No concurrency, because:
- Steps are dependent (version → bump → changelog → release)
- Git commands must serialize (to prevent races)
- HTTP requests are singular (except asset upload — also sequential)

**Spinner exception**: The TUI spinner runs on its own goroutine (Bubbletea). Closes via the `Stop()` signal.

**No goroutine leaks** because `context.Context` is not used — the pipeline is short-lived (seconds). If long-lived/cancellable operations are added later (e.g., parallel multi-platform release), context propagation must be added.

---

## 9. Error Recovery Strategies

| Scenario | Behavior |
|----------|----------|
| Git command fails | Pipeline stops, error wrap with detailed message |
| GitHub/GitLab token missing | Pipeline stops, "set X env var" message |
| GitHub/GitLab API 4xx/5xx | Pipeline stops, parse error from response body |
| Webhook send fails | **Non-fatal**: warn log + pipeline continues |
| "no commits since latest tag" | Graceful exit (info message, not error) |
| Conventional commit lint fails | Pipeline stops; bypass with `--ignore-commit-lint` |
| Dry-run | No side effects, only logs ("[dry-run] would do X") |
| No TTY + interactive prompt | Auto CI mode (default answer) — detected via `go-isatty` |
| `requireCleanWorkingDir` violation | Pipeline stops, "uncommitted changes" |

### Idempotency
- If tag already exists → "tag already exists" error (manual intervention required)
- If release already exists → GitHub/GitLab API error (manual intervention)
- If CHANGELOG.md already has content → preserves header and prepends new section (`UpdateChangelogFile`)

---

## 10. Deployment Topology

```
Developer Machine                 CI/CD Server                     GitHub/GitLab
─────────────────                 ────────────                     ─────────────
go install                        Docker (alpine:3.21)              REST API
  or                              Non-root user (releaser:1000)
GitHub Releases (tar.gz)          .release-it-go.yaml mount
  or
Docker (release-it-go:latest)     git identity check
                                  GITHUB_TOKEN env
                                  release-it-go --ci

                                  ↓
                         Pipeline runs →  Tag/release/changelog/notification
```

### Self-Hosting (`.github/workflows/release.yml`)
release-it-go releases its own binary using release-it-go:
1. CI workflow: release-it-go creates the tag
2. GoReleaser: triggered by the tag, builds multi-platform binaries + GitHub release

### Scaling
The pipeline runs as a **single instance** — parallel runs are not supported (race risk). In CI workflows, two concurrent release-it-go runs on the same branch will cause tag conflicts.

---

## 11. Test Architecture

| Type | Location | Approach |
|------|----------|----------|
| Unit | `internal/*/[name]_test.go` | Table-driven, `commandExecutor` mock |
| Integration | `test/integration/*_test.go` | `t.TempDir()` + `git init`, real git commands |
| Coverage | `make coverage` | HTML report, `coverage.html` |
| Race | `go test -race ./internal/...` | Required in CI |
| Vulnerability | `govulncheck ./...` | `make vuln`, before commit |

### Mock Pattern
```go
// internal/git/git_test.go
original := commandExecutor
defer func() { commandExecutor = original }()
commandExecutor = func(name string, args ...string) (string, error) {
    if args[0] == "tag" { return "v1.2.3", nil }
    return "", nil
}
```

### Integration Test Pattern (`test/integration/helpers_test.go`)
```go
dir := t.TempDir()
initGitRepo(t, dir)
createTag(t, dir, "v1.0.0")
createCommits(t, dir, []string{"feat: ...", "fix: ..."})
cfg := newTestConfig(dir)  // CI mode, push false, etc.
r := runner.NewRunner(cfg)
err := r.Run()
```

---

## 12. Notable Design Patterns

### Spinner Pattern
```go
r.ctx.Spinner.Start("Doing X")
if err := doX(); err != nil {
    r.ctx.Spinner.Stop(false)  // false = ✗
    return err
}
r.ctx.Spinner.Stop(true)  // true = ✓
```
In CI mode, `Start()` is a no-op; `Stop()` writes the result line (to prevent the duplicate-line bug — Phase 10).

### Template Variable Substitution
Placeholders like `${version}`, `${tagName}`, `${repo.host}`:
- Config fields: `tagName: "v${version}"`, `commitMessage: "release v${version}"`
- Hook commands: `"echo bumped to v${version}"`
- `renderTagName()` and `hook.renderTemplate()` perform the substitution.

### Error Wrap Convention
```go
return fmt.Errorf("creating GitHub release: %w", err)
//                  ^^ action            ^^ wrap
```
The root cause is reachable via `errors.Is`/`errors.As`; the user sees a stack-like context.

### Dry-Run Discipline
Every side-effecting step checks `r.ctx.IsDryRun`:
```go
if r.ctx.IsDryRun {
    r.ctx.Logger.DryRun("Would update %s", file)
    return nil
}
```
On the git side, the `isWriteOperation()` map knows which git commands are writes.

---

## 13. Known Limitations

- **Single-goroutine pipeline** — no parallel platform release (GitHub + GitLab are sequential)
- **No `context.Context`** — relies on OS Ctrl+C; no programmatic cancellation
- **No plugin system** (intentional, see DECISIONS.md ADR-006) — extensions are done via config + hooks
- **No rate limiting** — GitHub/GitLab API rate limits are the user's responsibility
- **Single-threaded asset upload** — large binaries can be slow to upload
- **Single-threaded notifications** — multiple webhooks are sent sequentially
- **Git CLI dependency** — `go-git` is not used (intentional, see DECISIONS.md ADR-002)

---

## 14. Important File Paths (Quick Reference)

| File | Description |
|------|-------------|
| `cmd/release-it-go/main.go` | Entry point; ldflags inject version |
| `internal/cli/root.go` | Cobra root, flags, runRelease() |
| `internal/runner/runner.go` | Pipeline definition (`pipelineStep` slice) |
| `internal/runner/context.go` | ReleaseContext + UpdateVars |
| `internal/config/config.go` | All config structs |
| `internal/config/defaults.go` | Default values — touch when adding a new field |
| `internal/git/git.go` | `commandExecutor` mock variable, isWriteOperation |
| `internal/changelog/lint.go` | Allowed conventional commit types |
| `internal/release/release.go` | ReleaseProvider interface, getToken |
| `internal/githook/githook.go` | `.hooks/` installer (Phase 20) |
| `internal/notification/notification.go` | Webhook client, RichNotificationContext |
| `internal/version/semver.go` | IncrementVersion, ParseVersion |
| `internal/ui/ci.go` | CI detection (env var + isatty) |
