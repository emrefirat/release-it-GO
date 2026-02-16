package ui

import (
	"os"
	"testing"
)

func TestColorsEnabled_Default(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	if !ColorsEnabled() {
		t.Error("expected ColorsEnabled() to return true when NO_COLOR is not set")
	}
}

func TestColorsEnabled_NoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	if ColorsEnabled() {
		t.Error("expected ColorsEnabled() to return false when NO_COLOR is set")
	}
}

func TestFormatFunctions_WithColors(t *testing.T) {
	os.Unsetenv("NO_COLOR")

	tests := []struct {
		name string
		fn   func(string) string
	}{
		{"FormatSuccess", FormatSuccess},
		{"FormatError", FormatError},
		{"FormatWarning", FormatWarning},
		{"FormatInfo", FormatInfo},
		{"FormatDim", FormatDim},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("test")
			if result == "" {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestFormatFunctions_NoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	tests := []struct {
		name     string
		fn       func(string) string
		expected string
	}{
		{"FormatSuccess", FormatSuccess, "test"},
		{"FormatError", FormatError, "test"},
		{"FormatWarning", FormatWarning, "test"},
		{"FormatInfo", FormatInfo, "test"},
		{"FormatDim", FormatDim, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn("test")
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatDryRun_WithColors(t *testing.T) {
	os.Unsetenv("NO_COLOR")
	result := FormatDryRun("test message")
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestFormatDryRun_NoColor(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	result := FormatDryRun("test message")
	expected := "[dry-run] test message"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
