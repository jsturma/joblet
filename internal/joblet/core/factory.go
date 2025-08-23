package core

import (
	"context"
	"fmt"
	"path/filepath"

	"joblet/internal/joblet/adapters"
	"joblet/internal/joblet/core/interfaces"
	"joblet/internal/joblet/core/volume"
	"joblet/internal/joblet/domain"
	"joblet/internal/joblet/events"
	interfaces_ext "joblet/internal/joblet/interfaces"
	"joblet/internal/joblet/monitoring"
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
	RuntimeResolver   *runtime.Resolver
	MonitoringService *monitoring.Service
	JobStore          JobStore
	NetworkStore      adapters.NetworkStoreAdapter
	VolumeStore       VolumeStore
	EventBus          events.EventBus
	ServiceRegistry   interfaces_ext.ServiceRegistry
}

// NewComponentFactory creates a new component factory for dependency injection
func NewComponentFactory(cfg *config.Config, platform platform.Platform) *ComponentFactory {
	factoryLogger := logger.New().WithField("component", "factory")
	return &ComponentFactory{
		config:         cfg,
		platform:       platform,
		logger:         factoryLogger,
		adapterFactory: adapters.NewAdapterFactory(factoryLogger),
	}
}

// CreateServices creates and wires all application services with proper dependency injection
func (f *ComponentFactory) CreateServices() (*ServiceComponents, error) {
	f.logger.Debug("initializing application services")

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

	volumeManager := f.createVolumeManager(volumeStore)
	runtimeResolver := f.createRuntimeComponents()
	workflowManager := f.createWorkflowManager()
	monitoringService := f.createMonitoringService()

	f.configureVolumeMonitoring(monitoringService, volumeManager)

	eventBus := events.NewInMemoryEventBus()
	joblet := NewJoblet(jobStore, f.config, networkStore)
	serviceRegistry := f.createServiceRegistry(joblet, volumeManager, networkStore, monitoringService, runtimeResolver)

	f.logger.Info("all application services initialized successfully")

	return &ServiceComponents{
		Joblet:            joblet,
		WorkflowManager:   workflowManager,
		VolumeManager:     volumeManager,
		RuntimeResolver:   runtimeResolver,
		MonitoringService: monitoringService,
		JobStore:          jobStore,
		NetworkStore:      networkStore,
		VolumeStore:       volumeStore,
		EventBus:          eventBus,
		ServiceRegistry:   serviceRegistry,
	}, nil
}

// createJobStore creates and configures the job storage adapter with in-memory backend
func (f *ComponentFactory) createJobStore() (JobStore, error) {
	f.logger.Debug("creating job store adapter")

	adapterFactory := adapters.NewAdapterFactory(f.logger)
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

	adapter, err := adapterFactory.CreateJobStoreAdapter(jobStoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create job store adapter: %w", err)
	}

	f.logger.Info("job store adapter created successfully")
	return adapter, nil
}

// createNetworkStore creates and configures the network storage adapter for job networking
func (f *ComponentFactory) createNetworkStore() (adapters.NetworkStoreAdapter, error) {
	f.logger.Debug("creating network store adapter")

	adapterFactory := adapters.NewAdapterFactory(f.logger)
	networkStoreConfig := &adapters.NetworkStoreConfig{
		NetworkStore: &adapters.StoreConfig{
			Backend: "memory",
		},
		AllocationStore: &adapters.StoreConfig{
			Backend: "memory",
		},
	}

	adapter, err := adapterFactory.CreateNetworkStoreAdapter(networkStoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create network store adapter: %w", err)
	}

	f.logger.Info("network store adapter created successfully")
	return adapter, nil
}

// createVolumeStore creates and configures the volume storage adapter for persistent data
func (f *ComponentFactory) createVolumeStore() (VolumeStore, error) {
	f.logger.Debug("creating volume store adapter")

	adapterFactory := adapters.NewAdapterFactory(f.logger)
	volumeStoreConfig := &adapters.VolumeStoreConfig{
		Store: &adapters.StoreConfig{
			Backend: "memory",
		},
	}

	adapter, err := adapterFactory.CreateVolumeStoreAdapter(volumeStoreConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume store adapter: %w", err)
	}

	f.logger.Info("volume store adapter created successfully")
	return adapter, nil
}

