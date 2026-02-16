package version

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple", "1.2.3", "1.2.3", false},
		{"with v prefix", "v1.2.3", "1.2.3", false},
		{"pre-release", "1.2.3-beta.1", "1.2.3-beta.1", false},
		{"with build metadata", "1.2.3+build.123", "1.2.3+build.123", false},
		{"zero version", "0.0.0", "0.0.0", false},
		{"empty string", "", "", true},
		{"invalid", "not-a-version", "", true},
		{"spaces", "  1.2.3  ", "1.2.3", false},
		{"v prefix with spaces", " v1.2.3 ", "1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("ParseVersion(%q) = %q, want %q", tt.input, got.String(), tt.want)
			}
		})
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		incType string
		preID   string
		want    string
		wantErr bool
	}{
		{"patch", "1.2.3", "patch", "", "1.2.4", false},
		{"minor", "1.2.3", "minor", "", "1.3.0", false},
		{"major", "1.2.3", "major", "", "2.0.0", false},
		{"premajor beta", "1.2.3", "premajor", "beta", "2.0.0-beta.0", false},
		{"preminor alpha", "1.2.3", "preminor", "alpha", "1.3.0-alpha.0", false},
		{"prepatch rc", "1.2.3", "prepatch", "rc", "1.2.4-rc.0", false},
		{"prerelease new", "1.2.3", "prerelease", "beta", "1.2.4-beta.0", false},
		{"prerelease increment same id", "2.0.0-beta.0", "prerelease", "beta", "2.0.0-beta.1", false},
		{"prerelease different id", "2.0.0-beta.1", "prerelease", "rc", "2.0.0-rc.0", false},
		{"invalid type", "1.2.3", "invalid", "", "", true},
		{"major from zero", "0.1.0", "major", "", "1.0.0", false},
		{"patch from pre-release", "1.2.3-beta.0", "patch", "", "1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, err := ParseVersion(tt.current)
			if err != nil {
				t.Fatalf("failed to parse current version: %v", err)
			}

			got, err := IncrementVersion(current, tt.incType, tt.preID)
			if (err != nil) != tt.wantErr {
				t.Errorf("IncrementVersion error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("IncrementVersion = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestIncrementVersion_NilCurrent(t *testing.T) {
	_, err := IncrementVersion(nil, "patch", "")
	if err == nil {
		t.Error("expected error for nil current version")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal", "1.2.3", "1.2.3", 0},
		{"a less than b", "1.2.3", "1.2.4", -1},
		{"a greater than b", "2.0.0", "1.9.9", 1},
		{"pre-release less than release", "1.2.3-beta.1", "1.2.3", -1},
		{"with v prefix", "v1.2.3", "1.2.3", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareVersions(tt.a, tt.b)
			if err != nil {
				t.Fatalf("CompareVersions error: %v", err)
			}
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCompareVersions_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
	}{
		{"invalid a", "invalid", "1.2.3"},
		{"invalid b", "1.2.3", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CompareVersions(tt.a, tt.b)
			if err == nil {
				t.Error("expected error for invalid version")
			}
		})
	}
}

func TestFormatVersion(t *testing.T) {
	v, _ := ParseVersion("1.2.3")
	got := FormatVersion(v)
	if got != "1.2.3" {
		t.Errorf("FormatVersion = %q, want %q", got, "1.2.3")
	}
}
