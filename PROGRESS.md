# PROGRESS.md — release-it-go Project Progress Tracker

> Tracks the overall project progress and the status of each phase.
> Must be updated at the end of every development session.

---

## Overall Status

| Phase | Title | Status | Progress |
|-------|-------|--------|----------|
| 1 | Core Foundation | Complete | 100% |
| 2 | Git Operations | Complete | 100% |
| 3 | Conventional Commits + Changelog | Complete | 100% |
| 4 | GitHub + GitLab Releases | Complete | 100% |
| 5 | Interactive UI + Hooks + Pipeline | Complete | 100% |
| 6 | Advanced Features | Complete | 100% |
| 7 | Testing, CI/CD, Documentation | Complete | 100% |
| 8 | Init Command & Dual Config | Complete | 100% |
| 9 | Conventional Commit Linting | Complete | 100% |
| 10 | UI/Output Improvements | Complete | 100% |
| 11 | Docker Container Support | Complete | 100% |
| 12 | Docker Pre-flight Checks | Complete | 100% |
| 13 | Webhook Notification (Slack + Teams) | Complete | 100% |
| 14 | YAML Config Writing + Init Format Selection | Complete | 100% |
| 15 | Branch-Aware Pre-Release Version Detection | Complete | 100% |
| 16 | Critical Bug Fixes — URL Parsing + CalVer | Complete | 100% |
| 17 | Pipeline Robustness Improvements | Complete | 100% |
| 18 | Config Compatibility & Edge Case Fixes | Complete | 100% |
| 19 | Test Coverage Strengthening | Complete | 100% |
| 20 | Git Hook Management (install / remove / check-msg) | Complete | 100% |
| 21 | P0 Test Coverage Completion (QA audit) | Complete | 100% |
| 22 | Atomic Git Push Default | Complete | 100% |

**Last Updated:** 2026-08-25
**Active Developer:** Claude
**Current Version:** v0.1.3 (Phase 22 complete — production-ready)

---

## Phase 1: Core Foundation

**Status:** Complete
**PRD:** `docs/phase_1.md`

### To Do

- [x] Go module init (`go mod init`)
- [x] Cobra CLI skeleton
- [x] Config struct definitions
- [x] Config loader (JSON/YAML/TOML)
- [x] Default values
- [x] CLI flags → config merge
- [x] Read version from git tag
- [x] Read version from VERSION file
- [x] Semver parse/increment/compare
- [x] Template variable rendering
- [x] Logger (verbose levels)
- [x] Makefile
- [x] Unit tests
- [x] CalVer struct and basic implementation

### Notes

- Test coverage: cli=82.9%, config=87.9%, log=100%, version=86.5%
- semver.IncPatch() strips pre-release on pre-release versions (1.2.3-beta.0 → 1.2.3); this is correct semver behavior
- Mapstructure tags added for Viper config unmarshaling
- `runGit` is a function variable for test mocking

---

## Phase 2: Git Operations

**Status:** Complete
**PRD:** `docs/phase_2.md`

### To Do

- [x] Git runner (command execution)
- [x] Prerequisite checks (branch, clean, upstream, commits)
- [x] Stage + Commit
- [x] Tag creation
- [x] Push
- [x] Repo info parse (HTTPS + SSH)
- [x] Simple git log changelog
- [x] Dry-run support
- [x] Unit tests

### Notes

- Test coverage: git=88.7%
- The `commandExecutor` function variable allows mocking git commands in tests
- `isWriteOperation` keeps read operations working in dry-run mode
- `TagExists` always runs the real git command (including in dry-run)
- HTTPS and SSH remote URL formats are parsed via regex

---

## Phase 3: Conventional Commits + Changelog

**Status:** Complete
**PRD:** `docs/phase_3.md`

### To Do

- [x] Conventional commit parser
- [x] Bump analyzer (major/minor/patch)
- [x] Conventional-changelog format
- [x] Keep-a-changelog format
- [x] CHANGELOG.md file update
- [x] Unit tests

### Notes

- Test coverage: changelog=93.3%
- Conventional commit parsed via regex (type, scope, !, description, body, footers)
- Breaking change detection: footer (BREAKING CHANGE:) and bang (feat!) supported
- Conventional-changelog: Features, Bug Fixes, Performance Improvements, Reverts, BREAKING CHANGES sections
- Keep-a-changelog: Added, Changed, Fixed, Removed sections
- `insertAfterHeader` preserves existing CHANGELOG.md content while prepending

---

## Phase 4: GitHub + GitLab Releases

**Status:** Complete
**PRD:** `docs/phase_4.md`

### To Do

- [x] Release provider interface
- [x] GitHub client (create, upload, comment)
- [x] GitLab client (create, upload, comment)
- [x] Token management
- [x] Asset upload (glob)
- [x] GitHub Enterprise support
- [x] GitLab CA certificate support
- [x] Dry-run support
- [x] API mock tests

### Notes

- Test coverage: release=73.7%
- No external SDK used; direct REST API calls via net/http
- GitHub: CreateRelease, UploadAssets, PostComment, ValidateToken, GHE URL, proxy, makeLatest, autoGenerate, discussionCategory
- GitLab: CreateRelease, UploadAssets (Generic Package + Release Link), PostComment (MR/issue), ValidateToken, CA cert, custom token header
- Mock API tests via httptest.NewServer
- Asset content type detection: 12+ formats (zip, tar.gz, dmg, deb, rpm, exe, sig, etc.)

---

## Phase 5: Interactive UI + Hooks + Pipeline

**Status:** Complete
**PRD:** `docs/phase_5.md`

### To Do

- [x] Version selection prompt
- [x] Confirm prompts
- [x] Spinner animation
- [x] Colored output
- [x] CI environment detection
- [x] Hook runner (before/after lifecycle)
- [x] Main pipeline orchestrator
- [x] Special modes (--changelog, --release-version, --only-version)
- [x] Summary output
- [x] Unit tests

### Notes

