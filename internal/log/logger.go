// Package log provides a structured logger with verbose level support
// and dry-run message formatting for release-it-go.
package log

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// VerboseNormal is the default log level (Info/Warn/Error only).
const VerboseNormal = 0

// VerboseLevel shows verbose messages (hook commands, git commands).
const VerboseLevel = 1

// DebugLevel shows all internal details.
const DebugLevel = 2

// Logger wraps slog with verbose level filtering and dry-run support.
type Logger struct {
	slogger *slog.Logger
	verbose int
	dryRun  bool
	output  io.Writer
}

// NewLogger creates a new Logger with the given verbose level and dry-run mode.
func NewLogger(verbose int, dryRun bool) *Logger {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return &Logger{
		slogger: slog.New(handler),
		verbose: verbose,
		dryRun:  dryRun,
		output:  os.Stderr,
	}
}

// NewLoggerWithWriter creates a Logger that writes to the given writer.
func NewLoggerWithWriter(verbose int, dryRun bool, w io.Writer) *Logger {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return &Logger{
		slogger: slog.New(handler),
		verbose: verbose,
		dryRun:  dryRun,
		output:  w,
	}
}

// Info logs a normal message (always visible).
func (l *Logger) Info(msg string, args ...any) {
	l.slogger.Info(fmt.Sprintf(msg, args...))
}

// Print writes a user-friendly message directly to output (no slog format).
// Always visible. Use for banner, version info, skip messages, etc.
func (l *Logger) Print(msg string, args ...any) {
	formatted := fmt.Sprintf(msg, args...)
	_, _ = fmt.Fprintln(l.output, formatted)
}

// Verbose logs a message visible only with -V flag (verbose >= 1).
// Outputs in indented dim format: "    ↳ message"
func (l *Logger) Verbose(msg string, args ...any) {
	if l.verbose >= VerboseLevel {
		_, _ = fmt.Fprintf(l.output, "    ↳ %s\n", fmt.Sprintf(msg, args...))
	}
}

// Debug logs a message visible only with -VV flag (verbose >= 2).
func (l *Logger) Debug(msg string, args ...any) {
	if l.verbose >= DebugLevel {
		l.slogger.Debug(fmt.Sprintf(msg, args...))
	}
}

// DryRun logs a dry-run prefixed message (always visible when called).
func (l *Logger) DryRun(msg string, args ...any) {
	l.slogger.Info(fmt.Sprintf("[dry-run] "+msg, args...))
}

// Warn logs a warning message (always visible).
func (l *Logger) Warn(msg string, args ...any) {
	l.slogger.Warn(fmt.Sprintf(msg, args...))
}

// Error logs an error message (always visible).
func (l *Logger) Error(msg string, args ...any) {
	l.slogger.Error(fmt.Sprintf(msg, args...))
}

// IsDryRun returns whether the logger is in dry-run mode.
func (l *Logger) IsDryRun() bool {
	return l.dryRun
}

// GetVerbose returns the current verbose level.
func (l *Logger) GetVerbose() int {
	return l.verbose
}
