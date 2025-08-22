package core

import (
	"joblet/pkg/config"
	"joblet/pkg/platform"
)

// ExampleCreateServices shows how to use the ComponentFactory
// to create all services with proper dependency injection
func ExampleCreateServices(cfg *config.Config) (*ServiceComponents, error) {
	// Create platform abstraction
	platformInterface := platform.NewPlatform()

	// Create factory
	factory := NewComponentFactory(cfg, platformInterface)

	// Create all services with proper dependency injection
	services, err := factory.CreateServices()
	if err != nil {
		return nil, err
	}

	// All services are now ready to use:
	// - services.Joblet (core business logic)
	// - services.WorkflowManager
	// - services.VolumeManager
	// - services.RuntimeManager
	// - services.MonitoringService
	// - Storage adapters (JobStore, NetworkStore, VolumeStore)

	return services, nil
}

// ExampleServerSetup demonstrates how a server would use the factory
// instead of manually creating and wiring components
func ExampleServerSetup(cfg *config.Config) error {
	// Create all services using factory
	services, err := ExampleCreateServices(cfg)
	if err != nil {
		return err
	}

	// Server layer only needs to worry about HTTP/gRPC handlers
	// and no longer needs to know about component wiring
	_ = services // Use services to create handlers

	return nil
}
