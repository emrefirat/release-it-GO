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
		{"invalid month 0", "26.0.5", true},
		{"invalid month 13", "26.13.5", true},
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

func TestCalVer_Parse_YYYYMMDD(t *testing.T) {
	cv := NewCalVer("yyyy.mm.dd", "calendar", "minor")

	tests := []struct {
		name      string
		input     string
		wantDay   int
		wantMinor int
		wantErr   bool
	}{
		{"date only", "2024.3.15", 15, 0, false},
		{"date with minor", "2024.3.15.2", 15, 2, false},
		{"first of month", "2024.1.1", 1, 0, false},
		{"invalid month", "2024.13.5", 0, 0, true},
		{"non-numeric day", "2024.3.abc", 0, 0, true},
		{"non-numeric minor", "2024.3.15.abc", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := cv.parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if parts.Day != tt.wantDay {
				t.Errorf("Day = %d, want %d", parts.Day, tt.wantDay)
			}
			if parts.Minor != tt.wantMinor {
				t.Errorf("Minor = %d, want %d", parts.Minor, tt.wantMinor)
			}
		})
	}
}

func TestCalVer_NextVersion_YYYYMMDD_FirstRelease(t *testing.T) {
	cv := NewCalVer("yyyy.mm.dd", "calendar", "minor")
	now := time.Now()

	got, err := cv.NextVersion("")
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	expected := fmt.Sprintf("%d.%d.%d", now.Year(), int(now.Month()), now.Day())
	if got != expected {
		t.Errorf("NextVersion(\"\") = %q, want %q", got, expected)
	}
}

func TestCalVer_NextVersion_YYYYMMDD_SameDay(t *testing.T) {
	cv := NewCalVer("yyyy.mm.dd", "calendar", "minor")
	now := time.Now()
	current := fmt.Sprintf("%d.%d.%d", now.Year(), int(now.Month()), now.Day())

	got, err := cv.NextVersion(current)
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	// Same day → increment minor: 2024.3.15 → 2024.3.15.1
	expected := fmt.Sprintf("%d.%d.%d.1", now.Year(), int(now.Month()), now.Day())
	if got != expected {
		t.Errorf("NextVersion(%q) = %q, want %q", current, got, expected)
	}
}

func TestCalVer_NextVersion_YYYYMMDD_SameDaySecondIncrement(t *testing.T) {
	cv := NewCalVer("yyyy.mm.dd", "calendar", "minor")
	now := time.Now()
	current := fmt.Sprintf("%d.%d.%d.1", now.Year(), int(now.Month()), now.Day())

	got, err := cv.NextVersion(current)
	if err != nil {
		t.Fatalf("NextVersion error: %v", err)
	}

	// Same day, minor 1 → 2024.3.15.2
	expected := fmt.Sprintf("%d.%d.%d.2", now.Year(), int(now.Month()), now.Day())
	if got != expected {
		t.Errorf("NextVersion(%q) = %q, want %q", current, got, expected)
	}
}
