package bumper

import (
	"testing"

	"github.com/emfi/release-it-go/internal/config"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		file     config.BumperFile
		expected string
	}{
		{"explicit json type", config.BumperFile{File: "file.txt", Type: "json"}, "json"},
		{"explicit yaml type", config.BumperFile{File: "file.txt", Type: "YAML"}, "yaml"},
		{"json extension", config.BumperFile{File: "package.json"}, "json"},
		{"yaml extension", config.BumperFile{File: "chart.yaml"}, "yaml"},
		{"yml extension", config.BumperFile{File: "config.yml"}, "yaml"},
		{"toml extension", config.BumperFile{File: "pyproject.toml"}, "toml"},
		{"ini extension", config.BumperFile{File: "setup.ini"}, "ini"},
		{"cfg extension", config.BumperFile{File: "setup.cfg"}, "ini"},
		{"text fallback", config.BumperFile{File: "VERSION"}, "text"},
		{"txt extension", config.BumperFile{File: "version.txt"}, "text"},
		{"unknown extension", config.BumperFile{File: "file.xyz"}, "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFormat(tt.file)
			if got != tt.expected {
				t.Errorf("detectFormat(%q) = %q, want %q", tt.file.File, got, tt.expected)
			}
		})
	}
}
