package config

import (
	"regexp"
	"strings"
)

var templatePattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// RenderTemplate replaces ${var} placeholders in the template string with
// values from the vars map. Unknown variables are left unchanged.
func RenderTemplate(tmpl string, vars map[string]string) string {
	if tmpl == "" || len(vars) == 0 {
		return tmpl
	}

	return templatePattern.ReplaceAllStringFunc(tmpl, func(match string) string {
		key := match[2 : len(match)-1] // strip ${ and }

		if val, ok := vars[key]; ok {
			return val
		}

		// Support dotted keys: ${repo.owner} looks up "repo.owner" first,
		// then falls back to leaving it unchanged.
		if strings.Contains(key, ".") {
			if val, ok := vars[key]; ok {
				return val
			}
		}

		return match
	})
}
