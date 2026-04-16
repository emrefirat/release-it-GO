# Troubleshooting

Common errors and how to fix them. Split by audience: **users** (running release-it-go) vs **developers** (working on the codebase).

---

## User Issues

### `tag already exists`

**Cause**: A tag with the target name is already in the repo (local or remote).

**Fix**:
```bash
# Check existing tags
git tag -l | tail

# Delete a local tag
git tag -d v1.2.3

# Delete a remote tag (be careful — coordinate with team)
git push origin --delete v1.2.3
```

If the tag was pushed and downloaded, deleting it may break consumers. Prefer bumping to the next version instead.

---

### `no upstream configured`

**Cause**: The current branch has no upstream tracking branch, but `git.requireUpstream: true` (default).

**Fix**:
```bash
# Set upstream on the current branch
git push -u origin $(git branch --show-current)
```

Or disable the check in your config:
```yaml
git:
  requireUpstream: false
```

If `git.push: false`, the upstream check is skipped automatically (since 2026-02-18 fix).

---

### `GitHub release is enabled but GITHUB_TOKEN is not set`

**Cause**: `github.release: true` but the env var named in `github.tokenRef` (default: `GITHUB_TOKEN`) is empty.

**Fix**:
```bash
export GITHUB_TOKEN=ghp_your_personal_access_token
release-it-go
```

The token needs `repo` scope (or fine-grained equivalent: `Contents: read & write`). Generate at: https://github.com/settings/tokens

For GitLab: `GITLAB_TOKEN` with `api` scope.

---

### `no commits since latest tag. Nothing to release.`

**Cause**: There are no commits between the latest tag and `HEAD`. This is informational, not an error.

**Fix**: Make a commit, then re-run. Or use `--no-increment` to re-release the current version (rarely needed).

---

### `commit message is not conventional`

**Cause**: A commit since the last tag doesn't follow Conventional Commits format (`type(scope): description`).

**Fix**:
```bash
# See which commits failed
release-it-go --check-commits -V

# Rewrite the commit message
git commit --amend -m "fix: correct the previous message"

# Or rebase to fix older commits
git rebase -i HEAD~3
```

To bypass (not recommended):
```bash
release-it-go --ignore-commit-lint
```

Or set in config:
```yaml
git:
  requireConventionalCommits: false
```

---

### `unsupported remote URL format: ...`

**Cause**: Your `origin` remote URL doesn't match the supported HTTPS or SSH patterns.

**Fix**: Check your remote URL:
```bash
git remote get-url origin
```

Supported formats:
- `https://github.com/owner/repo.git`
- `https://user:token@github.com/owner/repo.git` (credentials are stripped)
- `git@github.com:owner/repo.git`
- `ssh://git@github.com:22/owner/repo.git`
- Nested groups: `https://gitlab.com/group/subgroup/repo.git`

If your URL is exotic (custom port, unusual host), open an issue with the URL pattern (with credentials redacted).

---

### Wrong version detected (CalVer / SemVer confusion)

**Cause**: `calver.enabled: true` and `preReleaseId` are set together (mutually exclusive).

**Fix**: Pick one. Either remove `preReleaseId` from config and CLI, or set `calver.enabled: false`.

---

### Pre-release version not incrementing on the same branch

**Symptom**: Running `--preRelease=beta` twice produces the same version (`v1.2.0-beta.0` → `v1.2.0-beta.0`).

**Cause**: Was a bug fixed in 2026-02-16 — make sure you're on the latest version.

**Verify**: `release-it-go version` should show v0.1.3 or newer.

---

### Pre-release picks up tags from another branch

**Cause**: Pre-2026-02-20 versions used global tag list. Phase 15 fixed this with branch-aware detection (`git tag --merged HEAD`).

**Fix**: Upgrade to v0.1.0+. Pre-release tag detection now only considers tags merged into the current `HEAD`.

---

### Webhook notification didn't arrive (Slack / Teams)

**Cause**: Webhook step is **non-fatal** by design — failures log a warning but don't stop the release.

**Debug**:
```bash
release-it-go -V    # verbose: see hook commands and webhook attempts
release-it-go -VV   # debug: see HTTP request details
```

Common causes:
- `urlRef` env var not set
- Webhook URL invalid or expired
- Network timeout (default 30s; raise via `notification.webhooks[].timeout`)
- Slack: incoming webhook disabled in workspace settings
- Teams: connector revoked or channel deleted

