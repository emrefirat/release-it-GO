# Changelog

## [0.1.0](https://github.com/emrefirat/release-it-GO/compare/0.0.0...0.1.0) (2026-03-23)

### Features

* add --preRelease shorthand flag for sub-branch versioning
* add CHANGELOG.md file toggle to init wizard
* add Docker container support (Phase 11)
* add Docker pre-flight checks for git identity and API tokens
* add YAML config writing, init format selection, and explicit wizard fields
* add branch-aware pre-release version detection
* add conventional commit linting (Phase 9)
* add init --full-example to generate comprehensive config reference
* add init command with dual config support and legacy migration
* add npm release-it config backward compatibility
* add pipeline
* add webhook notification support for Slack and Teams (Phase 13)
* enable requireConventionalCommits by default
* implement Phase 1 - Core Foundation
* implement Phase 2 - Git Operations
* implement Phase 3 - Conventional Commits + Changelog
* implement Phase 4 - GitHub + GitLab Releases
* implement Phase 5 - Interactive UI + Hooks + Pipeline
* implement Phase 6 - Advanced Features
* improve UI/Output for user-friendly messages (Phase 10)
* show warning when no config file found, log loaded config path
* validate API tokens in prerequisites instead of release step

### Bug Fixes

* add debug logging to GitLab auth header for CI troubleshooting
* auto-detect CI mode when no TTY is available
* auto-detect CI_JOB_TOKEN and use Job-Token header for GitLab CI
* disable requireUpstream and requireCleanWorkingDir when push is off
* exclude untracked files from clean working directory check
* handle first release changelog when no tags exist
* keep requireCleanWorkingDir enabled when push is off
* pre-release version not incrementing with same ID
* remove requireBranch question from init wizard
* remove requireCommits question from init wizard
* require GIT_USER_NAME and GIT_USER_EMAIL in Docker entrypoint
* require commits before release by default
* resolve all golangci-lint errors across codebase
* resolve golangci-lint errors across codebase
* skip git identity check for info-only docker commands
* skip upstream check when push is disabled
* stage CHANGELOG.md before release commit
* strip credentials from HTTPS URLs in ParseRepoURL
* use /projects/:id for GitLab token validation instead of /user
* use curated JSON for init --full-example instead of struct marshaling
* use env variables for git identity in Docker instead of hardcoded defaults
* use tagName template in latestVersionToTag instead of hardcoded v prefix
* validate commit types against Angular preset in lint

### Reverts

* remove GitLab CI_JOB_TOKEN changes (4 commits)
