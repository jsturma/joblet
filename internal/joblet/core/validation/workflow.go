package validation

import (
	"fmt"
	"os"
	"path/filepath"

	"joblet/internal/joblet/workflow/types"
	"joblet/pkg/logger"
)

// WorkflowValidator provides comprehensive validation of workflow definitions
// before execution to prevent runtime failures and resource waste.
type WorkflowValidator struct {
	logger *logger.Logger

	// Service interfaces for validation
	volumeManager  VolumeManagerInterface
	runtimeManager RuntimeManagerInterface
}

// VolumeManagerInterface defines the interface for volume operations
type VolumeManagerInterface interface {
	VolumeExists(volumeName string) bool
}

// RuntimeManagerInterface defines the interface for runtime operations
type RuntimeManagerInterface interface {
	RuntimeExists(runtimeName string) bool
	ListRuntimes() []RuntimeInfo
}

// RuntimeInfo represents runtime information for validation
type RuntimeInfo struct {
	Name      string
	Version   string
	Available bool
}

// NewWorkflowValidator creates a new workflow validator with required dependencies
func NewWorkflowValidator(volumeManager VolumeManagerInterface, runtimeManager RuntimeManagerInterface) *WorkflowValidator {
	return &WorkflowValidator{
		logger:         logger.WithField("component", "workflow-validator"),
		volumeManager:  volumeManager,
		runtimeManager: runtimeManager,
	}
}

// ValidateWorkflow performs comprehensive pre-execution validation of a workflow
// This is the main entry point for server-side workflow validation
func (wv *WorkflowValidator) ValidateWorkflow(workflow types.WorkflowYAML) error {
	wv.logger.Info("starting comprehensive workflow validation")

	// 1. Check for circular dependencies
	if err := wv.validateNonCircularDependencies(workflow); err != nil {
		wv.logger.Error("circular dependency validation failed", "error", err)
		return fmt.Errorf("circular dependency detected: %w", err)
	}
	wv.logger.Debug("✅ No circular dependencies found")

	// 2. Validate all volumes exist
	if err := wv.validateVolumesExist(workflow); err != nil {
		wv.logger.Error("volume validation failed", "error", err)
		return fmt.Errorf("volume validation failed: %w", err)
	}
	wv.logger.Debug("✅ All required volumes exist")

	// 3. Validate all networks exist
	if err := wv.validateNetworksExist(workflow); err != nil {
		wv.logger.Error("network validation failed", "error", err)
		return fmt.Errorf("network validation failed: %w", err)
	}
	wv.logger.Debug("✅ All required networks exist")

	// 4. Validate all runtimes exist
	if err := wv.validateRuntimesExist(workflow); err != nil {
		wv.logger.Error("runtime validation failed", "error", err)
		return fmt.Errorf("runtime validation failed: %w", err)
	}
	wv.logger.Debug("✅ All required runtimes exist")

	// 5. Validate job dependencies reference existing jobs
	if err := wv.validateJobDependencies(workflow); err != nil {
		wv.logger.Error("job dependency validation failed", "error", err)
		return fmt.Errorf("job dependency validation failed: %w", err)
	}
	wv.logger.Debug("✅ All job dependencies are valid")

	wv.logger.Info("workflow validation completed successfully")
	return nil
}

