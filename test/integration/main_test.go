package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/emrefirat/release-it-GO/internal/testutil"
)

// TestMain isolates every integration test from the developer's git
// configuration (commit.gpgsign, core.hooksPath, init.defaultBranch, ...).
// The runner spawns the real git binary, so this has to be process-wide.
func TestMain(m *testing.M) {
	if err := testutil.IsolateGit(); err != nil {
		fmt.Fprintln(os.Stderr, "isolating git:", err)
		os.Exit(1)
	}
	code := m.Run()
	removeBuiltBinary()
	os.Exit(code)
}
