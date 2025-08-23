package installers

import (
	"context"
	"fmt"
	pb "joblet/api/gen"
	"os"
	"path/filepath"
)

// LocalRuntimeInstaller handles installation from local runtime directories
type LocalRuntimeInstaller struct {
	*BaseInstaller
	runtimesPath string
}

// NewLocalRuntimeInstaller creates a new local runtime installer
func NewLocalRuntimeInstaller(runtimesPath string) *LocalRuntimeInstaller {
	return &LocalRuntimeInstaller{
		BaseInstaller: NewBaseInstaller("local_runtime_install.sh"),
		runtimesPath:  runtimesPath,
	}
}

// Install installs a runtime from a local runtime directory
func (l *LocalRuntimeInstaller) Install(ctx context.Context, spec *InstallSpec) (*InstallResult, error) {
	// Find local setup script
	localRuntimePath := filepath.Join(l.runtimesPath, spec.RuntimeSpec)
	setupScriptPath := filepath.Join(localRuntimePath, "setup.sh")

	// Check if local runtime setup script exists
	if _, err := os.Stat(setupScriptPath); err != nil {
		return nil, fmt.Errorf("no local runtime found at %s", setupScriptPath)
	}

	// Prepare template data
	templateData := &TemplateData{
		RuntimeSpec:     spec.RuntimeSpec,
		SetupScriptPath: setupScriptPath,
		BuildArgs:       spec.BuildArgs,
	}

	// Render the installation script
	script, err := l.RenderTemplate(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate local runtime install script: %w", err)
	}

	return l.CreateInstallResult(script, true, fmt.Sprintf("Local runtime installation script generated for %s", spec.RuntimeSpec)), nil
}

// Validate checks if a local runtime exists
func (l *LocalRuntimeInstaller) Validate(source *pb.RuntimeSourceConfig) error {
	// Local runtime installer doesn't need source validation as it uses existing runtimes
	return nil
}

// GetSourceType returns the source type handled by this installer
func (l *LocalRuntimeInstaller) GetSourceType() string {
	return "local_runtime"
}

// CanInstall checks if this installer can handle the runtime spec
func (l *LocalRuntimeInstaller) CanInstall(runtimeSpec string) bool {
	localRuntimePath := filepath.Join(l.runtimesPath, runtimeSpec)
	setupScriptPath := filepath.Join(localRuntimePath, "setup.sh")

	_, err := os.Stat(setupScriptPath)
	return err == nil
}
