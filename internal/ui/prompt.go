package ui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Prompter defines the interface for user interaction.
type Prompter interface {
	// SelectVersion asks the user to select a version increment.
	SelectVersion(current string, recommended string, options []VersionOption) (string, error)

	// Confirm asks a yes/no question.
	Confirm(message string, defaultYes bool) (bool, error)

	// Input asks for free-text input.
	Input(message string, defaultValue string) (string, error)
}

// VersionOption represents a version choice in the selection prompt.
type VersionOption struct {
	Label       string // "patch (1.2.4)"
	Version     string // "1.2.4"
	Recommended bool
}

// --- NonInteractivePrompter ---

// NonInteractivePrompter automatically answers all prompts for CI mode.
type NonInteractivePrompter struct{}

// SelectVersion returns the recommended version, or the first option if none is recommended.
func (p *NonInteractivePrompter) SelectVersion(current string, recommended string, options []VersionOption) (string, error) {
	if recommended != "" {
		return recommended, nil
	}
	for _, opt := range options {
		if opt.Recommended {
			return opt.Version, nil
		}
	}
	if len(options) > 0 {
		return options[0].Version, nil
	}
	return current, nil
}

// Confirm always returns the default value.
func (p *NonInteractivePrompter) Confirm(message string, defaultYes bool) (bool, error) {
	return defaultYes, nil
}

// Input always returns the default value.
func (p *NonInteractivePrompter) Input(message string, defaultValue string) (string, error) {
	return defaultValue, nil
}

// --- InteractivePrompter ---

// InteractivePrompter uses bubbletea for terminal UI prompts.
type InteractivePrompter struct{}

// SelectVersion presents a list of version options to the user.
func (p *InteractivePrompter) SelectVersion(current string, recommended string, options []VersionOption) (string, error) {
	m := selectModel{
		question: fmt.Sprintf("Select increment (current: %s):", current),
		options:  options,
	}

	// Pre-select recommended option
	for i, opt := range options {
		if opt.Recommended {
			m.cursor = i
			break
		}
	}

	prog := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := prog.Run()
	if err != nil {
		return "", fmt.Errorf("version selection prompt: %w", err)
	}

	final := result.(selectModel)
	if final.cancelled {
		return "", fmt.Errorf("version selection cancelled")
	}

	return final.selected, nil
}

// Confirm asks the user a yes/no question.
func (p *InteractivePrompter) Confirm(message string, defaultYes bool) (bool, error) {
	m := confirmModel{
		question:   message,
		defaultYes: defaultYes,
		result:     defaultYes,
	}

	prog := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := prog.Run()
	if err != nil {
		return defaultYes, fmt.Errorf("confirm prompt: %w", err)
	}

	final := result.(confirmModel)
	if final.cancelled {
		return false, fmt.Errorf("confirmation cancelled")
	}

	return final.result, nil
}

// Input asks the user for free-text input.
func (p *InteractivePrompter) Input(message string, defaultValue string) (string, error) {
	m := inputModel{
		question:     message,
		defaultValue: defaultValue,
		value:        defaultValue,
	}

	prog := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	result, err := prog.Run()
	if err != nil {
		return defaultValue, fmt.Errorf("input prompt: %w", err)
	}

	final := result.(inputModel)
	if final.cancelled {
		return "", fmt.Errorf("input cancelled")
	}

	if final.value == "" {
		return defaultValue, nil
	}
	return final.value, nil
}

// --- Select Model (bubbletea) ---

type selectModel struct {
	question  string
	options   []VersionOption
	cursor    int
	selected  string
	cancelled bool
}

func (m selectModel) Init() tea.Cmd { return nil }

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.options[m.cursor].Version
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectModel) View() string {
	s := fmt.Sprintf("? %s\n\n", m.question)

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	for i, opt := range m.options {
		cursor := "  "
		label := opt.Label
		if i == m.cursor {
			cursor = "> "
			label = selectedStyle.Render(opt.Label)
		}
		if opt.Recommended {
			label += " [recommended]"
		}
		s += fmt.Sprintf("%s%s\n", cursor, label)
	}

	s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("↑/↓ to move, enter to select, q to cancel")

	return s
}

// --- Confirm Model (bubbletea) ---

type confirmModel struct {
	question   string
	defaultYes bool
	result     bool
	done       bool
	cancelled  bool
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "y", "Y":
			m.result = true
			m.done = true
			return m, tea.Quit
		case "n", "N":
			m.result = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.result = m.defaultYes
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	hint := "(y/N)"
	if m.defaultYes {
		hint = "(Y/n)"
	}
	return fmt.Sprintf("? %s %s ", m.question, hint)
}

// --- Input Model (bubbletea) ---

type inputModel struct {
	question     string
	defaultValue string
	value        string
	done         bool
	cancelled    bool
}

func (m inputModel) Init() tea.Cmd { return nil }

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "backspace":
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
		default:
			if len(msg.String()) == 1 && !strings.HasPrefix(msg.String(), "ctrl+") {
				m.value += msg.String()
			}
		}
	}
	return m, nil
}

func (m inputModel) View() string {
	prompt := fmt.Sprintf("? %s", m.question)
	if m.defaultValue != "" {
		prompt += fmt.Sprintf(" (%s)", m.defaultValue)
	}
	return fmt.Sprintf("%s: %s", prompt, m.value)
}
