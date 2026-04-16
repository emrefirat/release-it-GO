# DECISIONS — release-it-go

> Architecture Decision Records (ADR). Answers "why is the code like this?"
>
> Format: ADR-NNN: Decision / Date / Status / Context / Decision / Alternatives / Consequences

Last updated: 2026-04-16 | See also: [ARCHITECTURE.md](ARCHITECTURE.md), [CLAUDE.md](CLAUDE.md)

---

## ADR-001: Cobra as the CLI Framework

**Date**: 2026-02-16 (Phase 1)
**Status**: Accepted

### Context
Several CLI frameworks exist in the Go ecosystem: `cobra`, `urfave/cli`, `kingpin`, stdlib `flag`. We need subcommands, flags, completion, help generation, and shared persistent flags. release-it on npm uses `commander` — user expectation is similar DX.

### Decision
Use `spf13/cobra` v1.10.2.

### Alternatives
- **stdlib `flag`**: No subcommand support, lots of manual setup. **REJECTED**.
- **urfave/cli**: Lighter but completion/help generation is weaker compared to Cobra. **REJECTED**.
- **kingpin**: Unmaintained, last release was a long time ago. **REJECTED**.

### Consequences
**Positive**: Subcommand tree (root + init + version + completion + hooks), automatic shell completion, natural Viper integration.
**Negative**: Cobra is a heavy dependency that increases binary size slightly (acceptable — release automation CLI is used in Docker/CI, size is not critical).

---

## ADR-002: `exec.Command("git", ...)` for Git Operations (Instead of go-git)

**Date**: 2026-02-16 (Phase 2)
**Status**: Accepted

### Context
Go ecosystem has `go-git` (pure Go implementation) for git operations. The alternative is calling the git binary already on the user's system via exec.Command.

### Decision
Call the git CLI binary via exec.Command. The user must have git installed on the system.

### Alternatives
- **go-git**: Pure Go, no external dependency. **REJECTED**, for these reasons:
  - GPG signing support is limited (the user's gpg config doesn't work)
  - Doesn't run git hooks (e.g., pre-push) — release feature breaks
  - Doesn't use credential helpers (osxkeychain, gnome-keyring)
  - HTTPS proxy and `~/.gitconfig` behavior cannot be fully mimicked
  - LFS, partial clone, submodule behavior differs

### Consequences
**Positive**: The user's git config (signing, credentials, hooks) just works. SSH keys, GPG, GitHub CLI auth are used automatically. Low risk of unexpected behavior.
**Negative**: Git installation is required (we added `apk add git` to the Docker image). Parsing command output is string-based (text → struct); the git CLI output format may change slightly between versions.

### Mitigation
- `commandExecutor` function variable makes test mocking easy
- `IsGitInstalled()` prerequisite check
- `isWriteOperation()` map for dry-run safety

---

## ADR-003: Masterminds/semver/v3 as the Versioning Library

**Date**: 2026-02-16 (Phase 1)
**Status**: Accepted

### Context
We need semver parse/increment/compare. Pre-release IDs (`beta`, `alpha`, `rc`), wildcard constraints (`>=1.2.3, <2.0.0`), build metadata support.

### Decision
Use `Masterminds/semver/v3` v3.4.0.

### Alternatives
- **No stdlib option**: Go stdlib has a semver package (`golang.org/x/mod/semver`) but only for compare/parse, not increment. **REJECTED**.
- **blang/semver**: Unmaintained (last release 2019). **REJECTED**.

### Consequences
**Positive**: `IncMajor/IncMinor/IncPatch`, `Prerelease()`, constraint matching, npm-compatible behavior.
**Negative**: `IncPatch()` strips pre-release (`1.2.3-beta.0 → 1.2.3`) — this is correct semver behavior but can be unexpected. Bypassed via `incrementPreRelease()` (Phase 6 bug fix).

### Related
- ADR-005 (CalVer) — cannot be combined with semver, mutex.

---

## ADR-004: JSON + YAML + TOML as Config Formats (via Viper)

**Date**: 2026-02-16 (Phase 1) + 2026-02-17 (Phase 14 YAML write)
**Status**: Accepted

