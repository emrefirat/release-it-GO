# Changelog

## Unreleased

### Bug Fixes

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
