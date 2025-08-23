package domain

// JobType represents the type of job execution
type JobType string

const (
	// JobTypeStandard represents a normal production job with minimal chroot
	JobTypeStandard JobType = "standard"

	// JobTypeRuntimeBuild represents a runtime build job with builder chroot
	// Builder chroot provides full host OS environment (minus /opt/joblet) for
	// compiling and packaging runtime environments
	JobTypeRuntimeBuild JobType = "runtime-build"
)

// IsRuntimeBuild returns true if this is a runtime build job
func (jt JobType) IsRuntimeBuild() bool {
	return jt == JobTypeRuntimeBuild
}

// String returns the string representation of the job type
func (jt JobType) String() string {
	if jt == "" {
		return string(JobTypeStandard) // Default to standard if not specified
	}
	return string(jt)
}
