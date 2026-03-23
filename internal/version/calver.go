package version

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CalVer handles calendar-based versioning.
type CalVer struct {
	Format            string
	Increment         string
	FallbackIncrement string
}

// CalVerParts holds the parsed components of a calendar version.
type CalVerParts struct {
	Year  int
	Month int
	Day   int
	Minor int
}

// NewCalVer creates a CalVer instance from config values.
func NewCalVer(format, increment, fallbackIncrement string) *CalVer {
	return &CalVer{
		Format:            format,
		Increment:         increment,
		FallbackIncrement: fallbackIncrement,
	}
}

// NextVersion calculates the next calendar version based on the current version.
func (cv *CalVer) NextVersion(current string) (string, error) {
	now := time.Now()

	if current == "" {
		return cv.format(now, 0), nil
	}

	parts, err := cv.parse(current)
	if err != nil {
		return cv.format(now, 0), nil
	}

	if cv.calendarChanged(parts, now) {
		return cv.format(now, 0), nil
	}

	return cv.format(now, parts.Minor+1), nil
}

// parse splits a CalVer string into its components.
// For "yyyy.mm.dd" format: parts[2] = Day, parts[3] = Minor (optional).
// For other formats: parts[2] = Minor.
func (cv *CalVer) parse(version string) (*CalVerParts, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid calver format: %s", version)
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid year in calver: %s", parts[0])
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid month in calver: %s", parts[1])
	}

	if month < 1 || month > 12 {
		return nil, fmt.Errorf("invalid month in calver (must be 1-12): %d", month)
	}

	third, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid component in calver: %s", parts[2])
	}

	result := &CalVerParts{
		Year:  year,
		Month: month,
	}

	if cv.Format == "yyyy.mm.dd" {
		result.Day = third
		// Optional 4th component is minor (for multiple releases per day)
		if len(parts) >= 4 {
			minor, err := strconv.Atoi(parts[3])
			if err != nil {
				return nil, fmt.Errorf("invalid minor in calver: %s", parts[3])
			}
			result.Minor = minor
		}
	} else {
		result.Minor = third
	}

	return result, nil
}

// format creates a CalVer string from a time and minor number.
func (cv *CalVer) format(t time.Time, minor int) string {
	switch cv.Format {
	case "yyyy.mm.minor":
		return fmt.Sprintf("%d.%d.%d", t.Year(), int(t.Month()), minor)
	case "yyyy.mm.dd":
		if minor > 0 {
			return fmt.Sprintf("%d.%d.%d.%d", t.Year(), int(t.Month()), t.Day(), minor)
		}
		return fmt.Sprintf("%d.%d.%d", t.Year(), int(t.Month()), t.Day())
	default: // "yy.mm.minor"
		return fmt.Sprintf("%d.%d.%d", t.Year()%100, int(t.Month()), minor)
	}
}

// calendarChanged checks if the calendar period has changed.
func (cv *CalVer) calendarChanged(parts *CalVerParts, now time.Time) bool {
	currentYear := now.Year()

	// Normalize 2-digit year
	parsedYear := parts.Year
	if cv.Format == "yy.mm.minor" && parsedYear < 100 {
		parsedYear += 2000
	}

	switch cv.Format {
	case "yyyy.mm.dd":
		return parsedYear != currentYear || parts.Month != int(now.Month()) || parts.Day != now.Day()
	default: // "yy.mm.minor", "yyyy.mm.minor"
		return parsedYear != currentYear || parts.Month != int(now.Month())
	}
}