- Test coverage: ui=42.9% (bubbletea interactive models require terminal), hook=100%, runner=25.3% (pipeline steps need git mocks)
- Bubbletea v1.3.10 for interactive terminal UI (selectModel, confirmModel, inputModel)
- Lipgloss v1.1.0 for coloring, NO_COLOR environment variable supported
- CI detection: GITHUB_ACTIONS, GITLAB_CI, CIRCLECI, TRAVIS, JENKINS_URL, BITBUCKET_PIPELINE, CODEBUILD_BUILD_ID, TF_BUILD
- NonInteractivePrompter: auto-answers all prompts in CI mode
- HookRunner: 12 lifecycle events (before/after: init, bump, release, git:release, github:release, gitlab:release)
- Template variable rendering: ${version}, ${tagName}, ${changelog}, ${releaseUrl}, ${branchName}, ${repo.*}
- Pipeline: init → prerequisites → version → changelog → git:release → github:release → gitlab:release
- Before/after hook execution and UpdateVars on every step
- Dry-run supported in all steps

---

## Phase 6: Advanced Features

**Status:** Complete
**PRD:** `docs/phase_6.md`

### To Do

- [x] Bumper: read version from file (JSON/YAML/TOML/INI/text)
- [x] Bumper: write version to file
- [x] Bumper: glob pattern support
- [x] CalVer runner integration
- [x] Pre-release flows (already implemented in Phase 1's semver.go)
- [x] --no-increment mode
- [x] --only-version mode
- [x] --changelog and --release-version CLI modes
- [x] CalVer + SemVer conflict detection
- [x] Bumper pipeline step (bump step)
- [x] Unit tests

### Notes

- Test coverage: bumper=87.8%
- Bumper: JSON (nested path), YAML, TOML, INI ([section].key), text (consumeWholeFile) support
- Bumper: glob pattern (*.json, charts/*/Chart.yaml), prefix (^, ~), dry-run support
- CalVer: integrated into the pipeline via `runner.determineCalVer()`
- CalVer + pre-release cannot be used together (validation in CLI)
- "bump" step added to the pipeline: version → bump → changelog
- CLI modes: RunChangelogOnly, RunReleaseVersionOnly, RunOnlyVersion, RunNoIncrement
- RunOnlyVersion: switches to CI mode automatically after version selection
- RunNoIncrement: updates changelog and release without incrementing version
- Existing YAML/TOML deps used (go-yaml, go-toml via Viper)
- Simple INI parser written using stdlib (no external dep)

---

## Phase 7: Testing, CI/CD, Documentation

**Status:** Complete
**PRD:** `docs/phase_7.md`

### To Do

- [x] Integration tests
- [x] API mock tests
- [x] Coverage 80%+ target
- [x] GitHub Actions CI workflow
- [x] GitHub Actions Release workflow
- [x] GoReleaser config
- [x] Shell completions (bash/zsh/fish)
- [x] Build info (ldflags)

### Notes

- Test coverage: bumper=87.8%, changelog=93.3%, cli=83.0%, config=87.9%, git=86.8%, hook=100%, log=100%, release=86.7%, runner=80.6%, ui=78.6%, version=86.5%
- All packages reached 78%+ coverage; the overall 80%+ target was met
- 17 integration tests: full pipeline, patch/minor/major bump, dry-run, no-tags, changelog-only, release-version-only, disable commit/tag, conventional commit auto-detect, breaking change auto-major, bumper file update, keep-a-changelog, hook execution/failure, config JSON/YAML, no-increment, sequential releases
- Bubbletea model tests are run directly via Init/Update/View (no terminal required)
- GitLab upload assets, error handling, missing token tests added
- GitHub Actions CI: Go 1.22/1.23 matrix, test+lint+build, coverage check
- GitHub Actions Release: triggered by v* tag, automatic release via GoReleaser v2
- GoReleaser: linux/darwin/windows × amd64/arm64, ldflags (cli.Version/Commit/Date), nfpms (deb/rpm/apk)
- Shell completions: bash/zsh/fish/powershell via cobra, testable via cmd.OutOrStdout()
- Race condition tests pass on all packages

---

## Post-Release: Real Environment Tests and Improvements

**Status:** Complete

### To Do

- [x] Real GitLab environment release test (private test repo)
- [x] Security fix: HTTPS URL credential stripping (prevents token leakage in CHANGELOG)
- [x] GoReleaser ldflags fix (main.version/commit/date)
- [x] Old npm release-it config compatibility (normalizeJSON + applyPluginCompat)
  - [x] requireBranch: [] → string conversion
  - [x] gitlab.assets: {links:[]} → []string conversion
  - [x] Mapping changelog settings from the plugins section
  - [x] Cleanup of unknown fields like npm, versionFile
- [x] --preRelease shorthand flag added (sub-branch prerelease support)
- [x] PreRelease field added to GitLabConfig
- [x] Real GitLab CI/CD pipeline test (private test repo)
  - [x] Main branch: automatic release (v1.4.1)
  - [x] Sub-branch: prerelease (v1.5.0-beta.0)
  - [x] git push via SSH, GitLab release API via personal token
- [x] Compat tests (6 tests: conventional-changelog, keep-a-changelog, no-plugins, old npm format, requireBranch array, YAML ignored)

### Notes

- CRITICAL SECURITY FIX: HTTPS credential stripping added to ParseRepoURL. Previously, URLs in oauth2:token@host format were leaking into CHANGELOG compare links.
- npm release-it config compatibility: legacy .release-it.json files (npm, plugins, requireBranch:[], assets:{links:[]}) load without issues.
- --preRelease="identifier" shorthand: sets preReleaseId + automatically marks GitHub/GitLab release as pre-release.
- GitLab CI pipeline pushes via SSH (HTTPS token git push is not reliable), creates release via API token.
- Real environment tests: two private test repos (v0.2.0, v0.3.0, v1.0.0 on one; v1.4.1, v1.5.0-beta.0 on the other) successful.

---

## Phase 8: Init Command & Dual Config File Support

**Status:** Complete
**PRD:** `docs/phase_8.md`

### To Do

- [x] Add `.release-it-go.*` files to configSearchFiles with priority
- [x] Add generic `Select` method to the Prompter interface (bubbletea + non-interactive)
- [x] `config/writer.go` — WriteConfigJSON (smart defaults omission)
- [x] `config/migrate.go` — DetectLegacyConfig + MigrateLegacyConfig
- [x] `cli/init.go` — Init command and wizard flow
- [x] Register the init command in root.go
- [x] Unit tests (writer, migrate, init, prompt select, config priority)
- [x] docs/phase_8.md PRD document

### Notes

- Native config (.release-it-go.*) is searched before legacy config (.release-it.*)
- WriteConfigJSON only writes fields different from defaults (minimal JSON output)
- Migration flow: read legacy → backup → normalizeJSON + applyPluginCompat → write native
- Init wizard: platform, changelog, git ops, commit msg, tag format, branch selection
- Existing runner_test.go mock prompters had Select method added (interface compatibility)
- In the init wizard, when git push is disabled, `requireUpstream` is set to false automatically (upstream check is meaningless without push)
- `requireCleanWorkingDir` is ALWAYS active regardless of push (a dirty working dir during commit/tag is dangerous)
- The init command has no special flag; the root `--ci` flag passes through NonInteractivePrompter, which auto-answers all prompts with defaults
- In `--ci` mode, if `.release-it-go.json` already exists, Confirm("Overwrite?", default=false) → aborts (safe default)

---

## Phase 9: Conventional Commit Linting

**Status:** Complete
**PRD:** `docs/phase_9.md`

### To Do

- [x] `git/changelog.go` — CommitInfo struct + GetCommitsWithHashSinceTag()
- [x] `changelog/lint.go` — LintInput, LintResult, LintCommits() function
- [x] `config/config.go` — RequireConventionalCommits field
- [x] `runner/runner.go` — checkCommitLint() pipeline step + RunCheckCommits() mode
- [x] `cli/root.go` — --check-commits + --ignore-commit-lint flags
- [x] `cli/init.go` — "Require conventional commits?" question in the wizard
- [x] `changelog/lint_test.go` — LintCommits tests (8 tests)
- [x] `git/changelog_test.go` — GetCommitsWithHashSinceTag tests (3 tests)

### Notes

- Circular dependency avoided: the lint function lives in the `changelog` package and uses its own `LintInput` struct rather than `git.CommitInfo`
- Since runner imports both `git` and `changelog`, the conversion happens in runner
- Merge commits (`Merge `) and revert commits (`Revert `) auto-pass
- The `commitPattern` regex (in parser.go) is reused directly
- Pipeline order: init → prerequisites → commitlint → version → ...
- `--check-commits`: standalone lint mode, exits with code 1 on failure
- `--ignore-commit-lint`: overrides RequireConventionalCommits
- All tests pass with race detection

---

## Phase 10: UI/Output Improvements

**Status:** Complete

### To Do

- [x] `ui/colors.go` — 14 Unicode icon constants (IconSuccess, IconFail, IconVersion, IconTag, IconPush, IconRelease, IconChangelog, IconCommit, IconLint, IconSkip, IconLink, IconRocket, IconDryRun, IconWarning)
- [x] `ui/colors.go` — FormatBold() function
- [x] `log/logger.go` — Print() method (slog-format-free direct output)
- [x] `log/logger.go` — Verbose() format change (slog → `↳` indented dim format)
- [x] `ui/spinner.go` — CI Start() output (`-` → `⠋` spinner frame)
- [x] `ui/spinner.go` — CI Stop() output (`OK`/`FAIL` → colored `✓`/`✗` icons)
- [x] `runner/runner.go` — printBanner() (🚀 release-it-go / 🧪 dry-run banner)
- [x] `runner/runner.go` — Version message (Info → Print + 📦 icon)
- [x] `runner/runner.go` — Skip messages (Info → Print + ⏭️ icon)
- [x] `runner/runner.go` — printSummary() redesigned with lipgloss border box
- [x] `log/logger_test.go` — Print, Verbose format tests (4 new tests)
- [x] All tests pass (`go test ./... -race`)
- [x] `go vet` and `go fmt` clean

### Notes

- Icons are Unicode characters, not ANSI codes. `NO_COLOR` only disables lipgloss colors; icons always show.
- Logger.Print(): writes directly to stderr without slog formatting; for user-friendly messages.
- Logger.Verbose(): `    ↳ message` format, indented and dim color (verbose >= 1)
- Logger.Debug(): existing slog format preserved (no change)
- CI spinner: starts with `⠋` frame, ends with colored `✓`/`✗`
- printSummary: lipgloss RoundedBorder box, iconified lines, duration info
- printBanner: added at the start of Run(), RunOnlyVersion(), RunNoIncrement()
- 5 files affected: ui/colors.go, log/logger.go, ui/spinner.go, runner/runner.go, log/logger_test.go

---

## Phase 11: Docker Container Support

**Status:** Complete
**PRD:** `docs/phase_11.md`

### To Do

- [x] `docs/phase_11.md` PRD document
- [x] `.dockerignore` build context filter
- [x] `Dockerfile` multi-stage build (golang:1.24.3-alpine → alpine:3.21)
- [x] `Makefile` docker-build and docker-run targets
- [x] `PROGRESS.md` Phase 11 update

### Notes

- Multi-stage build: builder (golang:1.24.3-alpine) + runtime (alpine:3.21)
- Static binary: CGO_ENABLED=0 GOOS=linux, -trimpath -ldflags="-s -w"
- Runtime packages: git, openssh-client, ca-certificates
- Non-root user: releaser (UID/GID 1000, can be changed via build arg)
- `git safe.directory '*'` for safe access to mounted repos
- Build ARGs: VERSION, COMMIT, BUILD_DATE, USER_UID, USER_GID
- Estimated image size: ~30MB
- OCI metadata labels added

---

## Phase 12: Docker Pre-flight Checks

**Status:** Complete

### To Do

- [x] `git/prerequisites.go` — checkGitIdentity() function (user.name/user.email check)
- [x] `git/prerequisites.go` — identity check added to CheckPrerequisites()
- [x] `git/prerequisites_test.go` — Identity tests (5 tests: commit disabled, both set, name missing, email missing, both missing)
- [x] `runner/runner.go` — checkTokens() function (GitHub/GitLab token check)
- [x] `runner/runner.go` — token check added to checkPrerequisites()
- [x] `runner/runner_test.go` — Token tests (11 tests: release disabled, token missing/set, custom tokenRef, skipChecks, both platforms)
- [x] All tests pass (`go test ./... -race`)
- [x] `go vet` and `go build` clean

### Notes

- Git identity check is only performed if `git.commit: true` (unnecessary in tag-only or push-only scenarios)
- Token check is at the `runner` level because access to config (GitHub/GitLab settings) is required
- `skipChecks: true` skips the token check (when CI uses a different auth mechanism)
- Custom `tokenRef` support: the user can use a different env variable name
- Errors fail early (in prerequisites), not late in the pipeline

---

## Phase 13: Webhook Notification Support (Slack + Teams)

**Status:** Complete
**PRD:** `docs/phase_13.md`

### To Do

- [x] `internal/config/config.go` — NotificationConfig + WebhookConfig structs
- [x] `internal/config/defaults.go` — Default notification config (disabled, empty webhooks)
- [x] `internal/notification/notification.go` — Client, SendAll, HTTP POST, resolveURL, renderMessage
- [x] `internal/notification/slack.go` — Slack payload builder ({"text": "..."})
- [x] `internal/notification/teams.go` — Teams MessageCard payload builder
- [x] `internal/notification/notification_test.go` — 13 tests, 98%+ coverage (httptest mock server)
- [x] `internal/runner/runner.go` — sendNotification() pipeline step (added to all pipelines)
- [x] `internal/runner/runner_test.go` — 3 tests: disabled, empty webhooks, non-fatal error
- [x] `docs/phase_13.md` — PRD document
- [x] All tests pass (`go test ./... -race`)
- [x] `go vet` and `go build` clean

### Notes

- Notification non-fatal: a failure logs a warning but doesn't stop the pipeline
- Webhook URL is not written to config directly for security; the env variable name is specified via `urlRef`
- Platform-specific default templates exist for Slack and Teams
- The user can define a custom template via `messageTemplate`
- Timeout is configurable (default: 30 seconds)
- No HTTP call is made in dry-run mode

---

## Phase 14: YAML Config Writing + Init Format Selection

**Status:** Complete
**PRD:** `docs/phase_14.md`

### To Do

- [x] `internal/config/writer.go` — WriteConfigYAML + WriteConfigYAMLWith functions
- [x] `internal/config/writer.go` — WriteConfigJSONWith function (ForceFields support)
- [x] `internal/config/writer.go` — ForceFields type and toConfigMap (diffStructForce)
- [x] `internal/config/writer.go` — fullExampleYAML constant (commented YAML reference)
- [x] `internal/config/writer.go` — fullExampleJSON and WriteFullExampleJSON removed
- [x] `internal/config/migrate.go` — NativeConfigFileYAML constant
- [x] `internal/config/migrate.go` — NativeConfigFileForFormat() function
- [x] `internal/config/migrate.go` — DetectNativeConfigAny() function
- [x] `internal/config/migrate.go` — MigrateLegacyConfigTo() function (with format parameter)
- [x] `internal/cli/init.go` — Format selection question (JSON / YAML, first question)
- [x] `internal/cli/init.go` — Explicit writing of wizard-configured fields via ForceFields
- [x] `internal/cli/init.go` — Rename old config to .bak when format changes
- [x] `internal/cli/init.go` — --full-example YAML output (.release-it-go-full.yaml)
- [x] `internal/config/writer_test.go` — YAML writing tests (default, non-default, full example, loadable)
- [x] `internal/config/writer_test.go` — TestToConfigMap_ForceFieldsIncludesDefaults
- [x] `internal/cli/init_test.go` — TestRunInit_YAMLFormat
- [x] `internal/cli/init_test.go` — TestRunInit_FormatSwitch_RenamesOldConfig
- [x] `internal/cli/init_test.go` — TestRunInit_WizardWritesExplicitFields
- [x] `docs/phase_14.md` — PRD document
- [x] All tests pass (`go test ./... -race`)
- [x] `go vet` and `go fmt` clean

### Notes

- Thanks to YAML comment support, `--full-example` now contains a description for every option
- ForceFields mechanism: every field the wizard asks is written to the config file even if it equals the default (e.g., `commit: true`, `infile: "CHANGELOG.md"`)
- When the format changes (JSON→YAML), the old file is backed up as `.bak`; the two configs do not coexist
- Migration also supports format selection (MigrateLegacyConfigTo)
- `go.yaml.in/yaml/v3` (Viper's indirect dependency) is now used directly

---

## Phase 15: Branch-Aware Pre-Release Version Detection

**Status:** Complete

### To Do

- [x] `internal/git/tag.go` — GetLatestPreReleaseTagMerged() method (branch-scoped pre-release tag search via --merged HEAD)
- [x] `internal/git/tag.go` — GetLatestStableTagMerged() method (branch-scoped stable tag search via --merged HEAD)
- [x] `internal/runner/runner.go` — resolvePreReleaseBaseTag() method (continue/new series decision)
- [x] `internal/runner/runner.go` — determineVersion() updated (branch-aware resolution if preReleaseID set)
- [x] `internal/git/tag_test.go` — GetLatestPreReleaseTagMerged unit tests (8 tests)
- [x] `internal/git/tag_test.go` — GetLatestStableTagMerged unit tests (8 tests)
- [x] `test/integration/release_test.go` — PreRelease_BranchAware_ContinueSeries (continue series test)
- [x] `test/integration/release_test.go` — PreRelease_BranchAware_NewSeries (new series test)
- [x] `test/integration/release_test.go` — PreRelease_BranchAware_MasterAdvanced (master advanced, series continues test)
- [x] `test/integration/release_test.go` — PreRelease_NoFlag_BehaviorUnchanged (standard behavior preserved test)

### Notes

- Problem: When working on different branches with `--preRelease="deneme"`, tags on other branches (e.g., v2.0.0-beta.0) were being found as the latest tag and breaking series
- Solution: `git tag -l --merged HEAD --sort=-v:refname` only looks at tags reachable from the current branch
- Algorithm: Pre-release tag found AND base version >= stable → continue series; otherwise start a new series
- 3 scenarios supported: long-living branch (continue series), deleted/recreated branch (new series), master advanced but branch isolated (continue series)
- If PreReleaseID is empty (standard release), existing behavior is preserved
- TagMatch/TagExclude filters are applied in the new methods too
- 16 unit tests + 4 integration tests added; all tests pass with race detection

---

## Phase 20: Git Hook Management (install / remove / check-msg)

**Status:** Complete
**PRD:** `docs/phase_20.md`

### To Do

- [x] `internal/config/config.go` — git hook fields added to HooksConfig (PreCommit, CommitMsg, PrePush, PostCommit, PostMerge, PrepareCommitMsg)
- [x] `internal/githook/githook.go` — Installer (Install, Remove, generateScript, isManagedHook, HooksFromConfig, FindGitDir, FindProjectDir)
- [x] `internal/githook/githook_test.go` — Unit tests
- [x] `internal/cli/install.go` — `hooks install` and `hooks remove` subcommand group (Cobra)
- [x] `internal/cli/install_test.go` — CLI tests
- [x] `internal/cli/root.go` — `--check-msg <file|->|<message>` flag (single message validation for the commit-msg hook)
- [x] `internal/cli/root.go` — runCheckMsg() function (file path / stdin "-" / direct string support)
- [x] commitlint-style compact output (default), `-V` shows valid type list
- [x] `.hooks/` directory + `git config core.hooksPath` (husky-like)
- [x] Managed header (`# Managed by release-it-go — DO NOT EDIT`) for non-managed hook detection
- [x] `--force` flag to overwrite existing non-managed hooks
- [x] All tests pass; `make check` clean

### Notes

- Uses `.hooks/` (project root) instead of `.git/hooks/` → can be committed to the repo, shared across the team (husky pattern)
- `core.hooksPath` is set on install, reset on remove (if no remaining hooks)
- `--check-msg` 3 input modes: file path, `-` (stdin), direct string. The commit-msg hook script calls `release-it-go --check-msg "$1"`
- Conventional commit validation uses `LintCommits()` from `internal/changelog/lint.go` → same allowedTypes as `--check-commits`
- Output format mimics commitlint (compact, colored, action-oriented error message)
- Phase 20 was split across 5 commits: PRD → install command → hooks rename refactor → tests → check-msg flag → check-msg output improvement → string/stdin support

---

## Phase 21: P0 Test Coverage Completion (QA audit)

**Status:** Complete
**PRD:** `docs/phase_21.md`

### To Do

- [x] Bölüm 1 — `cli/root.go` helpers: `fileExists`, `reasonDescription` (table-driven)
- [x] Bölüm 2 — `cli/root.go` `runCheckMsg` with file / stdin / direct string modes plus verbose output
- [x] Bölüm 3 — `cli/install.go` `runHooksInstall` + `runHooksRemove` against real temp git repos
- [x] Bölüm 4 — `cli/init.go` `runInit` dispatcher (`--full-example`, CI mode, existing native config abort)
- [x] Bölüm 5 — `runner/runner.go` `githubRelease` + `gitlabRelease` dry-run, prompt accept/decline, missing token
- [x] Zero production code changes (tests only)
- [x] `make check` clean (fmt, vet, lint, test, build); `make vuln` unchanged (pre-existing stdlib advisories)

### Coverage Outcome

| Paket | Önce | Sonra | Hedef | Durum |
|-------|------|-------|-------|-------|
| `internal/cli` | 58.8% | **84.9%** | 75%+ | ✓ Exceeded |
| `internal/runner` | 78.1% | **83.5%** | 85%+ | ↘ 1.5% short |
| `runCheckMsg` | 0% | covered | 70%+ | ✓ |
| `runHooksInstall` | 0% | covered | 70%+ | ✓ |
| `runHooksRemove` | 0% | covered | 70%+ | ✓ |
| `runInit` | 0% | covered | 70%+ | ✓ |
| `fileExists` | 0% | 100% | 70%+ | ✓ |
| `reasonDescription` | 0% | covered | 70%+ | ✓ |
| `githubRelease` | 28.6% | **71.4%** | 70%+ | ✓ |
| `gitlabRelease` | 28.6% | **71.4%** | 70%+ | ✓ |

### Notes

- Helpers for stdin/stderr capture (`captureStderr`, `withStdin`) swap globals for the duration of a test closure — restored via defer. No production code touched.
- `setupGitRepo` helper spins up a fresh git repo in `t.TempDir()` with user identity configured; used by all hooks tests.
- Runner tests discovered that `NewRunner` sets `ctx.IsCI` via `ui.IsCI()`, which returns true under `go test` (no TTY on stdin). Tests that need the interactive branch must set `runner.ctx.IsCI = false` explicitly after `NewRunner`. Matches the existing `Interactive_Declined` test pattern.
- `runner` package ended at 83.5% vs 85% target. The remaining gap is `client.ValidateToken()` HTTP round-trip inside `checkTokens` and asset-upload paths inside `githubRelease`/`gitlabRelease` — both require either real HTTPS servers or production-side URL injection. Deferred to Phase 22 (P1 integration tests).
- All tests run under `-race`; no data races introduced.

---

## Phase 22: Atomic Git Push Default

**Status:** Complete
**PRD:** `docs/phase_22.md`

### To Do

- [x] `internal/git/push_test.go` — expect `--atomic` in default args + TestPush_CustomArgsOverridesDefault
- [x] `internal/config/config_test.go` — default test asserts `["--follow-tags", "--atomic"]`
- [x] `internal/config/defaults.go` — append `--atomic` to default `PushArgs`
- [x] `internal/config/writer.go` — fullExampleYAML reflects new default with explanation
- [x] `README.md` — config reference updated
- [x] `TROUBLESHOOTING.md` — entries for "atomic not supported" + orphan tag recovery
- [x] `CHANGELOG.md` — Unreleased entry
- [x] Coverage preserved (internal/git 91.8%, internal/config 88.4%)
- [x] All packages green under `-race`; vet + lint clean

### Notes

- Root cause discovered via user-reported CI failure: `git push --follow-tags` is not atomic; in a Jenkins run, the tag (0.10.0) was pushed while the master ref was rejected with `fetch first`, leaving an orphan tag on the remote. The next CI run computed the same version and tripped over the existing tag.
- `--atomic` (git 2.4+, 2015) forces all refs in a single transaction; if one ref is rejected, none land on the remote. Eliminates the orphan-tag scenario at the source.
- Behavior change is opt-out: legacy git servers (pre-2015) without atomic protocol support reject the push with a clear error message; users override `git.pushArgs` to revert.
- Test-first approach (CLAUDE.md Rule #5): tests updated to expect new default, observed red, then default changed → green. Atomic test+code commit (no broken-in-the-middle state).
- Zero new code in production path — only a one-element append in defaults.go.

---

## Bugs

- [x] BUG: First-release changelog "exit status 128" error (2026-02-16) → When `LatestVersion=0.0.0`, the `v0.0.0` tag was searched but no such tag exists. The `latestVersionToTag()` helper was added: returns empty for `0.0.0` or empty string, so `GetCommitsSinceTag("")` returns all commits. 3 sites affected: `RunChangelogOnly`, `generateChangelog`, `autoDetectIncrement`.
- [x] BUG: Init wizard asked commit/tag/push as a single question (2026-02-16) → If a user wanted commit+tag but not push, they were stuck. Questions split: "Enable git commit and tag?" + "Enable git push?" as two prompts. When push is disabled, `requireUpstream` is automatically false.
- [x] BUG: CHANGELOG.md was not included in the commit after creation (2026-02-16) → `Stage()` defaults to `git add . --update`, which only adds tracked files. The newly created CHANGELOG.md was untracked and skipped. Fix: `StageFile()` method added; at the end of `generateChangelog()`, CHANGELOG.md is staged explicitly via `git add`.
- [x] BUG: A release would happen even with no commits, producing empty CHANGELOG entries (2026-02-16) → `git.requireCommits` defaulted to `false`, allowing back-to-back releases without commits. Fix: default changed to `true`. "Require new commits before release?" added to the init wizard. Now if there are no commits since the last tag, an `no commits since latest tag` error is returned.
- [x] BUG: "no commits since latest tag" should be a graceful exit instead of an error (2026-02-16) → If the user has no commits, an info message and clean exit is preferable to an error. Fix: in prerequisites, replaced error with logger.Print + return nil.
- [x] BUG: CI spinner showed two lines (Start + Stop) (2026-02-16) → `⠋ Initializing...` + `✓ Initializing` repeated. Fix: in CI, Start() no longer outputs anything; only Stop() writes the result line.
- [x] BUG: init step shows ✗ (despite succeeding) (2026-02-16) → A non-fatal GetRepoInfo error was prematurely calling `Stop(false)` on the spinner. Fix: do not stop the spinner on optional errors.
- [x] BUG: printSummary lipgloss box was unnecessary and repetitive (2026-02-16) → User feedback: the frame was unnecessary detail. Fix: switched to a flat, minimal output format.
- [x] BUG: --preRelease re-run with the same ID didn't bump version (2026-02-16) → `1.6.0-deneme2.0 → 1.6.0-deneme2.0` produced the same version, "tag already exists" error. Reason: `prepatch` increment was stripping the existing pre-release and starting again at `.0`. Fix: if the current version already has the same pre-release ID, use the `"prerelease"` increment (which bumps the number: `.0 → .1`).
- [x] BUG: --check-commits accepted invalid commit types (2026-02-16) → Invalid types like `fic: deneme commit` were passing as conventional commits. Reason: the regex `\w+` accepted any word as a type. Fix: `allowedTypes` map added (Angular preset: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert), with type validation. For invalid types, returns "unknown type: fic". With `--verbose`, lists the commits being checked.
- [x] BUG: `latestVersionToTag()` hardcoded a `v` prefix instead of using the `tagName` template from config (2026-03-23) → In environments without a config file (default `tagName: "${version}"`), the tag was created as `0.1.0-main.0` while the changelog searched for `v0.1.0-main.0`. Fix: `latestVersionToTag()` now uses `renderTagName(tagNameTemplate, version)`. The `v` prefix is stripped from the version before applying the template, preventing `vv` duplication.
- [x] BUG: With changelog disabled, `git commit` produced "nothing to commit" error (2026-03-30) → With changelog and bumper disabled, there are no staged changes, but commit was attempted. Fix: `HasStagedChanges()` check added; if there are no staged changes, the commit is skipped and a verbose log entry is written.
- [x] BUG: When `tagName` config changed, old-format tags were being found as latest (2026-03-30) → Transitioning from `v${version}` → `${version}`, `GetLatestTag` still found `v1.5.0`, and the changelog searched for the `1.5.0` tag (which doesn't exist). Fix: `matchesTagNameFormat()` added for tag format filtering. A fallback mechanism for version continuity across format transitions, plus a raw tag fallback in changelog.
- [x] BUG: `hooks install` never pruned hooks removed from config (2026-08-25) → User report: preCommit was deleted from config, `hooks install` re-run, yet the pre-commit hook kept firing. `Install()` was purely additive — it only wrote configured hooks and never looked at previously installed managed scripts, and with an empty hooks section the CLI returned early before the installer ran at all, so `.hooks/<name>` + `core.hooksPath` stayed active forever. Fix: `Install()` now reconciles — managed hooks missing from config are deleted (`✓ Removed <name> (no longer in config)`), user-created hooks are never touched, hooks are written in deterministic `supportedGitHooks` order, `.hooks/` is no longer created when nothing is configured, and when a prune empties `.hooks/` entirely `core.hooksPath` is reset. CLI early-return removed so pruning runs even with an empty hooks section. 6 new unit tests + 2 new CLI tests.
- [x] BUG: `TestInstall_SkipsEmptyCommands` mutated the developer's repo git config (2026-08-25) → The test ran `Install()` without the `commandExecutor` mock while one hook (`{""}`) was actually written, so a real `git config core.hooksPath .hooks` executed in the test cwd — the release-it-go repo itself — silently disabling all repo git hooks (no `.hooks/` dir exists here). Found while investigating the prune bug: the repo's local config had the stray `core.hooksPath=.hooks` entry. Fix: `mockGitCommands(t)` added to the test; the polluted config entry was manually unset.
- [x] BUG: With `push: false`, "no upstream configured" error still appeared (2026-02-18) → `checkUpstream()` only looked at the `requireUpstream` flag, not the `push` state. In manually written configs with `push: false` and `requireUpstream` unspecified, the default `true` triggered the upstream check. The init wizard masked this by setting `requireUpstream = false`, but the actual check function had the bug. Fix: `!g.config.Push` check added inside `checkUpstream()`; when push is disabled, the upstream check is skipped. Test added.

---

## Change History

| Date | Developer | Change |
|------|-----------|--------|
| 2026-02-16 | - | Project started, PRD documents created |
| 2026-02-16 | Claude | Phase 1 complete: CLI, config, version, logger, template, tests |
| 2026-02-16 | Claude | Phase 2 complete: git runner, prerequisites, commit, tag, push, repo info, changelog, tests |
| 2026-02-16 | Claude | Phase 3 complete: conventional commit parser, bump analyzer, changelog renderers (conventional + keep-a-changelog), file update |
| 2026-02-16 | Claude | Phase 4 complete: GitHub + GitLab API client, release create, asset upload, comment, token management, GHE/CA cert support |
| 2026-02-16 | Claude | Phase 5 complete: bubbletea UI, lipgloss colors, spinner, CI detection, hook runner, pipeline orchestrator, dry-run, tests |
| 2026-02-16 | Claude | Phase 6 complete: bumper (JSON/YAML/TOML/INI/text), CalVer integration, CLI modes, pre-release flows, pipeline bump step |
| 2026-02-16 | Claude | Phase 7 complete: integration tests (17), coverage 80%+, CI/CD workflows, GoReleaser, shell completions, build info |
| 2026-02-16 | Claude | Security fix: HTTPS URL credential stripping, GoReleaser ldflags fix |
| 2026-02-16 | Claude | Config compat: npm release-it format compatibility (normalizeJSON, applyPluginCompat) |
| 2026-02-16 | Claude | feat: --preRelease shorthand flag, GitLab PreRelease field |
| 2026-02-16 | Claude | Real environment tests: GitLab CI pipeline (main + sub-branch prerelease) successful |
| 2026-02-16 | Claude | Phase 8 complete: init command, dual config support, legacy migration, smart config writer |
| 2026-02-16 | Claude | fix: first-release changelog error (0.0.0 tag not found), separate init wizard commit/tag/push |
| 2026-02-16 | Claude | Phase 9 complete: conventional commit linting, --check-commits, --ignore-commit-lint, pipeline integration |
| 2026-02-16 | Claude | Phase 10 complete: UI/Output improvements - icon constants, FormatBold, Logger.Print(), Verbose dim format, CI spinner icons, banner, printSummary lipgloss box |
| 2026-02-16 | Claude | fix: UI/Output improvements v2 - remove CI spinner duplicate line, init ✗ bug fix, "no commits" graceful exit, remove printSummary box, past-tense spinner messages |
| 2026-02-16 | Claude | fix: pre-release same-ID version not incrementing bug (prepatch → prerelease increment) |
| 2026-02-16 | Claude | fix: commit lint type validation - allowedTypes map rejects invalid types, --verbose shows commit list |
| 2026-02-16 | Claude | Phase 11 complete: Docker container support - multi-stage Dockerfile, .dockerignore, Makefile docker targets |
| 2026-02-16 | Claude | Phase 12 complete: Docker pre-flight checks - git identity check, token pre-flight check (GitHub/GitLab) |
| 2026-02-17 | Claude | Phase 13 complete: Webhook notification support - Slack + Teams, non-fatal pipeline step, urlRef security pattern, 98%+ coverage |
| 2026-02-17 | Claude | feat: "Write CHANGELOG.md file?" question added to init wizard - option to disable file writing while changelog is enabled |
| 2026-02-17 | Claude | fix: removed unnecessary "Required branch" question from init wizard - optional setting, can be added via config |
| 2026-02-17 | Claude | fix: removed unnecessary "Require new commits" question from init wizard - default behavior (true) is sufficient |
| 2026-02-17 | Claude | feat: requireConventionalCommits default set to true, question removed from init wizard |
| 2026-02-17 | Claude | feat: init --full-example command added - generates an example file with all config options |
| 2026-02-17 | Claude | feat: YAML config writing support + format selection (JSON/YAML) added to init wizard |
| 2026-02-17 | Claude | feat: init --full-example now generates commented YAML (.release-it-go-full.yaml) |
| 2026-02-17 | Claude | feat: MigrateLegacyConfigTo can output migration in the chosen format |
| 2026-02-17 | Claude | feat: DetectNativeConfigAny detects both JSON and YAML native configs |
| 2026-02-17 | Claude | feat: ForceFields ensures wizard-configured fields are written explicitly (even if default) |
| 2026-02-17 | Claude | fix: on format change (JSON→YAML) old config is backed up as .bak; double config prevented |
| 2026-02-17 | Claude | fix: auto-switch to CI mode when no TTY (Docker without -it) — go-isatty |
| 2026-02-18 | Claude | feat: CI/CD pipeline added, golangci-lint errors resolved, code quality tests added |
| 2026-02-18 | Claude | refactor: init wizard question order improved — format question moved to the end for better UX |
| 2026-02-18 | Claude | fix: upstream error with push false — checkUpstream wasn't checking push state |
| 2026-02-20 | Claude | Phase 15 complete: Branch-aware pre-release version detection - GetLatestPreReleaseTagMerged, GetLatestStableTagMerged, resolvePreReleaseBaseTag, 16 unit tests + 4 integration tests |
| 2026-02-21 | Claude | refactor: CLAUDE.md migrated to modular .claude/rules/ structure (6 rule files), Makefile improved (ldflags, coverage, vuln, check), Docker entrypoint info-only command support |
| 2026-02-21 | Claude | revert: GitLab CI_JOB_TOKEN changes reverted (ValidateToken /projects/:id + Job-Token header auto-detect) — CI_JOB_TOKEN doesn't have commit/tag/push permission, Project Access Token is required |
| 2026-03-23 | Claude | fix: latestVersionToTag uses tagName template instead of hardcoded v prefix — tag mismatch in environments without a config file resolved |
| 2026-03-23 | Claude | feat: warning message when no config file found, debug log of loaded config path |
| 2026-03-23 | Claude | test: resolvePreReleaseBaseTag + FormatBold tests (runner 79.2%, ui 71.8%) |
| 2026-03-23 | Claude | fix: code review findings - GitLab timeout, preRelease substring match, CA cert log, proxy log, dead code cleanup |
| 2026-03-23 | Claude | Phase 16 complete: GitLab nested group URL support + CalVer yyyy.mm.dd format fix |
| 2026-03-23 | Claude | Phase 17 complete: SSH port support, latestVersionToTag tests, empty release notes warning |
| 2026-03-23 | Claude | Phase 18 complete: LoadConfigFromBytes normalization, plugin override fix |
| 2026-03-23 | Claude | Phase 19 complete: Test coverage strengthening - 38 new tests, release 90.5%, git 90.5%, version 91.9% |
| 2026-03-30 | Claude | fix: empty commit error with changelog disabled - HasStagedChanges check |
| 2026-03-30 | Claude | fix: format-aware tag filtering on tagName config change - matchesTagNameFormat + fallback mechanism, 18 new tests |
| 2026-03-30 | Claude | feat: rich Teams MessageCard notifications - facts (Version, Last Release, Commits, Contributors), changelog section, configurable theme/image, GetCommitCountSinceTag + GetContributorsSinceTag, 3 new tests |
| 2026-03-30 | Claude | feat: ignoredContributors, themeColor, imageUrl added to webhook config - bot account filtering, messageTemplate override, full example YAML updated |
| 2026-03-31 | Claude | docs: Phase 20 PRD added (`docs/phase_20.md`) — git hook install command design |
| 2026-03-31 | Claude | feat: `release-it-go install` command — writes config hooks into .hooks/ directory, sets core.hooksPath (husky pattern) |
| 2026-03-31 | Claude | refactor: `install` command refactored into `hooks install` / `hooks remove` subcommand group |
| 2026-03-31 | Claude | test: hooks command and githook edge case tests added |
| 2026-03-31 | Claude | feat: `--check-msg` flag — single message validation for the commit-msg hook (Phase 20) |
| 2026-03-31 | Claude | fix: `--check-msg` output format standardized with `--check-commits` |
| 2026-03-31 | Claude | feat: commitlint-style professional output for `--check-msg` (compact + verbose modes) |
| 2026-04-01 | Claude | feat: `--check-msg` now supports string and stdin (`-`) input (in addition to file path) |
| 2026-04-16 | Claude | docs: documentation set created — CLAUDE.md expanded (~290 lines), ARCHITECTURE.md, DECISIONS.md added, PROGRESS.md updated with Phase 20 |
| 2026-04-16 | Claude | docs: full English translation of CLAUDE.md, ARCHITECTURE.md, DECISIONS.md, PROGRESS.md, .claude/rules/* — language consistency with README.md and code comments |
| 2026-04-16 | Claude | docs: CONTRIBUTING.md and TROUBLESHOOTING.md added — onboarding for new developers + common issue reference |
| 2026-04-17 | Claude | Phase 21 complete: P0 test coverage gaps closed — runCheckMsg, hooks install/remove, runInit dispatcher, runner github/gitlab dry-run integration; cli 58.8%→84.9%, runner 78.1%→83.5%; zero production code changes |
| 2026-04-21 | Claude | Phase 22 complete: --atomic added to default git.pushArgs to prevent orphan tags in parallel CI runs; behavior change is opt-out via config; coverage preserved |
| 2026-08-25 | Claude | fix: `hooks install` now prunes managed hooks removed from config (reconciliation model), resets core.hooksPath when .hooks/ empties, no longer creates .hooks/ with nothing configured; deterministic install order; test isolation fix (TestInstall_SkipsEmptyCommands ran real `git config` against the developer repo) |

---

## Future Improvements (Priority: Low)

- [ ] GitLab `ValidateToken()` endpoint could use `/projects/:id` instead of `/user` (for CI_JOB_TOKEN compatibility, though practical benefit is limited because CI_JOB_TOKEN cannot commit/push)
- [ ] Documentation for GitLab CI integration: `git remote set-url` setup, Project Access Token requirements, `GIT_DEPTH: 0` requirement

---

## Rules

1. **Update this file at the end of every session.**
2. Completed items are marked with `[x]`.
3. New items are added with `[ ]`.
4. Status field is updated: `Not Started` / `In Progress` / `Complete`
5. Progress percentage is updated.
6. Important decisions, blockers, or changes go in the Notes section.
7. New rows are added to the Change History table.
