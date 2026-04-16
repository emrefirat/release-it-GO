# Contributing to release-it-go

Thanks for your interest! This guide gets you from `git clone` to your first merged PR.

> Already familiar with the project? Jump to [Development Workflow](#development-workflow).

---

## Prerequisites

You need these installed locally:

| Tool | Version | Why |
|------|---------|-----|
| Go | 1.26.2+ | Build the binary |
| Git | 2.x+ | Required at runtime (release-it-go wraps the git CLI) |
| Make | any | Run `make check`, `make build`, etc. |
| golangci-lint | latest | `make lint` and `make check` |
| govulncheck | latest | `make vuln` and `make check` |
| Docker | 20+ (optional) | Only if you change Docker-related code |
| GoReleaser | v2+ (optional) | Only if you change `.goreleaser.yaml` |

### Install the tooling

```bash
# Verify Go version
go version  # must be 1.26.2 or newer

# golangci-lint (required for `make check`)
brew install golangci-lint                    # macOS
# or:
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# govulncheck (required for `make check`)
go install golang.org/x/vuln/cmd/govulncheck@latest

# gosec (optional, used for ad-hoc security audits)
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

---

## First-Time Setup

```bash
git clone https://github.com/emrefirat/release-it-GO
cd release-it-GO

# Verify everything works before you change anything
make check
```

`make check` runs `fmt + vet + lint + vuln + test + build`. If this fails on a clean checkout, file an issue — that's a broken `main`, not your problem.

---

## Development Workflow

### 1. Pick or open an issue

- Bugs: check `PROGRESS.md` → "Bugs" section first; the project tracks them inline
- Features: open an issue describing the use case before writing code (especially if it changes config schema or pipeline order)

### 2. Create a branch

```bash
git checkout -b feat/short-description
# or fix/, refactor/, docs/, test/, chore/
```

Branch names follow the conventional commit prefix you'd use.

### 3. Read before you write

This project has strong conventions. Skim these before changing code:

- `CLAUDE.md` — project conventions, naming, error handling, mock pattern
- `ARCHITECTURE.md` — pipeline, package responsibilities, interfaces
- `DECISIONS.md` — why things are the way they are (read before suggesting alternatives)
- `.claude/rules/*.md` — detailed rules per topic

If you're touching a specific package, read the package's existing tests first — they show the mock pattern and expected style.

### 4. Make changes

Follow the **"Adding a new ..." flows** in `CLAUDE.md`:
- New CLI flag → see "Adding a new CLI flag"
- New config field → see "Adding a new config field"
- New pipeline step → see "Adding a new pipeline step"
- New release platform → see "Adding a new release platform"

### 5. Test

```bash
# Unit tests for what you changed
go test ./internal/your-package/ -race -v

# Full test suite
make test

# Integration tests (slower, runs real git commands)
make test-integration

# Coverage report (HTML)
make coverage   # opens coverage.html
```

**Coverage targets**: 70% minimum, 85%+ for critical packages (`git`, `runner`, `release`, `version`).

### 6. Validate

```bash
make check   # fmt + vet + lint + vuln + test + build
```

This **must pass** before you commit. The CI runs the same checks.

### 7. Commit

This project enforces [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: fix bug in X
refactor: restructure Y
test: add tests for Z
docs: update CONTRIBUTING.md
chore: bump dependency
perf: speed up version detection
ci: update GitHub Actions workflow
```

Commits that fail conventional commit lint will block release, so the project lints them too:

```bash
# Validate a commit message before committing
./bin/release-it-go --check-msg "feat: add user login"

# Validate all commits since last tag
./bin/release-it-go --check-commits
```

If you have hooks installed (`./bin/release-it-go hooks install`), the `commit-msg` hook validates this automatically.

**Atomic commits**: one logical change per commit. Don't mix a refactor with a feature.

### 8. Update PROGRESS.md

This is a **project rule** (see `.claude/rules/progress-tracking.md`):

- Add an entry to the "Change History" table
- If you fixed a bug, add it to the "Bugs" section as `[x]` with a short cause/fix summary
- If you completed a phase, mark it complete and add notes

### 9. Open a PR

- Title: same conventional commit format as your main commit
- Description: what changed, why, and how you tested it
- Link the issue if there is one
- Check the CI passes (it runs `build + vet + test + race + lint`)

---

## Code Style

### Naming
- Packages: lowercase, single word (`config`, `git`, `release`)
- Exported: `PascalCase`
- Unexported: `camelCase`
- Interfaces: `-er` suffix (`Prompter`, `ReleaseProvider`)
- Tests: `TestFunc_Scenario_Expected`

### Error handling
```go
// Good — wrap with context
if err != nil {
    return fmt.Errorf("creating GitHub release: %w", err)
}

// Bad — silent ignore
result, _ := someFunc()
```

### Logging
Use the `internal/log` wrapper, not `fmt.Println`. Levels: `Print`, `Verbose` (`-V`), `Debug` (`-VV`), `DryRun`, `Warn`, `Error`.

### Testing git operations
The project uses a `commandExecutor` function variable for mocking. Pattern:

```go
original := commandExecutor
defer func() { commandExecutor = original }()
commandExecutor = func(name string, args ...string) (string, error) {
    return "v1.2.3", nil
}
```

---

## Pull Request Checklist

Before requesting review:

- [ ] `make check` passes locally
- [ ] Tests added/updated for your change
- [ ] Coverage didn't drop below the threshold for the package you touched
- [ ] `PROGRESS.md` updated (Change History + Bugs if applicable)
- [ ] Commit messages follow Conventional Commits
- [ ] PR title is short, clear, and matches the main commit
- [ ] If you touched config schema → backward compat preserved (`internal/config/compat.go`)
- [ ] If you added a security-sensitive code path → ran `gosec ./...`

---

## Project-Specific Gotchas

These are real things that have bitten contributors. Read these before they bite you.

1. **Don't mock the database/git in tests when you can use the real thing.** Integration tests use `t.TempDir()` + real `git init`. Mocks are fine for unit tests but integration tests must hit real git.

2. **`commandExecutor` is defined in two places** (`internal/git/git.go` and `internal/githook/githook.go`). Mock the right one for the package you're testing.

3. **`PROGRESS.md` is the source of truth for project status**, not git history alone. Keep it current.

4. **No plugin system, on purpose.** See `DECISIONS.md` ADR-006. Don't suggest one without reading it first.

5. **`.claude/rules/` is detailed rules.** `CLAUDE.md` is the summary. When in doubt, the rule files win.

6. **Pre-release versioning is branch-aware** (Phase 15). Test pre-release changes on multiple branches if you touch `internal/version/` or `resolvePreReleaseBaseTag`.

7. **`tagName` template changes are tricky.** Format transitions (`v${version}` → `${version}`) need fallback logic — see `matchesTagNameFormat` and the bug log entries from 2026-03-30.

---

## Where to Ask

- Project status / current work: `PROGRESS.md`
- "Why is it like this?": `DECISIONS.md`
- "How does X work?": `ARCHITECTURE.md`
- "How do I do X?": `CLAUDE.md` ("Adding a new ..." sections)
- Issues / bugs: GitHub Issues
- Stuck on something common: [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

---

## Release Process (Maintainers Only)

This project releases itself with itself (self-hosting). See `.github/workflows/release.yml`:

1. release-it-go creates the tag (`chore(release): release v0.1.x`)
2. GoReleaser is triggered by the tag and builds multi-platform binaries
3. GitHub Release is published with assets

To cut a release locally:

```bash
GITHUB_TOKEN=... ./bin/release-it-go --ci
```

For the full reference, read the workflow files in `.github/workflows/`.
