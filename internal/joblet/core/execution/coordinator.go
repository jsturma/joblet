package execution

import (
	"context"
	"fmt"

	"joblet/pkg/logger"
	"joblet/pkg/platform"
)

// ExecutionCoordinator coordinates different execution services
// This replaces the monolithic ExecutionEngine with a focused coordinator
type ExecutionCoordinator struct {
	environmentManager EnvironmentManager
	networkManager     NetworkManager
	processManager     ProcessManager
	isolationManager   IsolationManager
	logger             *logger.Logger
}

// NewExecutionCoordinator creates a new execution coordinator
func NewExecutionCoordinator(
	envManager EnvironmentManager,
	netManager NetworkManager,
	procManager ProcessManager,
	isolManager IsolationManager,
	logger *logger.Logger,
) *ExecutionCoordinator {
	return &ExecutionCoordinator{
		environmentManager: envManager,
		networkManager:     netManager,
		processManager:     procManager,
		isolationManager:   isolManager,
		logger:             logger.WithField("component", "execution-coordinator"),
	}
}

// StartJob implements JobExecutor interface
func (ec *ExecutionCoordinator) StartJob(ctx context.Context, opts *StartProcessOptions) (platform.Command, error) {
	log := ec.logger.WithField("jobID", opts.Job.Uuid)
	log.Debug("coordinating job start", "hasUploads", len(opts.Uploads) > 0)

	// 1. Create isolation environment
	_, err := ec.isolationManager.CreateIsolatedEnvironment(opts.Job.Uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to create isolated environment: %w", err)
	}

	// 2. Setup workspace and process uploads
	workspaceDir, err := ec.environmentManager.PrepareWorkspace(opts.Job.Uuid, opts.Uploads)
	if err != nil {
		if destroyErr := ec.isolationManager.DestroyIsolatedEnvironment(opts.Job.Uuid); destroyErr != nil {
			log.Warn("failed to destroy isolated environment during cleanup", "error", destroyErr)
		}
		return nil, fmt.Errorf("failed to prepare workspace: %w", err)
	}

	// 3. Setup networking
	var networkAlloc *NetworkAllocation
	if opts.Job.Network != "" {
		networkAlloc, err = ec.networkManager.SetupNetworking(ctx, opts.Job.Uuid, opts.Job.Network)
		if err != nil {
			ec.cleanup(opts.Job.Uuid, workspaceDir)
			return nil, fmt.Errorf("failed to setup networking: %w", err)
		}
	}

	// 4. Build environment
	environment := ec.environmentManager.BuildEnvironment(opts.Job, "execute")

	// 5. Set init path to joblet binary for proper two-stage execution
	// The joblet binary runs in init mode and then exec's to the actual command
	initPath := "/opt/joblet/joblet"
	if opts.Job.Runtime != "" {
		// For jobs with runtime, use the runtime's init path
		runtimeInitPath, err := ec.environmentManager.GetRuntimeInitPath(ctx, opts.Job.Runtime)
		if err != nil {
			ec.cleanup(opts.Job.Uuid, workspaceDir)
			if networkAlloc != nil {
				if cleanupErr := ec.networkManager.CleanupNetworking(ctx, opts.Job.Uuid); cleanupErr != nil {
					log.Warn("failed to cleanup networking during runtime init path resolution failure", "error", cleanupErr)
				}
			}
			return nil, fmt.Errorf("failed to resolve runtime init path: %w", err)
		}
		initPath = runtimeInitPath
		log.Debug("using runtime init path", "initPath", initPath, "runtime", opts.Job.Runtime)
	}
	log.Debug("using joblet binary as init for namespace isolation", "initPath", initPath)

	// 6. Launch process
	launchConfig := &LaunchConfig{
		InitPath:    initPath, // Use resolved absolute path
		JobID:       opts.Job.Uuid,
		Command:     opts.Job.Command,
		Args:        opts.Job.Args,
		Environment: environment,
		// Additional config will be set based on isolation context
	}

	result, err := ec.processManager.LaunchProcess(ctx, launchConfig)
	if err != nil {
		ec.cleanup(opts.Job.Uuid, workspaceDir)
		if networkAlloc != nil {
			if cleanupErr := ec.networkManager.CleanupNetworking(ctx, opts.Job.Uuid); cleanupErr != nil {
				log.Warn("failed to cleanup networking during process launch failure", "error", cleanupErr)
			}
		}
		return nil, fmt.Errorf("failed to launch process: %w", err)
	}

	log.Info("job started successfully", "pid", result.PID)
	return result.Command, nil
}

// StopJob implements JobExecutor interface
func (ec *ExecutionCoordinator) StopJob(ctx context.Context, jobID string) error {
	log := ec.logger.WithField("jobID", jobID)
	log.Debug("coordinating job stop")

	var errs []error

	// Cleanup in reverse order
	if err := ec.networkManager.CleanupNetworking(ctx, jobID); err != nil {
		errs = append(errs, fmt.Errorf("network cleanup failed: %w", err))
	}

	if err := ec.environmentManager.CleanupWorkspace(jobID); err != nil {
		errs = append(errs, fmt.Errorf("workspace cleanup failed: %w", err))
	}

	if err := ec.isolationManager.DestroyIsolatedEnvironment(jobID); err != nil {
		errs = append(errs, fmt.Errorf("isolation cleanup failed: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	log.Info("job stopped successfully")
	return nil
}

// cleanup performs cleanup operations
func (ec *ExecutionCoordinator) cleanup(jobID, workspaceDir string) {
	if err := ec.environmentManager.CleanupWorkspace(jobID); err != nil {
		ec.logger.Warn("workspace cleanup failed during error recovery", "jobID", jobID, "error", err)
	}

	if err := ec.isolationManager.DestroyIsolatedEnvironment(jobID); err != nil {
		ec.logger.Warn("isolation cleanup failed during error recovery", "jobID", jobID, "error", err)
	}
}
