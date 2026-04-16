# Testing Rules

## Requirements

- **Unit tests are REQUIRED for every new function.**
- Minimum test coverage: 70% (target 80%).
- DO NOT commit if tests fail.
- Race detection: `go test -race ./...`

## Test Naming

```go
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {
    // Arrange — Prepare test data
    // Act — Call the function under test
    // Assert — Verify the results
}
```

## Table-Driven Tests (Preferred)

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("got = %v, expected %v", got, tt.expected)
            }
        })
    }
}
```

## Mock Pattern (Git Operations)

In this project, git commands run through a `commandExecutor` function variable. Mock in tests:

```go
original := commandExecutor
defer func() { commandExecutor = original }()
commandExecutor = func(name string, args ...string) (string, error) {
    return "mock output", nil
}
```

## Test Commands

```bash
make test               # All tests (-v -cover -race)
make test-unit          # internal/ tests only
make test-integration   # test/integration/ only
make coverage           # Generate HTML coverage report
```

## Layout

- Unit tests: each package's own `*_test.go` files
- Integration tests: `test/integration/` (creates a real git repo)
- Test fixtures: `test/integration/fixtures/` (config samples)