// validateNonCircularDependencies checks for circular dependencies using DFS
func (wv *WorkflowValidator) validateNonCircularDependencies(workflow types.WorkflowYAML) error {
	// Build dependency graph
	graph := make(map[string][]string)
	for jobName, job := range workflow.Jobs {
		graph[jobName] = []string{}
		for _, req := range job.Requires {
			for depJob := range req {
				graph[jobName] = append(graph[jobName], depJob)
			}
		}
	}

	// Check for cycles using DFS with coloring
	color := make(map[string]int) // 0=white, 1=gray, 2=black

	var detectCycle func(string) error
	detectCycle = func(node string) error {
		if color[node] == 1 { // gray = currently being processed
			return fmt.Errorf("circular dependency involving job '%s'", node)
		}
		if color[node] == 2 { // black = already processed
			return nil
		}

		color[node] = 1 // mark as gray
		for _, neighbor := range graph[node] {
			if err := detectCycle(neighbor); err != nil {
				return err
			}
		}
		color[node] = 2 // mark as black
		return nil
	}

	for jobName := range workflow.Jobs {
		if color[jobName] == 0 {
			if err := detectCycle(jobName); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateVolumesExist checks that all referenced volumes exist on the server
func (wv *WorkflowValidator) validateVolumesExist(workflow types.WorkflowYAML) error {
	requiredVolumes := make(map[string]bool)

	// Collect all volumes referenced in jobs
	for _, job := range workflow.Jobs {
		for _, volume := range job.Volumes {
			if volume != "" {
				requiredVolumes[volume] = true
			}
		}
	}

	if len(requiredVolumes) == 0 {
		wv.logger.Debug("no volumes required by workflow")
		return nil
	}

	// Check each volume exists
	var missingVolumes []string
	for volumeName := range requiredVolumes {
		if !wv.volumeManager.VolumeExists(volumeName) {
			// Also check filesystem path as fallback
			volumePath := filepath.Join("/opt/joblet/volumes", volumeName, "data")
			if _, err := os.Stat(volumePath); os.IsNotExist(err) {
				missingVolumes = append(missingVolumes, volumeName)
				wv.logger.Warn("volume not found", "volume", volumeName)
			}
		}
	}

	if len(missingVolumes) > 0 {
		return fmt.Errorf("missing volumes: %v", missingVolumes)
	}

	return nil
}

// validateNetworksExist checks that all referenced networks exist on the server
func (wv *WorkflowValidator) validateNetworksExist(workflow types.WorkflowYAML) error {
	requiredNetworks := make(map[string]bool)

	// Collect all networks referenced in jobs
	for _, job := range workflow.Jobs {
		if job.Network != "" {
			requiredNetworks[job.Network] = true
		}
	}

	if len(requiredNetworks) == 0 {
		wv.logger.Debug("no custom networks specified in workflow")
		return nil
	}

	// Built-in networks are always available
	builtinNetworks := map[string]bool{
		"none":     true,
		"isolated": true,
		"bridge":   true,
	}

	// Check each required network
	var missingNetworks []string
	for networkName := range requiredNetworks {
		if !builtinNetworks[networkName] {
			// For now, we accept any custom network name
			// In future, this could check against a network manager interface
			// similar to how volume and runtime validation work
			wv.logger.Debug("custom network specified", "network", networkName)
		}
	}

	if len(missingNetworks) > 0 {
		return fmt.Errorf("missing networks: %v", missingNetworks)
	}

	return nil
}

// validateRuntimesExist checks that all referenced runtimes exist and are available
func (wv *WorkflowValidator) validateRuntimesExist(workflow types.WorkflowYAML) error {
	requiredRuntimes := make(map[string]bool)

	// Collect all runtimes referenced in jobs
	for _, job := range workflow.Jobs {
		if job.Runtime != "" {
			requiredRuntimes[job.Runtime] = true
		}
	}

	if len(requiredRuntimes) == 0 {
		wv.logger.Debug("no runtimes specified in workflow")
		return nil
	}

	// Get available runtimes from runtime manager
	availableRuntimes := make(map[string]bool)
	runtimes := wv.runtimeManager.ListRuntimes()
	for _, runtime := range runtimes {
		if runtime.Available {
			// Support both hyphen and colon format
			availableRuntimes[runtime.Name] = true
			// Normalize runtime name (server may store as "python-3.11-ml" but workflow uses "python:3.11-ml")
			if colonVersion := normalizeRuntimeName(runtime.Name); colonVersion != runtime.Name {
				availableRuntimes[colonVersion] = true
			}
		}
	}

	// Check each required runtime exists
	var missingRuntimes []string
	for runtimeName := range requiredRuntimes {
		// Try both original name and normalized version
		normalizedName := normalizeRuntimeName(runtimeName)

		if !availableRuntimes[runtimeName] && !availableRuntimes[normalizedName] {
			// Also try direct runtime manager check
			if !wv.runtimeManager.RuntimeExists(runtimeName) && !wv.runtimeManager.RuntimeExists(normalizedName) {
				missingRuntimes = append(missingRuntimes, runtimeName)
				wv.logger.Warn("runtime not found", "runtime", runtimeName)
			}
		}
	}

	if len(missingRuntimes) > 0 {
		return fmt.Errorf("missing runtimes: %v", missingRuntimes)
	}

	return nil
}

// validateJobDependencies checks that all job dependencies reference existing jobs
func (wv *WorkflowValidator) validateJobDependencies(workflow types.WorkflowYAML) error {
	// Get all job names
	allJobs := make(map[string]bool)
	for jobName := range workflow.Jobs {
		allJobs[jobName] = true
	}

	// Check dependencies
	for jobName, job := range workflow.Jobs {
		for _, req := range job.Requires {
			// req is map[string]string, iterate through key-value pairs
			for depJobName := range req {
				if !allJobs[depJobName] {
					wv.logger.Error("invalid job dependency", "job", jobName, "dependency", depJobName)
					return fmt.Errorf("job '%s' depends on non-existent job '%s'", jobName, depJobName)
				}
			}
		}
	}

	return nil
}

// normalizeRuntimeName converts between hyphen and colon format
// e.g., "python-3.11-ml" <-> "python:3.11-ml"
func normalizeRuntimeName(runtimeName string) string {
	// If it contains a colon, convert to hyphen format
	if len(runtimeName) > 0 {
		// Find first colon and replace with hyphen
		for i, char := range runtimeName {
			if char == ':' {
				return runtimeName[:i] + "-" + runtimeName[i+1:]
			}
		}
		// Find first hyphen and replace with colon
		for i, char := range runtimeName {
			if char == '-' {
				return runtimeName[:i] + ":" + runtimeName[i+1:]
			}
		}
	}
	return runtimeName
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	Valid                bool
	CircularDependencies bool
	VolumesValid         bool
	NetworksValid        bool
	RuntimesValid        bool
	DependenciesValid    bool
	Errors               []string
	Warnings             []string
}

// ValidateWorkflowWithSummary performs validation and returns detailed summary
func (wv *WorkflowValidator) ValidateWorkflowWithSummary(workflow types.WorkflowYAML) *ValidationSummary {
	summary := &ValidationSummary{
		Valid:                true,
		CircularDependencies: true,
		VolumesValid:         true,
		NetworksValid:        true,
		RuntimesValid:        true,
		DependenciesValid:    true,
		Errors:               []string{},
		Warnings:             []string{},
	}

	// Check circular dependencies
	if err := wv.validateNonCircularDependencies(workflow); err != nil {
		summary.Valid = false
		summary.CircularDependencies = false
		summary.Errors = append(summary.Errors, fmt.Sprintf("Circular dependencies: %v", err))
	}

	// Check volumes
	if err := wv.validateVolumesExist(workflow); err != nil {
		summary.Valid = false
		summary.VolumesValid = false
		summary.Errors = append(summary.Errors, fmt.Sprintf("Volume validation: %v", err))
	}

	// Check networks
	if err := wv.validateNetworksExist(workflow); err != nil {
		summary.Valid = false
		summary.NetworksValid = false
		summary.Errors = append(summary.Errors, fmt.Sprintf("Network validation: %v", err))
	}

	// Check runtimes
	if err := wv.validateRuntimesExist(workflow); err != nil {
		summary.Valid = false
		summary.RuntimesValid = false
		summary.Errors = append(summary.Errors, fmt.Sprintf("Runtime validation: %v", err))
	}

	// Check job dependencies
	if err := wv.validateJobDependencies(workflow); err != nil {
		summary.Valid = false
		summary.DependenciesValid = false
		summary.Errors = append(summary.Errors, fmt.Sprintf("Job dependencies: %v", err))
	}

	return summary
}
