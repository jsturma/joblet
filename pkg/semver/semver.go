package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a semantic version with major, minor, and patch components
type Version struct {
	major int
	minor int
	patch int
}

// NewVersion parses a semantic version string (e.g., "1.2.3")
// Returns an error if the version string is not in valid MAJOR.MINOR.PATCH format
func NewVersion(version string) (*Version, error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: expected MAJOR.MINOR.PATCH, got %q", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	if major < 0 || minor < 0 || patch < 0 {
		return nil, fmt.Errorf("version components must be non-negative")
	}

	return &Version{
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}

// GreaterThan returns true if v is greater than other
func (v *Version) GreaterThan(other *Version) bool {
	if v.major != other.major {
		return v.major > other.major
	}
	if v.minor != other.minor {
		return v.minor > other.minor
	}
	return v.patch > other.patch
}

// String returns the string representation of the version
func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

// Equal returns true if v equals other
func (v *Version) Equal(other *Version) bool {
	return v.major == other.major && v.minor == other.minor && v.patch == other.patch
}

// LessThan returns true if v is less than other
func (v *Version) LessThan(other *Version) bool {
	return !v.GreaterThan(other) && !v.Equal(other)
}
