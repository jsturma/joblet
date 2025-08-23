package installers

import (
	"context"
	"fmt"
	pb "joblet/api/gen"
)

// Manager manages runtime installers and delegates installation requests
type Manager struct {
	installers   map[string]RuntimeInstaller
	localRuntime *LocalRuntimeInstaller
}

// NewManager creates a new installer manager
func NewManager(runtimesPath string) *Manager {
	localRuntime := NewLocalRuntimeInstaller(runtimesPath)

	return &Manager{
		installers: map[string]RuntimeInstaller{
			"github": NewGitHubInstaller(),
			"script": NewScriptInstaller(),
			"local":  NewLocalInstaller(),
		},
		localRuntime: localRuntime,
	}
}

// Install installs a runtime using the appropriate installer
func (m *Manager) Install(ctx context.Context, spec *InstallSpec) (*InstallResult, error) {
	// If no source is specified, try local runtime first
	if spec.Source == nil {
		if m.localRuntime.CanInstall(spec.RuntimeSpec) {
			return m.localRuntime.Install(ctx, spec)
		}
		return nil, fmt.Errorf("no source configuration provided and no local runtime found for %s", spec.RuntimeSpec)
	}

	// Get the appropriate installer based on source type
	installer, err := m.getInstaller(spec.Source)
	if err != nil {
		return nil, err
	}

	// Validate the source configuration
	if err := installer.Validate(spec.Source); err != nil {
		return nil, fmt.Errorf("invalid source configuration: %w", err)
	}

	// Install the runtime
	return installer.Install(ctx, spec)
}

// getInstaller returns the appropriate installer for the given source type
func (m *Manager) getInstaller(source *pb.RuntimeSourceConfig) (RuntimeInstaller, error) {
	switch source.SourceType.(type) {
	case *pb.RuntimeSourceConfig_Github:
		return m.installers["github"], nil
	case *pb.RuntimeSourceConfig_Script:
		return m.installers["script"], nil
	case *pb.RuntimeSourceConfig_Local:
		return m.installers["local"], nil
	default:
		return nil, fmt.Errorf("unsupported source type")
	}
}

// GetAvailableInstallers returns the list of available installer types
func (m *Manager) GetAvailableInstallers() []string {
	installers := make([]string, 0, len(m.installers))
	for name := range m.installers {
		installers = append(installers, name)
	}
	installers = append(installers, m.localRuntime.GetSourceType())
	return installers
}

// ValidateSource validates a source configuration
func (m *Manager) ValidateSource(source *pb.RuntimeSourceConfig) error {
	installer, err := m.getInstaller(source)
	if err != nil {
		return err
	}
	return installer.Validate(source)
}
