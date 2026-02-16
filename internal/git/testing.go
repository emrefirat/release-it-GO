package git

// SetCommandExecutorForTest replaces the command executor function for testing.
// Returns a function that restores the original executor.
// This should only be used in tests.
func SetCommandExecutorForTest(fn func(string, ...string) (string, error)) func() {
	original := commandExecutor
	commandExecutor = fn
	return func() {
		commandExecutor = original
	}
}
