package installers

import (
	"context"
	"fmt"
	pb "joblet/api/gen"
	"strings"
)

// ScriptInstaller handles runtime installation from direct scripts
type ScriptInstaller struct {
	*BaseInstaller
}

// NewScriptInstaller creates a new script runtime installer
func NewScriptInstaller() *ScriptInstaller {
	return &ScriptInstaller{
		BaseInstaller: NewBaseInstaller("script_install.sh"),
	}
}

// Install installs a runtime using a direct script
func (s *ScriptInstaller) Install(ctx context.Context, spec *InstallSpec) (*InstallResult, error) {
	script, ok := spec.Source.SourceType.(*pb.RuntimeSourceConfig_Script)
	if !ok {
		return nil, fmt.Errorf("invalid source type for Script installer")
	}

	// Prepare template data
	templateData := &TemplateData{
		RuntimeSpec:  spec.RuntimeSpec,
		SetupScript:  script.Script.SetupScript,
		Dependencies: strings.Join(script.Script.Dependencies, " "),
		BuildArgs:    spec.BuildArgs,
	}

	// Render the installation script
	installScript, err := s.RenderTemplate(templateData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate script install script: %w", err)
	}

	return s.CreateInstallResult(installScript, true, fmt.Sprintf("Script installation script generated for %s", spec.RuntimeSpec)), nil
}

// Validate checks if the script source configuration is valid
func (s *ScriptInstaller) Validate(source *pb.RuntimeSourceConfig) error {
	script, ok := source.SourceType.(*pb.RuntimeSourceConfig_Script)
	if !ok {
		return fmt.Errorf("invalid source type for Script installer")
	}

	if script.Script.SetupScript == "" {
		return fmt.Errorf("setup script cannot be empty")
	}

	return nil
}

// GetSourceType returns the source type handled by this installer
func (s *ScriptInstaller) GetSourceType() string {
	return "script"
}
