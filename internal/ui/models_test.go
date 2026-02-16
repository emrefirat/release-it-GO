package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- selectModel Tests ---

func TestSelectModel_Init(t *testing.T) {
	m := selectModel{
		question: "Select version:",
		options: []VersionOption{
			{Label: "patch (1.0.1)", Version: "1.0.1"},
		},
	}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestSelectModel_View(t *testing.T) {
	options := []VersionOption{
		{Label: "patch (1.0.1)", Version: "1.0.1", Recommended: false},
		{Label: "minor (1.1.0)", Version: "1.1.0", Recommended: true},
		{Label: "major (2.0.0)", Version: "2.0.0", Recommended: false},
	}

	m := selectModel{
		question: "Select increment (current: 1.0.0):",
		options:  options,
		cursor:   0,
	}

	view := m.View()

	if !strings.Contains(view, "Select increment (current: 1.0.0):") {
		t.Error("View should contain the question")
	}
	if !strings.Contains(view, "> ") {
		t.Error("View should contain cursor indicator")
	}
	if !strings.Contains(view, "[recommended]") {
		t.Error("View should show [recommended] label")
	}
	if !strings.Contains(view, "enter to select") {
		t.Error("View should contain help text")
	}
}

func TestSelectModel_View_CursorPosition(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
		{Label: "minor", Version: "1.1.0"},
	}

	m := selectModel{
		question: "Pick:",
		options:  options,
		cursor:   1,
	}

	view := m.View()
	// The cursor should be on the second item
	lines := strings.Split(view, "\n")
	foundCursorOnSecond := false
	for _, line := range lines {
		if strings.HasPrefix(line, "> ") && strings.Contains(line, "minor") {
			foundCursorOnSecond = true
		}
	}
	if !foundCursorOnSecond {
		t.Error("Cursor should be on the second option (minor)")
	}
}

func TestSelectModel_Update_MoveDown(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
		{Label: "minor", Version: "1.1.0"},
		{Label: "major", Version: "2.0.0"},
	}

	m := selectModel{question: "Pick:", options: options, cursor: 0}

	// Move down
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if cmd != nil {
		t.Error("down key should not produce a command")
	}
	updated := result.(selectModel)
	if updated.cursor != 1 {
		t.Errorf("cursor should be 1, got %d", updated.cursor)
	}

	// Move down again
	result, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated = result.(selectModel)
	if updated.cursor != 2 {
		t.Errorf("cursor should be 2, got %d", updated.cursor)
	}

	// Move down at boundary (should stay)
	result, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated = result.(selectModel)
	if updated.cursor != 2 {
		t.Errorf("cursor should stay at 2, got %d", updated.cursor)
	}
}

func TestSelectModel_Update_MoveUp(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
		{Label: "minor", Version: "1.1.0"},
	}

	m := selectModel{question: "Pick:", options: options, cursor: 1}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if cmd != nil {
		t.Error("up key should not produce a command")
	}
	updated := result.(selectModel)
	if updated.cursor != 0 {
		t.Errorf("cursor should be 0, got %d", updated.cursor)
	}

	// Move up at boundary (should stay at 0)
	result, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated = result.(selectModel)
	if updated.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", updated.cursor)
	}
}

func TestSelectModel_Update_MoveWithJK(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
		{Label: "minor", Version: "1.1.0"},
	}

	m := selectModel{question: "Pick:", options: options, cursor: 0}

	// j moves down
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := result.(selectModel)
	if updated.cursor != 1 {
		t.Errorf("'j' should move cursor down to 1, got %d", updated.cursor)
	}

	// k moves up
	result, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated = result.(selectModel)
	if updated.cursor != 0 {
		t.Errorf("'k' should move cursor up to 0, got %d", updated.cursor)
	}
}

func TestSelectModel_Update_Enter(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
		{Label: "minor", Version: "1.1.0"},
	}

	m := selectModel{question: "Pick:", options: options, cursor: 1}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(selectModel)

	if cmd == nil {
		t.Error("enter should produce tea.Quit command")
	}
	if updated.selected != "1.1.0" {
		t.Errorf("selected should be '1.1.0', got %q", updated.selected)
	}
	if updated.cancelled {
		t.Error("should not be cancelled on enter")
	}
}

