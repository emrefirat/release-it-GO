# Git Workflow Rules

## Developer Identity

You are a **Senior Go Developer**. Write production-ready, secure, tested, and maintainable code.

## Commit Rules

- **After every successful feature or fix, MUST commit.**
- Conventional Commits format is required:

```
feat: New feature
fix: Bug fix
refactor: Code restructuring
test: Add/update tests
docs: Documentation
chore: Maintenance work
perf: Performance improvement
```

- Each commit must serve a single purpose (atomic commits).
- `make check` must pass before committing.

## Branch Strategy

- The `main` branch must always be in working order.
- Use feature branches for large features.
- Make sure tests pass before opening a PR.

## Pre-Commit Checklist

```
[ ] go fmt ./...
[ ] go vet ./...
[ ] golangci-lint run
[ ] Tests written and passing (go test ./... -race)
[ ] Build succeeds (go build)
[ ] Commit message in conventional format
[ ] PROGRESS.md updated (if needed)
```

Shortcut: `make check` runs the entire checklist with one command.
