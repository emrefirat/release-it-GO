# Changelog

## Unreleased

### Bug Fixes

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
