// Package version provides version detection and manipulation for release-it-go.
// It supports semantic versioning (semver) and calendar versioning (calver).
package version

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// ErrInvalidVersion is returned when a version string cannot be parsed.
var ErrInvalidVersion = fmt.Errorf("invalid version")

// ParseVersion parses a version string into a semver.Version.
// Accepts formats: "1.2.3", "v1.2.3", "1.2.3-beta.1".
func ParseVersion(v string) (*semver.Version, error) {
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, fmt.Errorf("%w: empty version string", ErrInvalidVersion)
	}

	parsed, err := semver.NewVersion(v)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidVersion, v)
	}

	return parsed, nil
}

// IncrementVersion increments the version based on the given type.
// Supported types: "major", "minor", "patch", "premajor", "preminor", "prepatch", "prerelease".
func IncrementVersion(current *semver.Version, incrementType string, preReleaseID string) (*semver.Version, error) {
	if current == nil {
		return nil, fmt.Errorf("current version is nil")
	}

	switch incrementType {
	case "major":
		next := current.IncMajor()
		return &next, nil

	case "minor":
		next := current.IncMinor()
		return &next, nil

	case "patch":
		next := current.IncPatch()
		return &next, nil

	case "premajor":
		next := current.IncMajor()
		return addPreRelease(next, preReleaseID, 0)

	case "preminor":
		next := current.IncMinor()
		return addPreRelease(next, preReleaseID, 0)

	case "prepatch":
		next := current.IncPatch()
		return addPreRelease(next, preReleaseID, 0)

	case "prerelease":
		return incrementPreRelease(current, preReleaseID)

	default:
		return nil, fmt.Errorf("unsupported increment type: %s", incrementType)
	}
}

// CompareVersions compares two version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func CompareVersions(a, b string) (int, error) {
	va, err := ParseVersion(a)
	if err != nil {
		return 0, fmt.Errorf("parsing version a: %w", err)
	}

	vb, err := ParseVersion(b)
	if err != nil {
		return 0, fmt.Errorf("parsing version b: %w", err)
	}

	return va.Compare(vb), nil
}

// FormatVersion returns the version string without the "v" prefix.
func FormatVersion(v *semver.Version) string {
	return v.String()
}

// addPreRelease creates a new version with the given pre-release identifier and number.
// If preReleaseID is empty, defaults to "0" producing versions like "1.0.0-0.0".
// Callers should validate preReleaseID before calling if a meaningful ID is required.
func addPreRelease(base semver.Version, preReleaseID string, num int) (*semver.Version, error) {
	if preReleaseID == "" {
		preReleaseID = "0"
	}

	preStr := fmt.Sprintf("%s.%d", preReleaseID, num)
	next, err := base.SetPrerelease(preStr)
	if err != nil {
		return nil, fmt.Errorf("setting pre-release: %w", err)
	}

	return &next, nil
}

// incrementPreRelease increments the pre-release component of a version.
func incrementPreRelease(current *semver.Version, preReleaseID string) (*semver.Version, error) {
	pre := current.Prerelease()
	if pre == "" {
		// Not a pre-release yet, bump patch and add pre-release
		next := current.IncPatch()
		return addPreRelease(next, preReleaseID, 0)
	}

	currentID, currentNum := parsePreRelease(pre)

	if preReleaseID == "" || currentID == preReleaseID {
		// Same ID, increment number
		return setPreRelease(*current, currentID, currentNum+1)
	}

	// Different ID, start new pre-release series
	return setPreRelease(*current, preReleaseID, 0)
}

// parsePreRelease splits a pre-release string like "beta.1" into ("beta", 1).
func parsePreRelease(pre string) (string, int) {
	parts := strings.Split(pre, ".")
	if len(parts) < 2 {
		return pre, 0
	}

	num, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return pre, 0
	}

	id := strings.Join(parts[:len(parts)-1], ".")
	return id, num
}

// setPreRelease creates a new version with the given pre-release id and number.
func setPreRelease(base semver.Version, id string, num int) (*semver.Version, error) {
	preStr := fmt.Sprintf("%s.%d", id, num)
	next, err := base.SetPrerelease(preStr)
	if err != nil {
		return nil, fmt.Errorf("setting pre-release: %w", err)
	}
	return &next, nil
}
