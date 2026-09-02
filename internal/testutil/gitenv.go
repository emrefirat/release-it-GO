// Package testutil holds helpers shared by test packages that drive a real
// git binary (cli, githook, integration).
package testutil

import (
	"fmt"
	"os"
)

// IsolateGit makes git ignore the developer's global and system configuration
// for the current process, so settings such as commit.gpgsign, core.hooksPath,
// or init.defaultBranch cannot leak into tests. The default branch of new
// repositories is pinned to "main" and terminal prompts are disabled.
//
// Call it from TestMain before m.Run(): the git wrapper and the hook installer
// spawn git through exec.Command, which inherits the process environment.
func IsolateGit() error {
	vars := map[string]string{
		"GIT_CONFIG_GLOBAL":   os.DevNull,
		"GIT_CONFIG_NOSYSTEM": "1",
		"GIT_TERMINAL_PROMPT": "0",
		// Environment-level config (git ≥ 2.31): deterministic branch name
		// without touching any config file.
		"GIT_CONFIG_COUNT":   "1",
		"GIT_CONFIG_KEY_0":   "init.defaultBranch",
		"GIT_CONFIG_VALUE_0": "main",
	}
	for key, value := range vars {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("setting %s: %w", key, err)
		}
	}
	return nil
}
