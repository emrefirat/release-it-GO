// release-it-go is a release automation tool for Git projects.
// It handles Git tagging, changelog generation, and GitHub/GitLab releases.
package main

import "release-it-go/internal/cli"

// Build information, injected via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.Version = version
	cli.Commit = commit
	cli.Date = date

	cli.Execute()
}
