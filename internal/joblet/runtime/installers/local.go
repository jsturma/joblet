package installers

import (
	"context"
	"fmt"
	pb "joblet/api/gen"
)

// LocalInstaller handles runtime installation from local files
type LocalInstaller struct {
	*BaseInstaller
}

// NewLocalInstaller creates a new local runtime installer
func NewLocalInstaller() *LocalInstaller {
	return &LocalInstaller{
		BaseInstaller: NewBaseInstaller("local_install.sh"),
	}
}

// Install installs a runtime from local files
func (l *LocalInstaller) Install(ctx context.Context, spec *InstallSpec) (*InstallResult, error) {
	local, ok := spec.Source.SourceType.(*pb.RuntimeSourceConfig_Local)
	if !ok {
		return nil, fmt.Errorf("invalid source type for Local installer")
	}

	// Prepare template data
	templateData := &TemplateData{
		RuntimeSpec: spec.RuntimeSpec,
		BuildScript: local.Local.BuildScript,
		BuildArgs:   spec.BuildArgs,
	}

	// Render the installation script
	script, err := l.RenderTemplate(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate local install script: %w", err)
	}

	return l.CreateInstallResult(script, true, fmt.Sprintf("Local installation script generated for %s", spec.RuntimeSpec)), nil
}

// Validate checks if the local source configuration is valid
func (l *LocalInstaller) Validate(source *pb.RuntimeSourceConfig) error {
	local, ok := source.SourceType.(*pb.RuntimeSourceConfig_Local)
	if !ok {
		return fmt.Errorf("invalid source type for Local installer")
	}

	if len(local.Local.Files) == 0 {
		return fmt.Errorf("local installer requires at least one file")
	}

	if local.Local.BuildScript == "" {
		return fmt.Errorf("build script cannot be empty")
	}

	return nil
}

// GetSourceType returns the source type handled by this installer
func (l *LocalInstaller) GetSourceType() string {
	return "local"
}
