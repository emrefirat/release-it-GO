package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/emrefirat/release-it-GO/internal/testutil"
)

// TestMain isolates the tests that spawn a real git binary from the
// developer's global/system git configuration (see testutil.IsolateGit).
func TestMain(m *testing.M) {
	if err := testutil.IsolateGit(); err != nil {
		fmt.Fprintln(os.Stderr, "isolating git:", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
