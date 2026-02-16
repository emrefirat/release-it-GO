package version

import (
	"fmt"
	"testing"
	"time"
)

func TestCalVer_NextVersion_FirstVersion(t *testing.T) {
	cv := NewCalVer("yy.mm.minor", "calendar", "minor")

	got, err := cv.NextVersion("")
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	now := time.Now()
	expected := fmt.Sprintf("%d.%d.0", now.Year()%100, int(now.Month()))
	if got != expected {
		t.Errorf("NextVersion(\"\") = %q, want %q", got, expected)
	}
}

func TestCalVer_NextVersion_SameMonth(t *testing.T) {
	cv := NewCalVer("yy.mm.minor", "calendar", "minor")
	now := time.Now()
	current := fmt.Sprintf("%d.%d.0", now.Year()%100, int(now.Month()))

	got, err := cv.NextVersion(current)
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	expected := fmt.Sprintf("%d.%d.1", now.Year()%100, int(now.Month()))
	if got != expected {
		t.Errorf("NextVersion(%q) = %q, want %q", current, got, expected)
	}
}

func TestCalVer_NextVersion_DifferentMonth(t *testing.T) {
	cv := NewCalVer("yy.mm.minor", "calendar", "minor")
	now := time.Now()

	// Use a previous month
	prevMonth := int(now.Month()) - 1
	if prevMonth <= 0 {
		prevMonth = 12
	}
	current := fmt.Sprintf("%d.%d.5", now.Year()%100, prevMonth)

	got, err := cv.NextVersion(current)
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	expected := fmt.Sprintf("%d.%d.0", now.Year()%100, int(now.Month()))
	if got != expected {
		t.Errorf("NextVersion(%q) = %q, want %q", current, got, expected)
	}
}

func TestCalVer_NextVersion_YYYYFormat(t *testing.T) {
	cv := NewCalVer("yyyy.mm.minor", "calendar", "minor")
	now := time.Now()
	current := fmt.Sprintf("%d.%d.0", now.Year(), int(now.Month()))

	got, err := cv.NextVersion(current)
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	expected := fmt.Sprintf("%d.%d.1", now.Year(), int(now.Month()))
	if got != expected {
		t.Errorf("NextVersion(%q) = %q, want %q", current, got, expected)
	}
}

func TestCalVer_NextVersion_InvalidCurrent(t *testing.T) {
	cv := NewCalVer("yy.mm.minor", "calendar", "minor")

	got, err := cv.NextVersion("invalid")
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	now := time.Now()
	expected := fmt.Sprintf("%d.%d.0", now.Year()%100, int(now.Month()))
	if got != expected {
		t.Errorf("NextVersion(\"invalid\") = %q, want %q (should fallback to new version)", got, expected)
	}
}

func TestCalVer_Parse(t *testing.T) {
	cv := NewCalVer("yy.mm.minor", "calendar", "minor")

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "26.2.0", false},
		{"too few parts", "26.2", true},
		{"non-numeric year", "abc.2.0", true},
		{"non-numeric month", "26.abc.0", true},
		{"non-numeric minor", "26.2.abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cv.parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
