package log

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(0, false)
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	if logger.GetVerbose() != 0 {
		t.Errorf("verbose = %d, expected 0", logger.GetVerbose())
	}
	if logger.IsDryRun() {
		t.Error("dryRun should be false")
	}
}

func TestLogger_Info_AlwaysVisible(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Info("test message")

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Info should always be visible, got: %q", buf.String())
	}
}

func TestLogger_Verbose_HiddenAtNormal(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Verbose("verbose message")

	if strings.Contains(buf.String(), "verbose message") {
		t.Error("Verbose should be hidden at normal level")
	}
}

func TestLogger_Verbose_VisibleAtVerboseLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseLevel, false, &buf)

	logger.Verbose("verbose message")

	output := buf.String()
	if !strings.Contains(output, "verbose message") {
		t.Errorf("Verbose should be visible at verbose level, got: %q", output)
	}
	if !strings.Contains(output, "↳") {
		t.Errorf("Verbose should contain ↳ prefix, got: %q", output)
	}
}

func TestLogger_Debug_HiddenAtNormal(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Debug("debug message")

	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug should be hidden at normal level")
	}
}

func TestLogger_Debug_HiddenAtVerbose(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseLevel, false, &buf)

	logger.Debug("debug message")

	if strings.Contains(buf.String(), "debug message") {
		t.Error("Debug should be hidden at verbose level")
	}
}

func TestLogger_Debug_VisibleAtDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(DebugLevel, false, &buf)

	logger.Debug("debug message")

	if !strings.Contains(buf.String(), "debug message") {
		t.Errorf("Debug should be visible at debug level, got: %q", buf.String())
	}
}

func TestLogger_DryRun_Prefix(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, true, &buf)

	logger.DryRun("would create tag %s", "v1.0.0")

	output := buf.String()
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("DryRun should contain [dry-run] prefix, got: %q", output)
	}
	if !strings.Contains(output, "would create tag v1.0.0") {
		t.Errorf("DryRun should contain formatted message, got: %q", output)
	}
}

func TestLogger_Warn_AlwaysVisible(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Warn("warning message")

	if !strings.Contains(buf.String(), "warning message") {
		t.Errorf("Warn should always be visible, got: %q", buf.String())
	}
}

func TestLogger_Error_AlwaysVisible(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Error("error message")

	if !strings.Contains(buf.String(), "error message") {
		t.Errorf("Error should always be visible, got: %q", buf.String())
	}
}

func TestLogger_IsDryRun(t *testing.T) {
	logger := NewLogger(0, true)
	if !logger.IsDryRun() {
		t.Error("IsDryRun should return true")
	}

	logger2 := NewLogger(0, false)
	if logger2.IsDryRun() {
		t.Error("IsDryRun should return false")
	}
}

func TestLogger_GetVerbose(t *testing.T) {
	tests := []struct {
		level int
	}{
		{VerboseNormal},
		{VerboseLevel},
		{DebugLevel},
	}

	for _, tt := range tests {
		logger := NewLogger(tt.level, false)
		if logger.GetVerbose() != tt.level {
			t.Errorf("GetVerbose() = %d, want %d", logger.GetVerbose(), tt.level)
		}
	}
}

func TestLogger_InfoFormatting(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Info("version %s released on %s", "1.0.0", "2026-02-16")

	output := buf.String()
	if !strings.Contains(output, "version 1.0.0 released on 2026-02-16") {
		t.Errorf("Info should support fmt.Sprintf formatting, got: %q", output)
	}
}

func TestLogger_Print_AlwaysVisible(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Print("  📦 Version: %s → %s", "1.0.0", "1.1.0")

	output := buf.String()
	if !strings.Contains(output, "📦 Version: 1.0.0 → 1.1.0") {
		t.Errorf("Print should output formatted message directly, got: %q", output)
	}
	// Print should NOT contain slog format elements
	if strings.Contains(output, "level=") || strings.Contains(output, "time=") {
		t.Errorf("Print should not use slog format, got: %q", output)
	}
}

func TestLogger_Print_Formatting(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseNormal, false, &buf)

	logger.Print("hello %s", "world")

	output := buf.String()
	if output != "hello world\n" {
		t.Errorf("Print should output 'hello world\\n', got: %q", output)
	}
}

func TestLogger_Verbose_DimFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLoggerWithWriter(VerboseLevel, false, &buf)

	logger.Verbose("repo: %s/%s", "owner", "repo")

	output := buf.String()
	expected := "    ↳ repo: owner/repo\n"
	if output != expected {
		t.Errorf("Verbose should output %q, got: %q", expected, output)
	}
}
