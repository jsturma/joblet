package installers

import (
	"context"
	pb "joblet/api/gen"
)

// RuntimeInstaller defines the interface for installing runtimes from different sources
type RuntimeInstaller interface {
	// Install installs a runtime from the configured source
	Install(ctx context.Context, spec *InstallSpec) (*InstallResult, error)

	// Validate checks if the installer can handle the given source configuration
	Validate(source *pb.RuntimeSourceConfig) error

	// GetSourceType returns the source type this installer handles
	GetSourceType() string
}

// InstallSpec contains the parameters needed for runtime installation
type InstallSpec struct {
	RuntimeSpec string
	Source      *pb.RuntimeSourceConfig
	BuildArgs   map[string]string
	TargetPath  string
}

// InstallResult contains the result of a runtime installation
type InstallResult struct {
	Success bool
	Command string
	Args    []string
	Error   error
	Message string
}

// TemplateData contains the data passed to installation templates
type TemplateData struct {
	RuntimeSpec     string
	Repository      string
	Branch          string
	Path            string
	Dependencies    string
	SetupScript     string
	BuildScript     string
	SetupScriptPath string
	BuildArgs       map[string]string
}
