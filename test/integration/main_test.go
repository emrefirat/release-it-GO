package integration

import (
	"fmt"
	"os"
	"testing"

	"release-it-go/internal/testutil"
)

// TestMain isolates every integration test from the developer's git
// configuration (commit.gpgsign, core.hooksPath, init.defaultBranch, ...).
// The runner spawns the real git binary, so this has to be process-wide.
func TestMain(m *testing.M) {
	if err := testutil.IsolateGit(); err != nil {
		fmt.Fprintln(os.Stderr, "isolating git:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
