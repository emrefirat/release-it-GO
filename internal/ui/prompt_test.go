package ui

import "testing"

func TestNonInteractivePrompter_SelectVersion_Recommended(t *testing.T) {
	p := &NonInteractivePrompter{}
	options := []VersionOption{
		{Label: "patch (1.0.1)", Version: "1.0.1", Recommended: false},
		{Label: "minor (1.1.0)", Version: "1.1.0", Recommended: true},
		{Label: "major (2.0.0)", Version: "2.0.0", Recommended: false},
	}

	selected, err := p.SelectVersion("1.0.0", "1.1.0", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "1.1.0" {
		t.Errorf("expected 1.1.0, got %s", selected)
	}
}

func TestNonInteractivePrompter_SelectVersion_NoRecommended(t *testing.T) {
	p := &NonInteractivePrompter{}
	options := []VersionOption{
		{Label: "patch (1.0.1)", Version: "1.0.1", Recommended: false},
		{Label: "minor (1.1.0)", Version: "1.1.0", Recommended: false},
	}

	selected, err := p.SelectVersion("1.0.0", "", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "1.0.1" {
		t.Errorf("expected 1.0.1 (first option), got %s", selected)
	}
}

func TestNonInteractivePrompter_SelectVersion_EmptyOptions(t *testing.T) {
	p := &NonInteractivePrompter{}
	selected, err := p.SelectVersion("1.0.0", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "1.0.0" {
		t.Errorf("expected 1.0.0 (current), got %s", selected)
	}
}

func TestNonInteractivePrompter_SelectVersion_WithRecommendedParam(t *testing.T) {
	p := &NonInteractivePrompter{}
	options := []VersionOption{
		{Label: "patch (1.0.1)", Version: "1.0.1", Recommended: false},
	}

	selected, err := p.SelectVersion("1.0.0", "2.0.0", options)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "2.0.0" {
		t.Errorf("expected 2.0.0 (recommended param), got %s", selected)
	}
}

func TestNonInteractivePrompter_Confirm(t *testing.T) {
	p := &NonInteractivePrompter{}

	tests := []struct {
		name       string
		defaultYes bool
		expected   bool
	}{
		{"default yes", true, true},
		{"default no", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Confirm("Continue?", tt.defaultYes)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNonInteractivePrompter_Input(t *testing.T) {
	p := &NonInteractivePrompter{}

	result, err := p.Input("Enter version:", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", result)
	}
}

func TestNonInteractivePrompter_Select(t *testing.T) {
	p := &NonInteractivePrompter{}

	tests := []struct {
		name         string
		options      []string
		defaultIndex int
		expected     int
	}{
		{"default index 0", []string{"a", "b", "c"}, 0, 0},
		{"default index 1", []string{"a", "b", "c"}, 1, 1},
		{"default index 2", []string{"a", "b", "c"}, 2, 2},
		{"out of bounds negative", []string{"a", "b"}, -1, 0},
		{"out of bounds positive", []string{"a", "b"}, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx, err := p.Select("Choose:", tt.options, tt.defaultIndex)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if idx != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, idx)
			}
		})
	}
}

func TestGenericSelectModel_Init(t *testing.T) {
	m := genericSelectModel{
		question: "Pick one:",
		options:  []string{"A", "B", "C"},
		cursor:   0,
	}

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestGenericSelectModel_View(t *testing.T) {
	m := genericSelectModel{
		question: "Pick one:",
		options:  []string{"A", "B", "C"},
		cursor:   1,
	}

	view := m.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}
