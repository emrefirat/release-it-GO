//go:build windows

package hook

import (
	"os/exec"
	"syscall"
)

// setVerbatimCmdLine passes the command line to CreateProcess as-is instead
// of Go's per-argument quoting, which cmd.exe does not understand.
func setVerbatimCmdLine(c *exec.Cmd, cmdLine string) {
	if c.SysProcAttr == nil {
		c.SysProcAttr = &syscall.SysProcAttr{}
	}
	c.SysProcAttr.CmdLine = cmdLine
}