### Context
Users come from different ecosystems:
- npm release-it users → JSON
- Go ecosystem → YAML (k8s, GitHub Actions influence) and TOML (Cargo, Hugo influence)
- Forcing a single format reduces adoption

### Decision
Support JSON + YAML + TOML equally. Auto-detect via Viper (by extension).

### Alternatives
- **JSON only**: Excellent npm compatibility, but Go developers expect YAML/TOML. **REJECTED**.
- **YAML only**: Popular due to GitHub Actions, but npm migration is hard. **REJECTED**.

### Consequences
**Positive**: High adoption, direct migration from npm release-it (`.release-it.json` is read directly).
**Negative**: Viper is a heavy dependency (~indirect deps). Native package JSON + manual YAML/TOML alternatives could have been tried, but Viper's features (env var override, multi-source merge) saved a lot of code.

### Side Decision
**Writing**: The init wizard outputs JSON or YAML (no TOML write — TOML writing in Viper is not mature, manual writer would be complex). YAML output was chosen for comment support (`fullExampleYAML`) (Phase 14).

---

## ADR-005: Calendar Versioning (CalVer) Built-in

**Date**: 2026-02-16 (Phase 1) + 2026-03-23 (Phase 16 fix)
**Status**: Accepted

### Context
Some projects use CalVer instead of semver (Ubuntu 22.04, JetBrains 2026.1). `release-it` on npm supported it via the calver plugin.

### Decision
Accept CalVer as a built-in feature. Cannot be combined with SemVer (requires `calver.enabled: true` in config). Format string: `yy.mm.minor`, `yyyy.mm.dd`, custom.

### Alternatives
- **Skip CalVer**: The Go ecosystem mostly uses semver. **REJECTED** — niche but real user group exists, and since there's no plugin system (ADR-006), it must be built-in.

### Consequences
**Positive**: Both semver and calver in a single binary — the user doesn't have to choose a library.
**Negative**: Added a `calver` section to config (noise for the 95% who never use it). A mutex check (`if cfg.CalVer.Enabled && cfg.PreReleaseID != ""`) was needed — semver and calver cannot be mixed.

---

## ADR-006: NO Plugin System — Everything Built-in

**Date**: 2026-02-16 (Phase 1 PRD)
**Status**: Accepted (intentional)

### Context
npm `release-it` has a plugin system (`@release-it/conventional-changelog`, `@release-it/bumper`, etc.). This makes sense in the Node.js world — npm install can dynamically load. In Go, plugins (`plugin.Open`) are limited to Linux/macOS, no Windows, distributing .so/.dylib is complex.

### Decision
NO plugin system. All features are built into the binary:
- Bumper, changelog, notification, calver, GitHub, GitLab
- Extension path: lifecycle hooks (shell commands)

### Alternatives
- **Go plugin**: Platform-limited, no Windows, ABI difficulties. **REJECTED**.
- **WASM plugin**: Early (2026), wasmtime/wazero infrastructure required, user complexity. **REJECTED**.
- **gRPC plugin (HashiCorp pattern)**: Dependency overhead, IPC complexity. **REJECTED**.

### Consequences
**Positive**: Single binary, single install, single update path. Cross-platform without issues. Dependencies managed centrally.
**Negative**: Adding a new feature requires going through the request/approval process (community PRs or fork). Users can't "write my own plugin" — they use hooks + shell scripts instead.

### Mitigation
The lifecycle hook system (`before:bump`, `after:release`, etc.) replaces plugins. Users add their own logic with `before:bump: ["./my-script.sh"]`.

---

## ADR-007: npm release-it Config Backward Compatibility

**Date**: 2026-02-16 (Phase 8)
**Status**: Accepted

### Context
release-it (npm) has been used for years; thousands of projects have a `.release-it.json`. Migration friction would kill adoption.

### Decision
Support `.release-it.json` (legacy) and `.release-it-go.json` (native) simultaneously. Native priority. Convert npm release-it's type inconsistencies (`hooks: false`, `plugins.bumper: ["..."]`) via `internal/config/compat.go`.

### Alternatives
- **Native only**: Faster implementation, but every user must migrate. **REJECTED**.
- **Migrate-only mode**: `release-it-go migrate` command, then read native only. **PARTIALLY ACCEPTED** — the `init` command offers migration, but we still support legacy config read-only.

