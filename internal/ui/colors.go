package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	// StyleSuccess renders text in green.
	StyleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	// StyleWarning renders text in yellow.
	StyleWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	// StyleError renders text in red.
	StyleError = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	// StyleInfo renders text in blue.
	StyleInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	// StyleDim renders text in gray.
	StyleDim = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	// StyleBold renders text in bold.
	StyleBold = lipgloss.NewStyle().Bold(true)
)

// ColorsEnabled returns true if colored output is enabled.
// Respects the NO_COLOR environment variable.
func ColorsEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}

// FormatSuccess formats a message with green color.
func FormatSuccess(msg string) string {
	if !ColorsEnabled() {
		return msg
	}
	return StyleSuccess.Render(msg)
}

// FormatError formats a message with red color.
func FormatError(msg string) string {
	if !ColorsEnabled() {
		return msg
	}
	return StyleError.Render(msg)
}

// FormatWarning formats a message with yellow color.
func FormatWarning(msg string) string {
	if !ColorsEnabled() {
		return msg
	}
	return StyleWarning.Render(msg)
}

// FormatInfo formats a message with blue color.
func FormatInfo(msg string) string {
	if !ColorsEnabled() {
		return msg
	}
	return StyleInfo.Render(msg)
}

// FormatDim formats a message with gray color.
func FormatDim(msg string) string {
	if !ColorsEnabled() {
		return msg
	}
	return StyleDim.Render(msg)
}

// FormatDryRun formats a dry-run message with dim styling.
func FormatDryRun(msg string) string {
	if !ColorsEnabled() {
		return "[dry-run] " + msg
	}
	return StyleDim.Render("[dry-run] " + msg)
}
