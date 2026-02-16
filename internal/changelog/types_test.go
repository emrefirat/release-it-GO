package changelog

import "testing"

func TestBumpType_String(t *testing.T) {
	tests := []struct {
		name string
		bump BumpType
		want string
	}{
		{"none", BumpNone, ""},
		{"patch", BumpPatch, "patch"},
		{"minor", BumpMinor, "minor"},
		{"major", BumpMajor, "major"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bump.String()
			if got != tt.want {
				t.Errorf("BumpType(%d).String() = %q, want %q", tt.bump, got, tt.want)
			}
		})
	}
}
