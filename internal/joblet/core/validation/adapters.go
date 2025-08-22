package validation

import (
	"context"
	"joblet/internal/joblet/core/volume"
	"joblet/internal/joblet/runtime"
	"joblet/pkg/logger"
)

// VolumeManagerAdapter adapts the volume.Manager to VolumeManagerInterface
type VolumeManagerAdapter struct {
	manager *volume.Manager
	logger  *logger.Logger
}

// NewVolumeManagerAdapter creates a new adapter for volume.Manager
func NewVolumeManagerAdapter(manager *volume.Manager) *VolumeManagerAdapter {
	return &VolumeManagerAdapter{
		manager: manager,
		logger:  logger.WithField("component", "volume-manager-adapter"),
	}
}

// VolumeExists checks if a volume exists using the volume manager
func (vma *VolumeManagerAdapter) VolumeExists(volumeName string) bool {
	_, exists := vma.manager.GetVolume(volumeName)
	return exists
}

// RuntimeManagerAdapter adapts the runtime.Manager to RuntimeManagerInterface
type RuntimeManagerAdapter struct {
	manager  *runtime.Manager
	resolver *runtime.Resolver
	logger   *logger.Logger
}

// NewRuntimeManagerAdapter creates a new adapter for runtime.Manager
func NewRuntimeManagerAdapter(manager *runtime.Manager, resolver *runtime.Resolver) *RuntimeManagerAdapter {
	return &RuntimeManagerAdapter{
		manager:  manager,
		resolver: resolver,
		logger:   logger.WithField("component", "runtime-manager-adapter"),
	}
}

// RuntimeExists checks if a runtime exists and is available
func (rma *RuntimeManagerAdapter) RuntimeExists(runtimeName string) bool {
	// Try to resolve the runtime configuration
	ctx := context.Background()
	_, err := rma.manager.ResolveRuntimeConfig(ctx, runtimeName)
	return err == nil
}

// ListRuntimes returns all available runtimes
func (rma *RuntimeManagerAdapter) ListRuntimes() []RuntimeInfo {
	// Get runtime info from resolver
	runtimeInfos, err := rma.resolver.ListRuntimes()
	if err != nil {
		rma.logger.Error("failed to list runtimes", "error", err)
		return []RuntimeInfo{}
	}

	// Convert to our validation runtime info format
	result := make([]RuntimeInfo, len(runtimeInfos))
	for i, info := range runtimeInfos {
		result[i] = RuntimeInfo{
			Name:      info.Name,
			Version:   info.Version,
			Available: info.Available,
		}
	}

	return result
}
