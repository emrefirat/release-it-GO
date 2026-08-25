# release-it-go

> A Go reimplementation of [release-it](https://github.com/release-it/release-it) (npm). Git tag + GitHub/GitLab release + changelog + multi-file version bumping + webhook notifications, with no npm/Node.js dependency.

## Project Overview

A CLI tool that automates the release process as a single static Go binary. Computes semver/CalVer from conventional commits, creates GitHub/GitLab releases, generates changelogs, sends Slack/Teams webhooks, installs git hooks. Backward-compatible with the npm `release-it` config format.

## Tech Stack

| Decision | Choice | Version |
|----------|--------|---------|
| Language / Min | Go | 1.26.7+ |
| CLI framework | Cobra | v1.10.2 |
| Config | Viper (JSON + YAML + TOML) | v1.21.0 |
| Versioning | Masterminds/semver/v3 | v3.4.0 |
| TUI | Bubbletea + Lipgloss | v1.3.10 / v1.1.0 |
| Logger | log/slog (stdlib) | - |
| HTTP | net/http (stdlib) | - |
| Git | git CLI wrapper (exec.Command) | - |
| Version source | Git tag (primary), bumper input file (secondary) | - |
| Changelog | Conventional commits (Angular preset) + keep-a-changelog | - |
| Release platform | GitHub + GitLab REST API | - |
| Notification | Slack + Teams webhook | - |
| Plugin | NONE — everything is built-in (DECISIONS.md ADR-006) | - |
| Distribution | GitHub Releases (GoReleaser), `go install`, Docker | - |

## Architecture (Summary)

10-step pipeline orchestrator + lifecycle hook system. For full architecture → [ARCHITECTURE.md](ARCHITECTURE.md).

```
init → prerequisites → commitlint → version → bump → changelog → git:release → github:release → gitlab:release → notification
```

Every step supports `before:` / `after:` hooks. Dry-run is supported for all steps. The `notification` step is non-fatal (pipeline continues on failure).

## Directory Layout

```
cmd/release-it-go/
  main.go                  # Entry point (ldflags: version/commit/date)
internal/
  cli/                     # Cobra commands (root, init, hooks install/remove, version, completion)
  config/                  # Config loading, merge, write, migration, npm compat
  version/                 # Semver parse/increment, CalVer, branch-aware pre-release
  git/                     # Git CLI wrapper (commit, tag, push, prerequisites, repo info)
  githook/                 # Git hook installer (.hooks/ directory, core.hooksPath)
  changelog/               # Conventional commit parser, analyzer, lint, renderer
  release/                 # GitHub + GitLab REST client, asset upload, comment
  bumper/                  # Multi-file version updates (JSON/YAML/TOML/INI/text)
  hook/                    # Lifecycle hook runner (before:/after: events)
  notification/            # Webhook notifications (Slack + Teams MessageCard)
  ui/                      # Prompter (interactive/non-interactive), spinner, colors, CI detect
  runner/                  # Pipeline orchestrator + ReleaseContext (shared state)
  log/                     # Structured logger (slog wrapper, verbose levels)
test/integration/          # Real git repo integration tests
docs/phase_*.md            # Phase PRD documents (1-20)
```

## Code Conventions

### Naming
- Package: lowercase, single word (`config`, `git`, `release`, `githook`)
- Exported: PascalCase (`CreateRelease`, `ReleaseContext`)
- Unexported: camelCase (`renderTemplate`, `validateInput`)
- Interface: `-er` suffix (`Prompter`, `ReleaseProvider`)
- Test: `TestFunc_Scenario_Expected` (e.g., `TestParseRepoURL_HTTPS_StripsCredentials`)

### Error Handling Pattern (Project-Specific)
```go
// GOOD — wrap with %w, action-oriented message
result, err := someFunction()
if err != nil {
    return fmt.Errorf("creating GitHub release: %w", err)
}

// BAD — silent ignore
result, _ := someFunction()
```
- Wrap errors with `fmt.Errorf("...: %w", err)`; use `errors.Is`/`errors.As` to inspect root cause.
- Error messages should be actionable for users; don't leak internal details.
- No `panic` (only for unrecoverable `init()` situations — none in this project).

### Logging Pattern
```go
// internal/log package: 4 levels
logger.Print("user-facing (always shown)")          // banner, summary
logger.Verbose("    ↳ detail (only -V)")            // hook commands, git commands
logger.Debug("internal trace (only -VV)")           // exec details
logger.DryRun("[dry-run] would do but didn't")      // dry-run mode
logger.Warn("warning (always shown)")
logger.Error("error (always shown)")
```

### Git Command Mocking (Test Pattern)
Git operations run through a `commandExecutor` function variable. Defined separately in `internal/git/git.go` and `internal/githook/githook.go`. Mock in tests:
```go
original := commandExecutor
defer func() { commandExecutor = original }()
commandExecutor = func(name string, args ...string) (string, error) {
    return "v1.2.3", nil
}
```

### Interface Location
**Consumer-side**: Interfaces are defined where they're used; implementations may live in another package. Examples:
- `ui.Prompter` — used by `runner` and `cli`, defined in `ui` (UI side makes sense, has both interactive and CI implementations)
- `release.ReleaseProvider` — implemented by GitHub and GitLab clients, defined in `release`

### Context Propagation
Currently the pipeline doesn't use `context.Context` (CLI tool, short-lived, OS Ctrl+C is enough). HTTP requests use their own `*http.Client` timeout. If long-lived/cancellable operations are added later → introduce `context.Context` propagation.

## Commands

```bash
# Development
make check          # fmt + vet + lint + vuln + test + build (REQUIRED before commit)
make build          # Build binary (with version/commit/date injected via ldflags)
make test           # All tests (-v -cover -race)
make test-unit      # internal/ tests only
make test-integration  # test/integration/ only (real git repo)
make coverage       # HTML coverage report (coverage.html)
make lint           # golangci-lint
make vuln           # govulncheck security scan
make tidy           # go mod tidy + verify

# Docker
make docker-build VERSION=x.y.z
make docker-run

# Single test / specific scenario
go test ./internal/runner/ -run TestRunner_Run -race -v
go test ./internal/runner/ -run TestRunner_Run/scenario_name -race -v
```

## Adding New Features

### Adding a new pipeline step
1. Add a `{name, fn}` entry to the `pipelineStep` slice in `internal/runner/runner.go`
2. Write the step function: `func (r *Runner) yourStep() error { ... }`
3. Use the spinner: `r.ctx.Spinner.Start("...")` → `r.ctx.Spinner.Stop(true|false)`
4. Dry-run support: check `r.ctx.IsDryRun`, don't perform side effects
5. Before/after hook support comes for free (the runner handles it)
6. Non-fatal? See `sendNotification` — error log + `return nil`
7. Add an integration test: `test/integration/release_test.go`

### Adding a new CLI flag
1. `internal/cli/root.go` — declare the flag variable, register it in `NewRootCommand()`
2. Flag names use kebab-case: `--pre-release`, `--dry-run`
3. For boolean disable, use `--no-` prefix: `--no-git.push`
4. Every flag requires a clear `Usage` string
5. Wire the flag into `runRelease()` or the relevant runner method
6. Test: `internal/cli/root_test.go`

### Adding a new config field
1. `internal/config/config.go` — add the field to the struct, **JSON + YAML + TOML + mapstructure tags are required**
2. `internal/config/defaults.go` — set the default value
3. `internal/config/template.go` or `internal/cli/init.go` if the wizard is affected
4. `internal/config/writer.go` — add commented entry to the `fullExampleYAML` constant
5. Test: `internal/config/config_test.go`, `defaults_test.go`

### Adding a new git command wrapper
1. Add to the appropriate file under `internal/git/` (commit.go, tag.go, push.go, repo.go)
2. Use `g.run(args...)` or `g.runSilent(args...)`
3. Is the new command a write operation? Update the `readOnlyCommands` map in `internal/git/git.go`
4. Test: `commandExecutor` mock pattern in the relevant `*_test.go`

### Adding a new release platform
1. `internal/release/your_platform.go` — implement the `ReleaseProvider` interface
2. `NewYourPlatformClient()` constructor: use `getToken(tokenRef, skipChecks)`
3. `internal/config/config.go` — add a `YourPlatformConfig` struct
4. `internal/runner/runner.go` — add a new `r.yourPlatformRelease()` method, wire into the pipeline

## DO / DON'T (Project-Specific)

### DO
- Before adding a new dependency, **check the docs via Context7 MCP** (`resolve-library-id` + `query-docs`)
- Wrap every external call (git, HTTP) error with context
- Use the `tokenRef` pattern for tokens (env var name in config, value from env)
- Use the `urlRef` pattern for webhook URLs (env var name in config)
- **Apply credential stripping** for HTTPS remote URLs (regex already exists in `parseRepoURL`)
- Set a **30s+ timeout** for new HTTP clients (`config.Timeout` or `defaultTimeout`)
- Use Conventional Commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`, `perf:`
- Test coverage **min 70%**, critical packages (git, runner, release) **85%+** target
- Run `make check` before every commit (fmt + vet + lint + vuln + test + build)
- Update PROGRESS.md at the end of every session

### DON'T
- Use `panic` (except for unrecoverable init situations — none in this project)
- Log tokens/secrets/PII
- Ignore errors: `result, _ := ...` is forbidden
- Add side effects in `init()` functions (config init, global state)
- Use global state — solve via DI (through the config struct)
- Use `panic`/`os.Exit` in runner or internal packages — only in `cli.Execute()`
- Run git commands through a shell (`sh -c "git ..."`) — command injection risk (use `exec.Command("git", args...)`)
- Break npm release-it config compatibility — backward compat (`internal/config/compat.go`) must be preserved
- Violate `.claude/rules/` conventions — `make check` will catch most via lint

## Critical Domain Concepts

- **Conventional Commit**: `type(scope): description` format. `feat:` → minor, `fix:` → patch, `BREAKING CHANGE:` or `feat!:` → major. Allowed types: see `internal/changelog/lint.go`.
- **Pre-release**: Semver with pre-release ID: `1.2.3-beta.0`. Branch-aware (Phase 15) — only considers tags merged into HEAD.
- **CalVer**: Calendar version (`yy.mm.minor` default). Cannot be combined with SemVer pre-release.
- **Bumper**: Writing the version into multiple files (and reading from an input file). Supports JSON/YAML/TOML/INI/plain text.
- **Lifecycle Hook**: Shell commands that run on pipeline events like `before:bump`, `after:release` (`internal/hook/`).
- **Git Hook**: Scripts that run on git events like `pre-commit`, `commit-msg`, `pre-push` from the `.hooks/` directory (`internal/githook/`, Phase 20).
- **Token/URL Ref Pattern**: Sensitive values are not stored in config. Only the env var name is stored (`tokenRef: "GITHUB_TOKEN"`); the value is read at runtime via `os.Getenv()`.

## Required Practices

1. **Use Context7**: For new libraries / unsure existing library API, use `resolve-library-id` + `query-docs`. Don't write code based on guesses.
2. **Update PROGRESS.md**: At the end of every session, write down what you did, what you completed, and any bugs. For a new phase, also create `docs/phase_N.md` PRD.
3. **Security scan**: `make check` (which includes govulncheck) before every commit. Run `gosec ./...` for new HTTP clients / file operations / `exec.Command`.
4. **Pre-commit checklist**: Don't commit if `make check` fails.
5. **Conventional commits required**: `requireConventionalCommits: true` (default). `--ignore-commit-lint` can bypass it but is discouraged.

## Detailed Rules — `.claude/rules/`

Modular rule files. **Read these when the summary above isn't enough**:

| File | When to Read |
|------|--------------|
| [`.claude/rules/best-practices.md`](.claude/rules/best-practices.md) | New dependency, interface design, concurrency/HTTP/Docker patterns |
| [`.claude/rules/code-quality.md`](.claude/rules/code-quality.md) | Naming, error handling, logging, performance details |
| [`.claude/rules/git-workflow.md`](.claude/rules/git-workflow.md) | Commit conventions, branch strategy, pre-commit checklist |
| [`.claude/rules/testing.md`](.claude/rules/testing.md) | Test patterns, mock setup, coverage targets, table-driven format |
| [`.claude/rules/security.md`](.claude/rules/security.md) | Token management, when to run gosec/govulncheck, command injection safety |
| [`.claude/rules/progress-tracking.md`](.claude/rules/progress-tracking.md) | PROGRESS.md format, bug tracking, end-of-session process |

## Other Important Documents

| File | Description |
|------|-------------|
| [`PROGRESS.md`](PROGRESS.md) | Phase tracking, bug log, change history — **update at end of every session** |
| [`ARCHITECTURE.md`](ARCHITECTURE.md) | System architecture, pipeline detail, package dependencies, external system integration |
| [`DECISIONS.md`](DECISIONS.md) | ADR (Architecture Decision Records) — why Cobra, why git CLI, why no plugin system |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | New contributor onboarding (tools, workflow, PR checklist) |
| [`TROUBLESHOOTING.md`](TROUBLESHOOTING.md) | Common errors and fixes (user + developer sections) |
| [`README.md`](README.md) | User documentation (install, CLI usage, config examples) |
| [`docs/phase_N.md`](docs/) | PRD (Product Requirements Document) for each phase |

## Important Build / Config Files

| File | Description |
|------|-------------|
| `Makefile` | Build, test, lint, docker — version injection via ldflags |
| `.goreleaser.yaml` | Multi-platform release (linux/darwin/windows × amd64/arm64), nfpm (deb/rpm/apk) |
| `Dockerfile` | Multi-stage (golang:1.26.7-alpine builder + alpine:3.21 runtime), non-root user (releaser:1000), static binary |
| `docker-entrypoint.sh` | Git identity check for Docker, info-only command support |
| `.github/workflows/ci.yml` | CI: build + vet + test + race + coverage + lint (security job temporarily disabled) |
| `.github/workflows/build.yml` | Multi-platform build via GoReleaser |
| `.github/workflows/release.yml` | Self-hosting: release-it-go creates the tag, GoReleaser builds the binary release |
| `.golangci.yml` or golangci-lint config | Not present — uses defaults |

## Working Rules for Claude

1. **Read the relevant existing file before writing code.** Mimic the pattern; don't introduce new style.
2. **Conform to existing patterns before suggesting new ones.** Examples: spinner Start/Stop, error wrap, commandExecutor mock.
3. **Don't assume when unsure — ask.** Especially for config field changes, breaking changes, or anything affecting npm compat.
4. **Break large refactors into chunks; ask for approval.** Question single PRs with 500+ lines of changes.
5. **Update/add the relevant test after every change.** If no test exists, write the test first, then the fix.
6. **Don't propose a commit without running `make check`.** Don't commit if there are lint errors.
7. **Update PROGRESS.md.** New feature/fix/bug → add a row to the change history.
8. **DO NOT add co-author lines.** User preference (memory: feedback_no_coauthor).
9. **Read `.claude/rules/` for details.** The table above tells you which file is needed when.

## Document Metadata

- Last updated: 2026-04-16
- Current phase: Phase 20 complete (git hook installer + commit-msg hook validation)
- For the next update, check: `docs/phase_*.md` (new phase?), `internal/cli/root.go` (new command/flag?), `go.mod` (version change?), `.github/workflows/` (CI changes?)
