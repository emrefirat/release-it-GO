# Changelog

## Unreleased

### Features

* positional increment and explicit target versions, npm style: `release-it-go minor`, `release-it-go 1.5.0`, `-i 1.5.0` (previously positional args were silently ignored). Invalid arguments are rejected with a clear error.
* v-prefix tag inference: with the default `tagName`, a repo whose latest tag is v-prefixed keeps the prefix for new tags (and conventional-commit auto-increment now works there). Writing `tagName` in the config file disables the inference.
* template variables beyond `${version}` (`${branchName}`, `${latestVersion}`, `${repo.*}`, …) now render in `git.commitMessage`, `git.tagAnnotation`, and GitHub/GitLab `releaseName`.
* the interactive `--preRelease` menu now offers pre-release variants (including continuing the current series) instead of dropping the identifier.

### Bug Fixes

* `--check-msg` diagnostics redesigned for readability: what you wrote, what is wrong (with a `did you mean "fix"?` suggestion for typos and wrong-case types), the expected format, and an example built from your own message with the type corrected. `-V` lists every accepted type (`build` was missing) and only rules the linter actually enforces. `fixup!`, `squash!`, and `amend!` commits are accepted (so `git commit --fixup` works under the commit-msg hook), leading blank/comment lines in COMMIT_EDITMSG are skipped, the "No config file found" warning no longer fires in lint modes, and `--check-commits -V` no longer lists each failure twice.
* the bumper no longer destroys file formatting: version updates in JSON/YAML/TOML now replace only the version value (preserving key order, indentation, and comments), with the old full rewrite kept as a verified fallback.
* transient GitHub/GitLab/webhook failures (429/502/503/504) are retried with exponential backoff, honoring `Retry-After`; connection errors are replayed only for idempotent requests.
* lifecycle hooks now work on Windows (`%COMSPEC% /C` instead of requiring `sh`), and hook scripts receive `RELEASE_*` environment variables (`RELEASE_VERSION`, `RELEASE_TAG_NAME`, `RELEASE_REPO_OWNER`, ...).
* `tagMatch`/`tagExclude` support real glob patterns (`?`, character classes, mid-pattern `*`); the latest tag is picked by semver comparison, so a same-base pre-release no longer shadows the stable release.
* the release commit stages only what the release changed (bumper outputs + changelog); unrelated local edits are no longer swept in when `requireCleanWorkingDir` is disabled (`addUntrackedFiles: true` keeps the old sweep).
* git errors now carry both git's stderr and the root cause; a failed push prints the `--no-increment` recovery steps.
* an explicit increment (positional, `-i`, or config) no longer opens the interactive version prompt when it coincides with the auto-detected one.
* declining the Commit or Tag prompt no longer silently cancels the operations after it — commit, tag, and push are confirmed independently.
* `--no-increment` now completes when the release tag already exists at HEAD (npm's documented recovery flow after a failed push); a tag on a different commit remains a clear error.
* config migration no longer silently drops the `hooks`, `bumper`, `calver`, and `notification` sections (the writer serialized only 4 of 8 sections).
* `before:release` and `after:release` lifecycle hooks now fire (before the git release step and at pipeline end, matching npm release-it). They were accepted in config and documented, but the pipeline never emitted them.
* `ci`, `dry-run`, and `verbose` set in the config file are no longer overwritten by unset CLI flags — flag values only override the config when actually passed.
* webhook failure messages no longer contain the webhook URL (a bearer credential for Slack/Teams); the error names the webhook type and `urlRef` instead.
* `requireBranch` now supports comma-separated patterns with any-of matching (`"main,master"`, `"main,release/*"`), restoring npm release-it's array semantics for migrated configs.
* `BREAKING CHANGE:` footers now trigger a major bump and populate the BREAKING CHANGES changelog section. The pipeline previously fetched only commit subject lines, so the spec-canonical footer form was silently analyzed as minor/patch and only `feat!:` worked. Commit hashes now render in changelog entries, multiline footer values are parsed, and conventional-commit auto-increment applies the same tag-format-transition fallback the changelog already had.
* default `github.host` (`api.github.com`) was resolved as a GitHub Enterprise host, so every GitHub API call targeted the nonexistent `https://api.github.com/api/v3` path and failed with 404. The default is now `github.com` and `api.github.com` is accepted as an alias for the public API.
* GitLab TLS verification is now **on by default**: the zero value of `gitlab.secure` used to disable certificate verification, sending the API token over unverified TLS in every default-config run. Self-signed instances should set `gitlab.certificateAuthorityFile`(`Ref`) or explicitly opt out with `secure: false`. An invalid CA file no longer installs an empty root pool (which broke all connections with an opaque x509 error); it warns and falls back to the system roots.
* `hooks install` now prunes managed git hooks that were removed from the config (and resets `core.hooksPath` when none remain), instead of leaving stale hooks active forever.
* default `git.pushArgs` now includes `--atomic` so branch and tag refs land as a single transaction. Prevents orphan tags on the remote when concurrent CI advances the branch ref and the push is rejected. Users on legacy git servers (pre-2015) without atomic protocol support can revert via `git.pushArgs: ["--follow-tags"]` config override.

## [0.1.3](https://github.com/emrefirat/release-it-GO/compare/0.1.2...0.1.3) (2026-03-30)

## [0.1.2](https://github.com/emrefirat/release-it-GO/compare/0.1.1...0.1.2) (2026-03-30)

### Features

* add ignoredContributors, themeColor, imageUrl to webhook config
* rich Teams MessageCard notifications with facts and changelog

### Bug Fixes

* GitLab nested group URLs and CalVer yyyy.mm.dd format (Phase 16)
* config compat and edge case improvements (Phase 18)
* pipeline robustness improvements (Phase 17)
* remove non-existent --no-github.release flag from workflow
* skip empty commit when no staged changes exist
* tag format-aware filtering to handle tagName config changes

## [0.1.1](https://github.com/emrefirat/release-it-GO/compare/0.1.0...0.1.1) (2026-03-23)

### Bug Fixes

* code review findings - security, error handling, dead code
