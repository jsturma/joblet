package execution

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"joblet/internal/joblet/core/environment"
	"joblet/internal/joblet/core/upload"
	"joblet/internal/joblet/domain"
	"joblet/internal/joblet/runtime"
	"joblet/pkg/config"
	"joblet/pkg/logger"
	"joblet/pkg/platform"
)

// EnvironmentService handles job environment setup and management
type EnvironmentService struct {
	envBuilder     *environment.Builder
	uploadManager  *upload.Manager
	runtimeManager *runtime.Manager
	platform       platform.Platform
	config         *config.Config
	logger         *logger.Logger
}

// NewEnvironmentService creates a new environment service
func NewEnvironmentService(
	envBuilder *environment.Builder,
	uploadManager *upload.Manager,
	runtimeManager *runtime.Manager,
	platform platform.Platform,
	config *config.Config,
	logger *logger.Logger,
) *EnvironmentService {
	return &EnvironmentService{
		envBuilder:     envBuilder,
		uploadManager:  uploadManager,
		runtimeManager: runtimeManager,
		platform:       platform,
		config:         config,
		logger:         logger.WithField("component", "environment-service"),
	}
}

// BuildEnvironment builds the environment variables for a job
func (es *EnvironmentService) BuildEnvironment(job *domain.Job, phase string) []string {
	ctx := context.Background()
	baseEnv := es.platform.Environ()

	jobEnv := []string{
		"JOBLET_MODE=init",
		fmt.Sprintf("JOB_PHASE=%s", phase),
		fmt.Sprintf("JOB_ID=%s", job.Uuid),
		fmt.Sprintf("JOB_CGROUP_PATH=%s", "/sys/fs/cgroup"),
		fmt.Sprintf("JOB_CGROUP_HOST_PATH=%s", job.CgroupPath),
		fmt.Sprintf("JOB_MAX_CPU=%d", job.Limits.CPU.Value()),
		fmt.Sprintf("JOB_MAX_MEMORY=%d", job.Limits.Memory.Megabytes()),
		fmt.Sprintf("JOB_MAX_IOBPS=%d", job.Limits.IOBandwidth.BytesPerSecond()),
		fmt.Sprintf("JOB_COMMAND=%s", job.Command),
		fmt.Sprintf("JOB_ARGS_COUNT=%d", len(job.Args)),
	}

	// Add individual job arguments
	for i, arg := range job.Args {
		jobEnv = append(jobEnv, fmt.Sprintf("JOB_ARG_%d=%s", i, arg))
	}

	if !job.Limits.CPUCores.IsEmpty() {
		jobEnv = append(jobEnv, fmt.Sprintf("JOB_CPU_CORES=%s", job.Limits.CPUCores.String()))
	}

	// Add volume information
	if len(job.Volumes) > 0 {
		jobEnv = append(jobEnv, fmt.Sprintf("JOB_VOLUMES_COUNT=%d", len(job.Volumes)))
		for i, volume := range job.Volumes {
			jobEnv = append(jobEnv, fmt.Sprintf("JOB_VOLUME_%d=%s", i, volume))
		}
	}

	// Add runtime information
	if job.Runtime != "" {
		jobEnv = append(jobEnv, fmt.Sprintf("JOB_RUNTIME=%s", job.Runtime))

		if es.runtimeManager != nil {
			if runtimeConfig, err := es.runtimeManager.ResolveRuntimeConfig(ctx, job.Runtime); err == nil {
				jobEnv = append(jobEnv, fmt.Sprintf("RUNTIME_MANAGER_PATH=%s", es.config.Runtime.BasePath))
				runtimeEnv := es.runtimeManager.GetEnvironmentVariables(runtimeConfig)
				for key, value := range runtimeEnv {
					jobEnv = append(jobEnv, fmt.Sprintf("%s=%s", key, value))
				}
			}
		}
	}

	// Combine all environment variables
	env := append(baseEnv, jobEnv...)

	// Add regular environment variables from the job
	for key, value := range job.Environment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	// Add secret environment variables from the job
	for key, value := range job.SecretEnvironment {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}

	return env
}

// PrepareWorkspace prepares the workspace and processes uploads
func (es *EnvironmentService) PrepareWorkspace(jobID string, uploads []domain.FileUpload) (string, error) {
	// Create job base directory
	jobDir := filepath.Join(es.config.Filesystem.BaseDir, jobID)
	workDir := filepath.Join(jobDir, "work")

	if err := es.platform.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create work directory: %w", err)
	}

	// Process uploads if any
	if len(uploads) > 0 {
		if err := es.processUploads(uploads, workDir); err != nil {
			return "", fmt.Errorf("failed to process uploads: %w", err)
		}
	}

	return workDir, nil
}

// CleanupWorkspace cleans up the job workspace
func (es *EnvironmentService) CleanupWorkspace(jobID string) error {
	jobDir := filepath.Join(es.config.Filesystem.BaseDir, jobID)

	if err := os.RemoveAll(jobDir); err != nil {
		es.logger.Warn("failed to remove job directory", "jobID", jobID, "path", jobDir, "error", err)
		return err
	}

	return nil
}

// processUploads processes file uploads to the work directory
func (es *EnvironmentService) processUploads(uploads []domain.FileUpload, workDir string) error {
	es.logger.Debug("processing uploads", "count", len(uploads), "workDir", workDir)

	for _, upload := range uploads {
		fullPath := filepath.Join(workDir, upload.Path)

		if upload.IsDirectory {
			if err := os.MkdirAll(fullPath, os.FileMode(upload.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", upload.Path, err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", upload.Path, err)
			}

			if err := os.WriteFile(fullPath, upload.Content, os.FileMode(upload.Mode)); err != nil {
				return fmt.Errorf("failed to write file %s: %w", upload.Path, err)
			}
		}
	}

	return nil
}

// GetRuntimeInitPath resolves the init path for a runtime specification
func (es *EnvironmentService) GetRuntimeInitPath(ctx context.Context, runtimeSpec string) (string, error) {
	if runtimeSpec == "" {
		return "", fmt.Errorf("runtime specification is empty")
	}

	if es.runtimeManager == nil {
		return "", fmt.Errorf("runtime manager not available")
	}

	runtimeConfig, err := es.runtimeManager.ResolveRuntimeConfig(ctx, runtimeSpec)
	if err != nil {
		return "", fmt.Errorf("failed to resolve runtime config: %w", err)
	}

	if runtimeConfig == nil {
		return "", fmt.Errorf("runtime config not found")
	}

	// The runtime config should have an Init field that specifies the init binary
	if runtimeConfig.Init == "" {
		return "", fmt.Errorf("runtime config missing init path")
	}

	return runtimeConfig.Init, nil
}
