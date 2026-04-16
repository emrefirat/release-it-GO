# Code Quality Rules

## Documentation Lookup with Context7

When writing code or using a library/API, **use the Context7 MCP tool**:

1. First find the library ID with `resolve-library-id`
2. Then query current documentation with `query-docs`
3. Don't write code from guesses — always verify the current API

Use Context7 especially when:
- Adding a new dependency (semver, cobra, viper, bubbletea, lipgloss)
- Unsure about an existing dependency's API
- Using an unfamiliar Go stdlib package

## Naming Conventions

| Kind | Format | Example |
|------|--------|---------|
| Package | lowercase, single word | `config`, `git`, `release` |
| Exported | PascalCase | `CreateRelease`, `UserID` |
| Unexported | camelCase | `generateSlug`, `validateInput` |
| Constants | PascalCase or SCREAMING_SNAKE | `MaxRetries`, `DEFAULT_TIMEOUT` |
| Interfaces | -er suffix | `Reader`, `Prompter`, `ReleaseProvider` |
| Acronyms | Consistent case | `HTTPServer`, `userID` |

## Error Handling

```go
// WRONG — Don't ignore errors
result, _ := someFunction()

// RIGHT — Handle every error
result, err := someFunction()
if err != nil {
    return fmt.Errorf("someFunction failed: %w", err)
}
```

- Use `%w` for error wrapping.
- Log only where the error originates (top level: main or runner).
- Don't use `panic` (only for unrecoverable situations).

## Custom Error Types

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)
```

## Code Organization

- One responsibility per file.
- Functions shouldn't exceed 50 lines (when possible).
- Avoid deep nesting (3+ levels); use early return.
- Don't use magic numbers; define constants.

## Early Return Pattern

```go
func process(data *Data) error {
    if data == nil {
        return errors.New("nil data")
    }
    if !data.IsValid() {
        return errors.New("invalid data")
    }
    // actual logic
    return nil
}
```

## Things Not to Do

- Don't use `panic`
- Don't use global state
- Side effects in `init()` functions
- Circular dependencies
- God objects/functions
- Ignoring errors (`_`)

## Logging

- Use the `log/slog` stdlib.
- Prefer structured logging.
- Log levels: normal / verbose / debug (configurable).
- Don't log sensitive data (passwords, tokens, PII).

## Performance

- For large slices, specify capacity: `make([]Item, 0, len(data))`
- Avoid unnecessary allocations.
- Avoid goroutine leaks — use context.
