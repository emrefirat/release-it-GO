package config

import "testing"

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     string
		vars     map[string]string
		expected string
	}{
		{
			name:     "simple variable",
			tmpl:     "${version}",
			vars:     map[string]string{"version": "1.2.3"},
			expected: "1.2.3",
		},
		{
			name:     "multiple variables",
			tmpl:     "Release ${version} for ${name}",
			vars:     map[string]string{"version": "1.2.3", "name": "my-app"},
			expected: "Release 1.2.3 for my-app",
		},
		{
			name:     "unknown variable unchanged",
			tmpl:     "${unknown}",
			vars:     map[string]string{"version": "1.2.3"},
			expected: "${unknown}",
		},
		{
			name:     "empty template",
			tmpl:     "",
			vars:     map[string]string{"version": "1.2.3"},
			expected: "",
		},
		{
			name:     "no variables provided",
			tmpl:     "${version}",
			vars:     map[string]string{},
			expected: "${version}",
		},
		{
			name:     "nil vars",
			tmpl:     "${version}",
			vars:     nil,
			expected: "${version}",
		},
		{
			name:     "dotted variable",
			tmpl:     "${repo.owner}/${repo.repository}",
			vars:     map[string]string{"repo.owner": "emfi", "repo.repository": "release-it-go"},
			expected: "emfi/release-it-go",
		},
		{
			name:     "mixed known and unknown",
			tmpl:     "v${version} by ${author}",
			vars:     map[string]string{"version": "2.0.0"},
			expected: "v2.0.0 by ${author}",
		},
		{
			name:     "commit message template",
			tmpl:     "chore: release v${version}",
			vars:     map[string]string{"version": "1.3.0"},
			expected: "chore: release v1.3.0",
		},
		{
			name:     "no placeholders",
			tmpl:     "plain text without variables",
			vars:     map[string]string{"version": "1.0.0"},
			expected: "plain text without variables",
		},
		{
			name:     "variable at start and end",
			tmpl:     "${name}-${version}",
			vars:     map[string]string{"name": "app", "version": "1.0.0"},
			expected: "app-1.0.0",
		},
		{
			name:     "repeated variable",
			tmpl:     "${version} and ${version}",
			vars:     map[string]string{"version": "3.0.0"},
			expected: "3.0.0 and 3.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderTemplate(tt.tmpl, tt.vars)
			if got != tt.expected {
				t.Errorf("RenderTemplate(%q) = %q, expected %q", tt.tmpl, got, tt.expected)
			}
		})
	}
}