// createVolumeManager creates a volume manager to handle persistent volume operations
func (f *ComponentFactory) createVolumeManager(volumeStore VolumeStore) *volume.Manager {
	f.logger.Debug("creating volume manager")

	basePath := "/opt/joblet/volumes"
	if f.config.Filesystem.BaseDir != "" {
		basePath = filepath.Join(f.config.Filesystem.BaseDir, "volumes")
	}
	manager := volume.NewManager(volumeStore, f.platform, basePath)

	f.logger.Info("volume manager created successfully")
	return manager
}

// createRuntimeComponents creates a runtime resolver to manage execution environments
func (f *ComponentFactory) createRuntimeComponents() *runtime.Resolver {
	f.logger.Debug("creating runtime resolver", "basePath", f.config.Runtime.BasePath)
	return runtime.NewResolver(f.config.Runtime.BasePath, f.platform)
}

// createWorkflowManager creates a workflow manager to handle multi-job orchestration
func (f *ComponentFactory) createWorkflowManager() *workflow.WorkflowManager {
	f.logger.Debug("creating workflow manager")
	return workflow.NewWorkflowManager()
}

// createMonitoringService creates a monitoring service for system and job metrics collection
func (f *ComponentFactory) createMonitoringService() *monitoring.Service {
	f.logger.Debug("creating monitoring service")
	return monitoring.NewService(nil)
}

// configureVolumeMonitoring wires volume manager with monitoring service for disk usage tracking
func (f *ComponentFactory) configureVolumeMonitoring(monitoringService *monitoring.Service, volumeManager *volume.Manager) {
	f.logger.Debug("configuring volume monitoring integration")

	basePath := "/opt/joblet/volumes"
	if f.config.Filesystem.BaseDir != "" {
		basePath = filepath.Join(f.config.Filesystem.BaseDir, "volumes")
	}

	monitoringService.SetVolumeManager(volumeManager, basePath)
	f.logger.Info("volume monitoring integration configured", "volumeBasePath", basePath)
}

// createServiceRegistry creates a centralized service registry for dependency access
func (f *ComponentFactory) createServiceRegistry(
	joblet interfaces.Joblet,
	volumeManager *volume.Manager,
	networkStore adapters.NetworkStoreAdapter,
	monitoringService *monitoring.Service,
	runtimeResolver *runtime.Resolver) interfaces_ext.ServiceRegistry {

	f.logger.Debug("creating service registry")
	return &serviceRegistryImpl{
		joblet:            joblet,
		volumeManager:     volumeManager,
		networkStore:      networkStore,
		monitoringService: monitoringService,
		runtimeResolver:   runtimeResolver,
	}
}

type serviceRegistryImpl struct {
	joblet            interfaces.Joblet
	volumeManager     *volume.Manager
	networkStore      adapters.NetworkStoreAdapter
	monitoringService *monitoring.Service
	runtimeResolver   *runtime.Resolver
}

func (r *serviceRegistryImpl) GetJobService() interfaces_ext.JobServiceInterface {
	return &jobServiceWrapper{joblet: r.joblet}
}

func (r *serviceRegistryImpl) GetVolumeService() interfaces_ext.VolumeServiceInterface {
	return &volumeServiceWrapper{manager: r.volumeManager}
}

func (r *serviceRegistryImpl) GetNetworkService() interfaces_ext.NetworkServiceInterface {
	return &networkServiceWrapper{store: r.networkStore}
}

func (r *serviceRegistryImpl) GetMonitoringService() interfaces_ext.MonitoringServiceInterface {
	return &monitoringServiceWrapper{service: r.monitoringService}
}

func (r *serviceRegistryImpl) GetRuntimeService() interfaces_ext.RuntimeServiceInterface {
	return &runtimeServiceWrapper{resolver: r.runtimeResolver}
}

// Service wrappers - these adapt internal implementations to the interface contracts
type jobServiceWrapper struct {
	joblet interfaces.Joblet
}

func (j *jobServiceWrapper) StartJob(ctx context.Context, req interfaces.StartJobRequest) (*domain.Job, error) {
	return j.joblet.StartJob(ctx, req)
}

func (j *jobServiceWrapper) StopJob(ctx context.Context, req interfaces.StopJobRequest) error {
	return j.joblet.StopJob(ctx, req)
}

func (j *jobServiceWrapper) DeleteJob(ctx context.Context, req interfaces.DeleteJobRequest) error {
	return j.joblet.DeleteJob(ctx, req)
}

