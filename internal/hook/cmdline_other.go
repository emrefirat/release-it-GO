//go:build !windows

package hook

import "os/exec"

// setVerbatimCmdLine is a no-op outside Windows: sh -c receives the command
// as a single argument without any re-quoting.
func setVerbatimCmdLine(*exec.Cmd, string) {}
