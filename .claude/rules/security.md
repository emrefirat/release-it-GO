# Security Rules

## NEVER Do

- Store passwords/secrets in plain text or hardcode them in code
- Use user input without validation
- Log sensitive information (passwords, tokens, PII)
- Hardcoded credentials
- Use `panic` instead of error handling
- Expose internal details in error messages

## ALWAYS Do

- Apply input validation
- Use environment variables for tokens (`tokenRef` pattern)
- Keep dependencies up to date
- Apply credential stripping in HTTPS URLs

## Security Tools

`make check` (which includes govulncheck) must be run before committing. Additionally:

```bash
# Dependency vulnerability scan (`make vuln` in Makefile)
govulncheck ./...

# Static security analysis — run on suspicion or when adding a new feature
gosec ./...

# General static analysis (`make lint` in Makefile)
golangci-lint run
```

### When to Use gosec
- When adding new HTTP client, file operation, or `exec.Command` code
- When external input handling logic changes
- When a security concern arises during review
- As a final check before release

## Security Patterns Used in This Project

### Token Management
Tokens are never written to config. The env var name is specified via `tokenRef`:
```json
{
  "github": {
    "tokenRef": "GITHUB_TOKEN"
  }
}
```

### Webhook URL Security
Webhook URLs are read via env var using `urlRef`; they are not written directly to config.

### Command Execution Security
Git commands run via `exec.Command("git", args...)`. Never go through a shell (`sh -c`) — this prevents the command injection risk.

### Docker Security
- Non-root user (`releaser:1000`)
- Static binary (CGO_ENABLED=0)
- Minimal base image (alpine)
- Git identity env var requirement (for release operations)