### Consequences
**Positive**: Drop-in replacement for existing `.release-it.json` files. Minimal migration pain.
**Negative**: `compat.go` has maintenance overhead (as npm release-it features evolve). We're not arming ourselves: we don't auto-pick newly added npm release-it features; manual tracking is required.

### Related
- Phase 18 fix: normalization was added to `LoadConfigFromBytes` too (plugin override fix).

---

## ADR-008: Bubbletea + Lipgloss Interactive UI

**Date**: 2026-02-16 (Phase 5)
**Status**: Accepted

### Context
Need interactive prompts (version select, confirm, input). Non-interactive default answers in CI mode. Cross-platform (including Windows ConHost).

### Decision
Use `charmbracelet/bubbletea` (TUI framework) + `lipgloss` (styling) + `bubbles` (widgets).

### Alternatives
- **AlecAivazis/survey**: Mature, simple. **REJECTED**: maintenance has slowed (charm ecosystem more active), Windows TTY support is weak.
- **manifoldco/promptui**: Unmaintained. **REJECTED**.
- **Manual termios + ANSI**: Lots of work, platform-specific bug risk. **REJECTED**.

### Consequences
**Positive**: Modern, actively developed, cross-platform incl. Windows works without issues. Spinner, prompt, list select all available (bubbles).
**Negative**: Indirect dependency growth (charmbracelet/x/*, muesli/*). Binary size grew slightly (~5 MB).

### Side Decision
Defined a `Prompter` interface, with `InteractivePrompter` (bubbletea) and `NonInteractivePrompter` (CI auto-default) as two implementations. CI detection is via env vars + `go-isatty`.

---

## ADR-009: `tokenRef` / `urlRef` Pattern for Tokens / Webhook URLs

**Date**: 2026-02-16 (Phase 4) + 2026-02-17 (Phase 13 webhook)
**Status**: Accepted

### Context
Writing sensitive values into the config file is unsafe (gets committed, can leak into logs, leak risk). The common approach: env var.

### Decision
Only the env var **name** is stored in config:
```yaml
github:
  tokenRef: "GITHUB_TOKEN"
notification:
  webhooks:
    - urlRef: "SLACK_WEBHOOK_URL"
```
Read at runtime via `os.Getenv(ref)`. `getToken()` helper for validation.

### Alternatives
- **Direct token field**: `token: "ghp_..."`. **REJECTED** — security, accidental commit risk.
- **Secret manager integration (Vault/AWS Secrets Manager)**: Complexity. **REJECTED** — the user can do this with hooks if needed (`before:init: ["export GITHUB_TOKEN=$(vault read ...)"]`).
- **Encrypted config**: Key management problem (where do you read the key from?). **REJECTED**.

### Consequences
**Positive**: No risk of leaking secrets into config. Standard CI/CD pattern (env var). Easy migration.
**Negative**: If the user forgets to set the env var, "X env var not set" error. To counter, a prerequisite check (`checkTokens()`) was added.

---

## ADR-010: Branch-Aware Pre-Release (`git tag --merged HEAD`)

**Date**: 2026-02-20 (Phase 15)
**Status**: Accepted

### Context
Problem: When working on `feature-x` branch with `--preRelease=deneme`, the `v2.0.0-deneme.5` tag from `feature-y` branch was being picked up and breaking `feature-x`'s series (`feature-x` would want `v2.0.0-deneme.6` but `v1.2.0-deneme.0` was expected).

### Decision
Use `git tag -l --merged HEAD --sort=-v:refname` — only look at tags merged into HEAD. The `resolvePreReleaseBaseTag()` algorithm:
1. Find the latest pre-release tag (matching ID) merged into HEAD
2. Find the latest stable tag merged into HEAD
3. Pre-release base >= stable → continue series
4. Otherwise → start a new series from stable

### Alternatives
- **Tag namespacing (per branch name)**: `feature-x-v1.2.0-deneme.0`. **REJECTED** — non-conventional tags break downstream tooling.
- **Existing behavior (most recent tag)**: Bug persists. **REJECTED**.

### Consequences
**Positive**: 3 scenarios supported: long-living branch (continue series), deleted/recreated branch (new series), master advanced but branch isolated (continue series).
**Negative**: `--merged HEAD` query is an extra git command (perf impact minimal). If PreReleaseID is empty (standard release), the old behavior is preserved — backward compat.

---

## ADR-011: Webhook Notification Non-Fatal

**Date**: 2026-02-17 (Phase 13)
**Status**: Accepted

### Context
If a webhook (Slack/Teams) fails, should the pipeline stop? The release was already done; only the notification didn't go out.

### Decision
The notification step is **non-fatal**. On error: `logger.Warn(...)` + `return nil` (pipeline continues, exit code 0).

### Alternatives
- **Fatal**: A webhook failure kills the entire pipeline. **REJECTED** — the release was already successful; the user shouldn't lose the release.
- **Configurable**: `notification.fatal: true/false`. **POSTPONED** — no one needs it yet, easy to add later.

### Consequences
**Positive**: Wrong URL/network issues for webhooks don't break the release.
**Negative**: Risk of silent failure (the user expecting a Slack message might not see it). Mitigation: warn log is always shown, more detail with `-V`.

---

## ADR-012: `.hooks/` + `core.hooksPath` for Git Hooks (Instead of `.git/hooks/`)

**Date**: 2026-03-31 (Phase 20)
**Status**: Accepted

### Context
Phase 20: Install hook definitions from config as git hooks via `release-it-go install`. Two location options:
- `.git/hooks/` — git's default, but `.git/` is not versioned (every clone needs reinstall)
- `.hooks/` (project root) — versionable, but git doesn't read it by default

### Decision
Use `.hooks/` (project root) and point git to it via `git config core.hooksPath .hooks`. Husky pattern.

### Alternatives
- **`.git/hooks/`**: Manual reinstall on every clone, no sharing. **REJECTED**.
- **`.githooks/`**: A common alternative. **CHOSE INSTEAD** `.hooks/` — shorter, more neutral name than husky.

### Consequences
**Positive**: Hooks are committed to the repo and shared with the team. `release-it-go hooks install` sets `core.hooksPath` once per clone.
**Negative**: Have to remember to add to `.gitignore` (a new dev may not be aware).

### Side Detail
- A managed header (`# Managed by release-it-go — DO NOT EDIT`) distinguishes from user hooks
- `--force` overwrites existing non-managed hooks
- `hooks remove` only removes managed ones, leaves user hooks

---

## ADR-013: --check-msg Multi-Input Mode (file / stdin / string)

**Date**: 2026-04-01 (Phase 20 enhancement)
**Status**: Accepted

### Context
The `commit-msg` git hook calls `release-it-go --check-msg "$1"` ($1 = commit message file path). But the user may want to test manually: `release-it-go --check-msg "feat: test"` or `echo "feat: test" | release-it-go --check-msg -`.

### Decision
Support 3 input modes via a single `--check-msg` flag:
1. File path (if exists) → read from file
2. `-` → read from stdin
3. Direct string → use the string as-is

Distinction: `fileExists(input)` distinguishes file from string.

### Alternatives
- **3 separate flags**: `--check-msg-file`, `--check-msg-stdin`, `--check-msg-string`. **REJECTED** — unnecessary API surface.
- **File only**: Manual testing impossible. **REJECTED**.

### Consequences
**Positive**: A single flag, flexible usage. Works for both hooks and manual testing.
**Negative**: `fileExists()` check fragility — if a file named "test" exists one day, it's read as a file rather than a string. Mitigation: this case is rare, documented.

---

## Open / Discussed Decisions

Decisions that may be needed later:

- **ADR-XXX: Should `context.Context` propagation be added?** — Not currently; OS Ctrl+C is enough. If long-lived/cancellable operations are added (parallel multi-platform release), it must be added.
- **ADR-XXX: Plugin system (gRPC/WASM)** — Re-evaluate if user demand is high. Currently hooks + shell scripts are sufficient.
- **ADR-XXX: GitLab `ValidateToken()` `/projects/:id` endpoint** — Currently `/user`, incompatible with CI_JOB_TOKEN. Practical benefit is low because CI_JOB_TOKEN can't commit/push; postponed.
- **ADR-XXX: Bitbucket Server / Gitea / Forgejo support** — Easily added via `ReleaseProvider` interface, when needed.