func TestSelectModel_Update_CtrlC(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
	}

	m := selectModel{question: "Pick:", options: options, cursor: 0}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	updated := result.(selectModel)

	if cmd == nil {
		t.Error("ctrl+c should produce tea.Quit command")
	}
	if !updated.cancelled {
		t.Error("should be cancelled on ctrl+c")
	}
}

func TestSelectModel_Update_QuitWithQ(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
	}

	m := selectModel{question: "Pick:", options: options, cursor: 0}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	updated := result.(selectModel)

	if cmd == nil {
		t.Error("'q' should produce tea.Quit command")
	}
	if !updated.cancelled {
		t.Error("should be cancelled on 'q'")
	}
}

func TestSelectModel_Update_UnhandledMsg(t *testing.T) {
	options := []VersionOption{
		{Label: "patch", Version: "1.0.1"},
	}

	m := selectModel{question: "Pick:", options: options, cursor: 0}

	// Send a non-key message
	result, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated := result.(selectModel)

	if cmd != nil {
		t.Error("unhandled message should not produce a command")
	}
	if updated.cursor != 0 {
		t.Error("cursor should remain unchanged")
	}
}

// --- confirmModel Tests ---

func TestConfirmModel_Init(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestConfirmModel_View(t *testing.T) {
	tests := []struct {
		name       string
		defaultYes bool
		wantHint   string
	}{
		{"default yes", true, "(Y/n)"},
		{"default no", false, "(y/N)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := confirmModel{question: "Continue?", defaultYes: tt.defaultYes}
			view := m.View()

			if !strings.Contains(view, "Continue?") {
				t.Error("View should contain the question")
			}
			if !strings.Contains(view, tt.wantHint) {
				t.Errorf("View should contain hint %q, got %q", tt.wantHint, view)
			}
		})
	}
}

func TestConfirmModel_Update_YKey(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: false}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	updated := result.(confirmModel)

	if cmd == nil {
		t.Error("'y' should produce tea.Quit command")
	}
	if !updated.result {
		t.Error("result should be true after 'y'")
	}
	if !updated.done {
		t.Error("done should be true after 'y'")
	}
	if updated.cancelled {
		t.Error("should not be cancelled")
	}
}

func TestConfirmModel_Update_UpperY(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: false}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
	updated := result.(confirmModel)

	if cmd == nil {
		t.Error("'Y' should produce tea.Quit command")
	}
	if !updated.result {
		t.Error("result should be true after 'Y'")
	}
}

func TestConfirmModel_Update_NKey(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated := result.(confirmModel)

	if cmd == nil {
		t.Error("'n' should produce tea.Quit command")
	}
	if updated.result {
		t.Error("result should be false after 'n'")
	}
	if !updated.done {
		t.Error("done should be true after 'n'")
	}
}

func TestConfirmModel_Update_UpperN(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	updated := result.(confirmModel)

	if cmd == nil {
		t.Error("'N' should produce tea.Quit command")
	}
	if updated.result {
		t.Error("result should be false after 'N'")
	}
}

func TestConfirmModel_Update_Enter_DefaultYes(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(confirmModel)

	if cmd == nil {
		t.Error("enter should produce tea.Quit command")
	}
	if !updated.result {
		t.Error("result should be true (defaultYes=true)")
	}
	if !updated.done {
		t.Error("done should be true")
	}
}

func TestConfirmModel_Update_Enter_DefaultNo(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: false}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(confirmModel)

	if cmd == nil {
		t.Error("enter should produce tea.Quit command")
	}
	if updated.result {
		t.Error("result should be false (defaultYes=false)")
	}
}

func TestConfirmModel_Update_CtrlC(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	updated := result.(confirmModel)

	if cmd == nil {
		t.Error("ctrl+c should produce tea.Quit command")
	}
	if !updated.cancelled {
		t.Error("should be cancelled on ctrl+c")
	}
}

func TestConfirmModel_Update_UnhandledKey(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	updated := result.(confirmModel)

	if cmd != nil {
		t.Error("unhandled key should not produce a command")
	}
	if updated.done {
		t.Error("done should remain false for unhandled key")
	}
}

func TestConfirmModel_Update_NonKeyMsg(t *testing.T) {
	m := confirmModel{question: "Continue?", defaultYes: true}

	result, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated := result.(confirmModel)

	if cmd != nil {
		t.Error("non-key message should not produce a command")
	}
	if updated.done {
		t.Error("done should remain false")
	}
}

