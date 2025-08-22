package core

import (
	"fmt"
	"path/filepath"

	"joblet/internal/joblet/adapters"
	"joblet/internal/joblet/core/interfaces"
	"joblet/internal/joblet/core/volume"
	"joblet/internal/joblet/monitoring"
	"joblet/internal/joblet/monitoring/domain"
	"joblet/internal/joblet/pubsub"
	"joblet/internal/joblet/runtime"
	"joblet/internal/joblet/workflow"
	"joblet/pkg/buffer"
	"joblet/pkg/config"
	"joblet/pkg/logger"
	"joblet/pkg/platform"
)

// ComponentFactory handles creation and wiring of all system components
// This centralizes dependency injection and reduces coupling
type ComponentFactory struct {
	config         *config.Config
	platform       platform.Platform
	logger         *logger.Logger
	adapterFactory *adapters.AdapterFactory
}

// ServiceComponents contains all the core services needed by the application
type ServiceComponents struct {
	Joblet            interfaces.Joblet
	WorkflowManager   *workflow.WorkflowManager
	VolumeManager     *volume.Manager
	RuntimeManager    *runtime.Manager
	RuntimeResolver   *runtime.Resolver
	MonitoringService *monitoring.Service
	JobStore          JobStore
	NetworkStore      adapters.NetworkStoreAdapter
	VolumeStore       VolumeStore
}

// NewComponentFactory creates a new component factory
func NewComponentFactory(cfg *config.Config, platform platform.Platform) *ComponentFactory {
	factoryLogger := logger.New().WithField("component", "factory")
	return &ComponentFactory{
		config:         cfg,
		platform:       platform,
		logger:         factoryLogger,
		adapterFactory: adapters.NewAdapterFactory(factoryLogger),
	}
}

// CreateServices creates and wires all application services
func (f *ComponentFactory) CreateServices() (*ServiceComponents, error) {
	f.logger.Debug("initializing application services")

	// Create storage adapters first
	jobStore, err := f.createJobStore()
	if err != nil {
		return nil, err
	}

	networkStore, err := f.createNetworkStore()
	if err != nil {
		return nil, err
	}

	volumeStore, err := f.createVolumeStore()
	if err != nil {
		return nil, err
	}

	// Create managers
	volumeManager := f.createVolumeManager(volumeStore)
	runtimeManager, runtimeResolver := f.createRuntimeComponents()
	workflowManager := f.createWorkflowManager()

	// Create monitoring
	monitoringService := f.createMonitoringService()

	// Create core joblet with all dependencies
	joblet := NewJoblet(jobStore, f.config, networkStore)

	f.logger.Info("all application services initialized successfully")

	return &ServiceComponents{
		Joblet:            joblet,
		WorkflowManager:   workflowManager,
		VolumeManager:     volumeManager,
		RuntimeManager:    runtimeManager,
		RuntimeResolver:   runtimeResolver,
		MonitoringService: monitoringService,
		JobStore:          jobStore,
		NetworkStore:      networkStore,
		VolumeStore:       volumeStore,
	}, nil
}

// createJobStore creates and configures the job storage adapter
func (f *ComponentFactory) createJobStore() (JobStore, error) {
	f.logger.Debug("creating job store adapter")

	// Create adapter factory
	adapterFactory := adapters.NewAdapterFactory(f.logger)

	// Create default job store configuration for development
	jobStoreConfig := &adapters.JobStoreConfig{
		Store: &adapters.StoreConfig{
			Backend: "memory",
		},
		PubSub: &pubsub.PubSubConfig{
			BufferSize: 1000,
		},
		BufferManager: &adapters.BufferManagerConfig{
			DefaultBufferConfig: &buffer.BufferConfig{
				Type:                 "memory",
				MaxCapacity:          1048576, // 1MB
				MaxSubscribers:       10,
				SubscriberBufferSize: 1024,
				EnableMetrics:        false,
			},
			MaxBuffers:      100,
			CleanupInterval: "5m",
		},
		LogPersistence: &f.config.Buffers.LogPersistence,
	}

	// Create the adapter
	adapter, err := adapterFactory.CreateJobStoreAdapter(jobStoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create job store adapter: %w", err)
	}

	f.logger.Info("job store adapter created successfully")
	return adapter, nil
}

// createNetworkStore creates and configures the network storage adapter
func (f *ComponentFactory) createNetworkStore() (adapters.NetworkStoreAdapter, error) {
	f.logger.Debug("creating network store adapter")

	// Create adapter factory
	adapterFactory := adapters.NewAdapterFactory(f.logger)

	// Create default network store configuration for development
	networkStoreConfig := &adapters.NetworkStoreConfig{
		NetworkStore: &adapters.StoreConfig{
			Backend: "memory",
		},
		AllocationStore: &adapters.StoreConfig{
			Backend: "memory",
		},
	}

	// Create the adapter
	adapter, err := adapterFactory.CreateNetworkStoreAdapter(networkStoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create network store adapter: %w", err)
	}

	f.logger.Info("network store adapter created successfully")
	return adapter, nil
}

// createVolumeStore creates and configures the volume storage adapter
func (f *ComponentFactory) createVolumeStore() (VolumeStore, error) {
	f.logger.Debug("creating volume store adapter")

	// Create adapter factory
	adapterFactory := adapters.NewAdapterFactory(f.logger)

	// Create default volume store configuration for development
	volumeStoreConfig := &adapters.VolumeStoreConfig{
		Store: &adapters.StoreConfig{
			Backend: "memory",
		},
	}

	// Create the adapter
	adapter, err := adapterFactory.CreateVolumeStoreAdapter(volumeStoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume store adapter: %w", err)
	}

	f.logger.Info("volume store adapter created successfully")
	return adapter, nil
}

// createVolumeManager creates the volume manager with its dependencies
func (f *ComponentFactory) createVolumeManager(volumeStore VolumeStore) *volume.Manager {
	f.logger.Debug("creating volume manager")

	// Create volume manager with basic configuration
	basePath := "/opt/joblet/volumes" // Default volume path
	if f.config.Filesystem.BaseDir != "" {
		basePath = filepath.Join(f.config.Filesystem.BaseDir, "volumes")
	}
	manager := volume.NewManager(volumeStore, f.platform, basePath)

	f.logger.Info("volume manager created successfully")
	return manager
}

// createRuntimeComponents creates runtime manager and resolver
func (f *ComponentFactory) createRuntimeComponents() (*runtime.Manager, *runtime.Resolver) {
	f.logger.Debug("creating runtime components", "enabled", f.config.Runtime.Enabled)

	if f.config.Runtime.Enabled {
		f.logger.Info("runtime support enabled", "basePath", f.config.Runtime.BasePath)
		manager := runtime.NewManager(f.config.Runtime.BasePath, f.platform)
		resolver := runtime.NewResolver(f.config.Runtime.BasePath, f.platform)
		return manager, resolver
	}

	f.logger.Info("runtime support disabled - using minimal runtime components")
	// Create minimal runtime components for interface compliance
	manager := runtime.NewManager("", f.platform)
	resolver := runtime.NewResolver("", f.platform)
	return manager, resolver
}

// createWorkflowManager creates the workflow manager
func (f *ComponentFactory) createWorkflowManager() *workflow.WorkflowManager {
	f.logger.Debug("creating workflow manager")
	return workflow.NewWorkflowManager()
}

// createMonitoringService creates the monitoring service
func (f *ComponentFactory) createMonitoringService() *monitoring.Service {
	f.logger.Debug("creating monitoring service")
	// Use basic monitoring config - full implementation would map f.config properly
	return monitoring.NewService(&domain.MonitoringConfig{})
}