---

### Docker: `fatal: detected dubious ownership in repository`

**Cause**: When mounting a host repo into the container, git refuses to operate on a repo owned by a different UID.

**Fix**: The Dockerfile already runs `git config --global --add safe.directory '*'` for the `releaser` user. If you're using a custom image, add this line.

---

### Docker: `Author identity unknown`

**Cause**: Git needs a user identity to create commits. The Docker image checks for this and exits early with a clear message (Phase 12).

**Fix**:
```bash
docker run \
  -e GIT_AUTHOR_NAME="Your Name" \
  -e GIT_AUTHOR_EMAIL="you@example.com" \
  -e GIT_COMMITTER_NAME="Your Name" \
  -e GIT_COMMITTER_EMAIL="you@example.com" \
  -e GITHUB_TOKEN \
  -v $(pwd):/workspace \
  release-it-go:latest --ci
```

---

## Developer Issues

### `make check` fails with `command not found: golangci-lint`

**Fix**: Install golangci-lint (see [CONTRIBUTING.md](CONTRIBUTING.md#install-the-tooling)).

```bash
brew install golangci-lint
```

---

### `make check` fails with `command not found: govulncheck`

**Fix**:
```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Make sure `$(go env GOPATH)/bin` is in your `PATH`.

---

### `go test ./... -race` hangs or takes forever

**Cause**: An integration test is waiting for a real git command on a slow filesystem (often macOS with antivirus).

**Fix**:
```bash
# Run unit tests only (fast)
make test-unit

# Run a specific test
go test ./internal/runner/ -run TestRunner_Run -race -v
```

---

### Tests pass locally but fail in CI

Common causes:
1. **Case sensitivity**: macOS is case-insensitive, Linux is not. `CLAUDE.MD` vs `CLAUDE.md` etc.
2. **Locale**: Some tests assume `LC_ALL=C` or specific timezone. Run `LC_ALL=C TZ=UTC go test ./...` locally.
3. **Git config**: CI runs without your global git config (signing key, default branch, etc.). Make sure tests don't rely on global config.
4. **Goroutine leaks under `-race`**: CI runs `-race`, your local test might not.

---

### `commandExecutor` mock isn't being used

**Cause**: Mocking the wrong package's `commandExecutor`. There are **two** in this codebase:
- `internal/git/git.go` — for git operations
- `internal/githook/githook.go` — for git hook installer

Make sure you're mocking the one in the package whose code is under test.

---

### Coverage dropped after my change

**Fix**: Run `make coverage` and open `coverage.html`. Look for red lines in the package you changed. Add tests until you're back above the threshold.

Critical packages should stay at **85%+**: `git`, `runner`, `release`, `version`. Other internal packages: **70%+** minimum.

---

### Release / build workflow failed in CI

Common causes:
1. **`fetch-depth: 0` missing**: GoReleaser and release-it-go need full git history. The workflows already set this — check if you accidentally changed it.
2. **`GITHUB_TOKEN` permissions**: For self-hosted release workflow, the token needs `contents: write`.
3. **Tag conflict**: Two PRs merged with the same target version. Bump and re-tag.

---

### `npm release-it` config doesn't migrate cleanly

**Cause**: Some npm release-it plugin configs don't have a 1:1 mapping. The `internal/config/compat.go` handles common cases (e.g., `plugins.bumper`).

**Fix**:
```bash
release-it-go init   # if .release-it.json exists, this offers migration
```

If migration produces an unexpected config, file an issue with both the original `.release-it.json` and the migrated output. Don't manually edit `compat.go` without reading the existing patterns first.

---

### `golangci-lint` flags code I copied from elsewhere in the project

**Cause**: The lint rules are stricter than the project's older code in some places. Some legacy patterns are grandfathered.

**Fix**: Follow the lint suggestions for **new** code. Don't refactor old code in unrelated PRs to fix lint warnings — that creates churn. If the warning is wrong for the project's style, discuss in the PR.

---

## Still Stuck?

1. Search `PROGRESS.md` "Bugs" section — your problem may already be fixed
2. Check `git log --grep="<keyword>"` — the bug may have been documented in a commit message
3. Read the relevant `docs/phase_N.md` — the PRD often explains edge cases
4. Open an issue with:
   - Command you ran
   - Expected behavior
   - Actual output (with `-VV` debug logs if possible)
   - Your config (with secrets redacted)
   - Output of `release-it-go version`
