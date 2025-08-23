package runtime

import (
	"time"
)

// RuntimeConfig represents a runtime configuration loaded from runtime.yml
// Only includes fields that are actually implemented and used
type RuntimeConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Language    string            `yaml:"language" json:"language"`
	Version     string            `yaml:"version" json:"version"`
	Description string            `yaml:"description" json:"description"`
	Mounts      []MountSpec       `yaml:"mounts" json:"mounts"`
	Environment map[string]string `yaml:"environment" json:"environment"`

	// Keep only implemented features
	Requirements RuntimeRequirements `yaml:"requirements" json:"requirements"`
	Packages     []string            `yaml:"packages,omitempty" json:"packages,omitempty"`

	// Removed unused fields:
	// - Init string - not used anywhere in codebase
	// - PackageManager *PackageManagerConfig - defined but never implemented
}

// MountSpec defines how runtime directories should be mounted
type MountSpec struct {
	Source   string `yaml:"source" json:"source"` // Relative to runtime directory
	Target   string `yaml:"target" json:"target"` // Target in job chroot
	ReadOnly bool   `yaml:"readonly" json:"readonly"`

	// Removed unused fields:
	// - Selective []string - placeholder, not implemented (was ignored in parsing)
}

// PackageManagerConfig removed - was defined but never implemented
// If package manager integration is needed in the future, re-add this type

// RuntimeRequirements defines system requirements for a runtime
type RuntimeRequirements struct {
	Architectures []string `yaml:"architectures" json:"architectures"`

	// Removed unused fields:
	// - GPU bool - only validated, no actual GPU mounting implementation
	// If GPU support is needed, re-add this field and implement mounting logic
}

// RuntimeSpec represents a parsed runtime specification from the CLI
type RuntimeSpec struct {
	Language string   // e.g., "python", "java", "node"
	Version  string   // e.g., "3.11", "17", "18"
	Tags     []string // e.g., ["ml", "gpu"], ["scientific"]
}

// RuntimeInfo contains metadata about an available runtime
type RuntimeInfo struct {
	Name        string
	Language    string
	Version     string
	Description string
	Path        string
	Size        int64
	LastUsed    *time.Time
	Available   bool
}
