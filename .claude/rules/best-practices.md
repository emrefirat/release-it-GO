# Go Best Practices and Project-Specific Rules

## Dependency Management

- Before adding a new dependency, evaluate whether stdlib can solve it.
- Run `go mod tidy` after every dependency change.
- Verify checksums with `go mod verify`.
- Minimize the number of indirect dependencies.

## Interface Design

- Define interfaces on the consumer side, not the producer side.
- Prefer small interfaces (1-3 methods). Large interfaces should be split.
- Use stdlib interfaces like `io.Reader`, `io.Writer` when possible.

```go
// RIGHT — Small, focused interface
type Prompter interface {
    Confirm(msg string, def bool) (bool, error)
    SelectVersion(current, recommended string, options []VersionOption) (string, error)
}

// WRONG — God interface
type Everything interface {
    // 20+ methods...
}
```

## Context and Cancellation

- Accept `context.Context` for long-running operations.
- When starting a goroutine, always set up a cancellation mechanism via context.
- Set timeouts on HTTP clients.

## Concurrency

- Prefer channels over mutexes (when possible).
- Use the `defer cancel()` pattern to prevent goroutine leaks.
- Manage goroutine lifecycles with `sync.WaitGroup`.
- Check for race conditions with `go test -race ./...`.

## API and HTTP Client Rules

- Timeouts on HTTP clients are required (default: 30s).
- Use exponential backoff for retry logic.
- Always close the response body with `defer resp.Body.Close()`.
- Skipping TLS certificate verification (InsecureSkipVerify) is only allowed in test environments.

## Config Management

- Default values are centralized in `config/defaults.go`.
- Config structs in `config/config.go`.
- When adding a new config field: struct + default + JSON/YAML tag + test.
- Config file formats: JSON, YAML, TOML (auto-detected by Viper).

## CLI Flag Rules

- Define new flags in `internal/cli/root.go`.
- Flag names use kebab-case: `--pre-release`, `--dry-run`.
- For boolean flags, use `--no-` prefix to disable: `--no-git.push`.
- Every flag requires a clear `Usage` string.

## Adding a Pipeline Step

To add a new pipeline step:
1. Add to the `pipelineStep` struct in `runner.go`
2. Implement the step function (spinner + error handling)
3. Before/after hook support comes for free
4. Dry-run support is required
5. Add an integration test

## Changelog and Commit Analysis

- Commit parsing: `changelog/parser.go` (Angular preset)
- Bump analysis: `changelog/analyzer.go` (feat→minor, fix→patch, !→major)
- When adding a new commit type, update the `allowedTypes` map

## Docker Best Practices

- Use multi-stage builds (builder + runtime).
- Only the necessary deps in the runtime image (git, ca-certificates).
- Always build the binary statically: `CGO_ENABLED=0`.
- Run as non-root user.
- Keep `.dockerignore` up to date.

## File Operations

- Use `filepath.Join` for file paths (platform-independent).
- Use `os.CreateTemp` or `t.TempDir()` (in tests) for temp files.
- File permissions: 0644 (file), 0755 (directory, executable).
- Use `defer f.Close()` after opening a file.

## Error Messages

- User-facing messages should be clear and actionable.
- Don't expose internal error details to the user.
- Show more detail in verbose mode (-v).
- Wrap errors with `%w` so the root cause is reachable.