func (j *jobServiceWrapper) ExecuteScheduledJob(ctx context.Context, req interfaces.ExecuteScheduledJobRequest) error {
	return j.joblet.ExecuteScheduledJob(ctx, req)
}

func (j *jobServiceWrapper) GetJobStatus(ctx context.Context, jobID string) (*domain.Job, bool) {
	// This would need to be implemented via a job store interface
	return nil, false
}

func (j *jobServiceWrapper) ListJobs(ctx context.Context) []*domain.Job {
	// This would need to be implemented via a job store interface
	return nil
}

// Additional service wrappers
type volumeServiceWrapper struct {
	manager *volume.Manager
}

func (v *volumeServiceWrapper) CreateVolume(ctx context.Context, name string, size int64) error {
	// Adapt to the actual volume manager interface signature
	_, err := v.manager.CreateVolume(name, "", domain.VolumeTypeFilesystem)
	return err
}

func (v *volumeServiceWrapper) DeleteVolume(ctx context.Context, name string) error {
	// Implement delete - may need to add to volume.Manager
	return fmt.Errorf("delete volume not implemented")
}

func (v *volumeServiceWrapper) ListVolumes(ctx context.Context) ([]string, error) {
	volumes := v.manager.ListVolumes()
	names := make([]string, len(volumes))
	for i, vol := range volumes {
		names[i] = vol.Name
	}
	return names, nil
}

func (v *volumeServiceWrapper) MountVolume(ctx context.Context, volumeName, mountPath string) error {
	// Implement mount - may need to add to volume.Manager
	return fmt.Errorf("mount volume not implemented")
}

type networkServiceWrapper struct {
	store adapters.NetworkStoreAdapter
}

func (n *networkServiceWrapper) CreateNetwork(ctx context.Context, name string, config interface{}) error {
	// Implementation would depend on the network store interface
	return fmt.Errorf("not implemented")
}

func (n *networkServiceWrapper) DeleteNetwork(ctx context.Context, name string) error {
	// Implementation would depend on the network store interface
	return fmt.Errorf("not implemented")
}

func (n *networkServiceWrapper) ListNetworks(ctx context.Context) ([]string, error) {
	// Implementation would depend on the network store interface
	return nil, fmt.Errorf("not implemented")
}

func (n *networkServiceWrapper) AssignJobToNetwork(ctx context.Context, jobID, networkName string) error {
	// Implementation would depend on the network store interface
	return fmt.Errorf("not implemented")
}

type monitoringServiceWrapper struct {
	service *monitoring.Service
}

func (m *monitoringServiceWrapper) CollectSystemMetrics(ctx context.Context) (map[string]interface{}, error) {
	// Convert monitoring service response to generic map
	return map[string]interface{}{
		"placeholder": true,
	}, nil
}

func (m *monitoringServiceWrapper) GetJobMetrics(ctx context.Context, jobID string) (map[string]interface{}, error) {
	// Convert monitoring service response to generic map
	return map[string]interface{}{
		"jobId": jobID,
	}, nil
}

func (m *monitoringServiceWrapper) StartMonitoring(ctx context.Context, jobID string) error {
	// Implementation would depend on the monitoring service interface
	return nil
}

func (m *monitoringServiceWrapper) StopMonitoring(ctx context.Context, jobID string) error {
	// Implementation would depend on the monitoring service interface
	return nil
}

type runtimeServiceWrapper struct {
	resolver *runtime.Resolver
}

func (r *runtimeServiceWrapper) ListRuntimes(ctx context.Context) ([]string, error) {
	runtimes, err := r.resolver.ListRuntimes()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(runtimes))
	for i, rt := range runtimes {
		names[i] = rt.Name
	}
	return names, nil
}

func (r *runtimeServiceWrapper) InstallRuntime(ctx context.Context, spec string) error {
	// Implementation would need to be added to resolver or use different interface
	return fmt.Errorf("not implemented")
}

func (r *runtimeServiceWrapper) RemoveRuntime(ctx context.Context, spec string) error {
	// Implementation would need to be added to resolver or use different interface
	return fmt.Errorf("not implemented")
}

func (r *runtimeServiceWrapper) ValidateRuntime(ctx context.Context, spec string) (bool, error) {
	_, err := r.resolver.ResolveRuntime(spec)
	return err == nil, err
}