// --- inputModel Tests ---

func TestInputModel_Init(t *testing.T) {
	m := inputModel{question: "Enter tag:", defaultValue: "v1.0.0", value: "v1.0.0"}
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestInputModel_View(t *testing.T) {
	tests := []struct {
		name         string
		question     string
		defaultValue string
		value        string
		wantContains []string
	}{
		{
			"with default",
			"Enter tag:",
			"v1.0.0",
			"v1.0.0",
			[]string{"Enter tag:", "(v1.0.0)", "v1.0.0"},
		},
		{
			"no default",
			"Enter notes:",
			"",
			"hello",
			[]string{"Enter notes:", "hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := inputModel{
				question:     tt.question,
				defaultValue: tt.defaultValue,
				value:        tt.value,
			}
			view := m.View()
			for _, want := range tt.wantContains {
				if !strings.Contains(view, want) {
					t.Errorf("View should contain %q, got %q", want, view)
				}
			}
		})
	}
}

func TestInputModel_View_NoDefault(t *testing.T) {
	m := inputModel{question: "Enter notes:", defaultValue: "", value: ""}
	view := m.View()

	// Should not contain parenthesized default
	if strings.Contains(view, "()") {
		t.Error("View should not contain empty parentheses")
	}
}

func TestInputModel_Update_TypeCharacters(t *testing.T) {
	m := inputModel{question: "Enter:", value: ""}

	// Type 'h'
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if cmd != nil {
		t.Error("typing should not produce a command")
	}
	updated := result.(inputModel)
	if updated.value != "h" {
		t.Errorf("value should be 'h', got %q", updated.value)
	}

	// Type 'i'
	result, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	updated = result.(inputModel)
	if updated.value != "hi" {
		t.Errorf("value should be 'hi', got %q", updated.value)
	}
}

func TestInputModel_Update_Backspace(t *testing.T) {
	m := inputModel{question: "Enter:", value: "hello"}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if cmd != nil {
		t.Error("backspace should not produce a command")
	}
	updated := result.(inputModel)
	if updated.value != "hell" {
		t.Errorf("value should be 'hell', got %q", updated.value)
	}
}

func TestInputModel_Update_Backspace_EmptyValue(t *testing.T) {
	m := inputModel{question: "Enter:", value: ""}

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	updated := result.(inputModel)
	if updated.value != "" {
		t.Errorf("value should remain empty, got %q", updated.value)
	}
}

func TestInputModel_Update_Enter(t *testing.T) {
	m := inputModel{question: "Enter:", value: "my-tag"}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := result.(inputModel)

	if cmd == nil {
		t.Error("enter should produce tea.Quit command")
	}
	if !updated.done {
		t.Error("done should be true after enter")
	}
	if updated.value != "my-tag" {
		t.Errorf("value should be 'my-tag', got %q", updated.value)
	}
}

func TestInputModel_Update_CtrlC(t *testing.T) {
	m := inputModel{question: "Enter:", value: "something"}

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	updated := result.(inputModel)

	if cmd == nil {
		t.Error("ctrl+c should produce tea.Quit command")
	}
	if !updated.cancelled {
		t.Error("should be cancelled on ctrl+c")
	}
}

func TestInputModel_Update_NonKeyMsg(t *testing.T) {
	m := inputModel{question: "Enter:", value: "test"}

	result, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated := result.(inputModel)

	if cmd != nil {
		t.Error("non-key message should not produce a command")
	}
	if updated.value != "test" {
		t.Error("value should remain unchanged")
	}
	if updated.done {
		t.Error("done should remain false")
	}
}

func TestInputModel_Update_FullFlow(t *testing.T) {
	m := inputModel{question: "Enter tag:", defaultValue: "v1.0.0", value: "v1.0.0"}

	// Clear the default by pressing backspace multiple times
	var model tea.Model = m
	for i := 0; i < 6; i++ {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	}

	// Type new value
	for _, ch := range "v2.0.0" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Press enter
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	final := model.(inputModel)

	if cmd == nil {
		t.Error("enter should produce quit command")
	}
	if final.value != "v2.0.0" {
		t.Errorf("value should be 'v2.0.0', got %q", final.value)
	}
	if !final.done {
		t.Error("done should be true")
	}
}
