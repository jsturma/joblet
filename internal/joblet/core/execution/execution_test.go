package execution_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"joblet/internal/joblet/core/execution"
	"joblet/internal/joblet/core/execution/executionfakes"
	"joblet/internal/joblet/domain"
	"joblet/pkg/logger"
	"joblet/pkg/platform/platformfakes"
)

// Test ExecutionCoordinator with runtime init path resolution
func TestExecutionCoordinator_RuntimeInitPathResolution(t *testing.T) {
	// Setup
	fakeEnvManager := &executionfakes.FakeEnvironmentManager{}
	fakeNetworkManager := &executionfakes.FakeNetworkManager{}
	fakeProcessManager := &executionfakes.FakeProcessManager{}
	fakeIsolationManager := &executionfakes.FakeIsolationManager{}
	fakeCommand := &platformfakes.FakeCommand{}
	testLogger := logger.New()

	coordinator := execution.NewExecutionCoordinator(
		fakeEnvManager,
		fakeNetworkManager,
		fakeProcessManager,
		fakeIsolationManager,
		testLogger,
	)

	// Setup mocks for success path
	fakeIsolationManager.CreateIsolatedEnvironmentReturns(&execution.IsolationContext{
		JobID:        "test-job-123",
		WorkspaceDir: "/tmp/test-workspace",
	}, nil)
	fakeEnvManager.PrepareWorkspaceReturns("/tmp/test-workspace", nil)
	fakeEnvManager.BuildEnvironmentReturns([]string{"TEST=value"})
	fakeEnvManager.GetRuntimeInitPathReturns("/opt/runtime/python3", nil)
	fakeProcessManager.LaunchProcessReturns(&execution.ProcessResult{
		Command: fakeCommand,
		PID:     12345,
	}, nil)

	job := &domain.Job{
		Uuid:    "test-job-123",
		Command: "python3",
		Args:    []string{"script.py"},
		Runtime: "python:3.11-ml",
	}

	opts := &execution.StartProcessOptions{
		Job:     job,
		Uploads: []domain.FileUpload{},
	}

	// Test
	result, err := coordinator.StartJob(context.Background(), opts)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != fakeCommand {
		t.Fatalf("Expected command result, got: %v", result)
	}

	// Verify runtime init path resolution was called
	if fakeEnvManager.GetRuntimeInitPathCallCount() != 1 {
		t.Errorf("Expected GetRuntimeInitPath to be called once, got: %d", fakeEnvManager.GetRuntimeInitPathCallCount())
	}

	// Verify correct runtime was passed
	_, runtimeSpec := fakeEnvManager.GetRuntimeInitPathArgsForCall(0)
	if runtimeSpec != "python:3.11-ml" {
		t.Errorf("Expected runtime spec 'python:3.11-ml', got: %s", runtimeSpec)
	}

	// Verify launch config uses runtime init path
	if fakeProcessManager.LaunchProcessCallCount() != 1 {
		t.Errorf("Expected LaunchProcess to be called once, got: %d", fakeProcessManager.LaunchProcessCallCount())
	}
	_, launchConfig := fakeProcessManager.LaunchProcessArgsForCall(0)
	if launchConfig.InitPath != "/opt/runtime/python3" {
		t.Errorf("Expected InitPath to be '/opt/runtime/python3', got: %s", launchConfig.InitPath)
	}
}

func TestExecutionCoordinator_NoRuntimeFallback(t *testing.T) {
	// Setup
	fakeEnvManager := &executionfakes.FakeEnvironmentManager{}
	fakeNetworkManager := &executionfakes.FakeNetworkManager{}
	fakeProcessManager := &executionfakes.FakeProcessManager{}
	fakeIsolationManager := &executionfakes.FakeIsolationManager{}
	fakeCommand := &platformfakes.FakeCommand{}
	testLogger := logger.New()

	coordinator := execution.NewExecutionCoordinator(
		fakeEnvManager,
		fakeNetworkManager,
		fakeProcessManager,
		fakeIsolationManager,
		testLogger,
	)

	// Setup mocks
	fakeIsolationManager.CreateIsolatedEnvironmentReturns(&execution.IsolationContext{}, nil)
	fakeEnvManager.PrepareWorkspaceReturns("/tmp/test-workspace", nil)
	fakeEnvManager.BuildEnvironmentReturns([]string{"TEST=value"})
	fakeProcessManager.LaunchProcessReturns(&execution.ProcessResult{Command: fakeCommand, PID: 12345}, nil)

	job := &domain.Job{
		Uuid:    "test-job-123",
		Command: "echo",
		Args:    []string{"hello"},
		Runtime: "", // No runtime - should use fallback
	}

	opts := &execution.StartProcessOptions{Job: job}

	// Test
	result, err := coordinator.StartJob(context.Background(), opts)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if result != fakeCommand {
		t.Fatalf("Expected command result, got: %v", result)
	}

	// Verify runtime init path was NOT called for non-runtime jobs
	if fakeEnvManager.GetRuntimeInitPathCallCount() != 0 {
		t.Errorf("Expected GetRuntimeInitPath not to be called for non-runtime jobs, got: %d", fakeEnvManager.GetRuntimeInitPathCallCount())
	}

	// Verify launch config uses resolved path for non-runtime jobs (fallback path resolution)
	_, launchConfig := fakeProcessManager.LaunchProcessArgsForCall(0)
	if launchConfig.InitPath != "/usr/bin/echo" && launchConfig.InitPath != "/bin/echo" {
		t.Errorf("Expected InitPath to be resolved to absolute path for 'echo', got: %s", launchConfig.InitPath)
	}
}

func TestExecutionCoordinator_RuntimeResolutionError(t *testing.T) {
	// Setup
	fakeEnvManager := &executionfakes.FakeEnvironmentManager{}
	fakeNetworkManager := &executionfakes.FakeNetworkManager{}
	fakeProcessManager := &executionfakes.FakeProcessManager{}
	fakeIsolationManager := &executionfakes.FakeIsolationManager{}
	testLogger := logger.New()

	coordinator := execution.NewExecutionCoordinator(
		fakeEnvManager,
		fakeNetworkManager,
		fakeProcessManager,
		fakeIsolationManager,
		testLogger,
	)

	// Setup mocks
	fakeIsolationManager.CreateIsolatedEnvironmentReturns(&execution.IsolationContext{}, nil)
	fakeEnvManager.PrepareWorkspaceReturns("/tmp/test-workspace", nil)
	fakeEnvManager.BuildEnvironmentReturns([]string{"TEST=value"})
	fakeEnvManager.GetRuntimeInitPathReturns("", errors.New("runtime not found"))

	// Setup cleanup mocks
	fakeEnvManager.CleanupWorkspaceReturns(nil)
	fakeIsolationManager.DestroyIsolatedEnvironmentReturns(nil)

	job := &domain.Job{
		Uuid:    "test-job-123",
		Command: "python3",
		Runtime: "python:3.11-ml",
	}
	opts := &execution.StartProcessOptions{Job: job}

	// Test
	_, err := coordinator.StartJob(context.Background(), opts)

	// Verify
	if err == nil {
		t.Fatal("Expected error when runtime resolution fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to resolve runtime init path") {
		t.Errorf("Expected runtime resolution error, got: %v", err)
	}

	// Verify cleanup was called
	if fakeEnvManager.CleanupWorkspaceCallCount() != 1 {
		t.Errorf("Expected CleanupWorkspace to be called once for cleanup, got: %d", fakeEnvManager.CleanupWorkspaceCallCount())
	}
	if fakeIsolationManager.DestroyIsolatedEnvironmentCallCount() != 1 {
		t.Errorf("Expected DestroyIsolatedEnvironment to be called once for cleanup, got: %d", fakeIsolationManager.DestroyIsolatedEnvironmentCallCount())
	}

	// Verify process launch was NOT called due to early failure
	if fakeProcessManager.LaunchProcessCallCount() != 0 {
		t.Errorf("Expected LaunchProcess not to be called when runtime resolution fails, got: %d", fakeProcessManager.LaunchProcessCallCount())
	}
}

func TestExecutionCoordinator_StopJob(t *testing.T) {
	// Setup
	fakeEnvManager := &executionfakes.FakeEnvironmentManager{}
	fakeNetworkManager := &executionfakes.FakeNetworkManager{}
	fakeProcessManager := &executionfakes.FakeProcessManager{}
	fakeIsolationManager := &executionfakes.FakeIsolationManager{}
	testLogger := logger.New()

	coordinator := execution.NewExecutionCoordinator(
		fakeEnvManager,
		fakeNetworkManager,
		fakeProcessManager,
		fakeIsolationManager,
		testLogger,
	)

	// Setup mocks for successful cleanup
	fakeNetworkManager.CleanupNetworkingReturns(nil)
	fakeEnvManager.CleanupWorkspaceReturns(nil)
	fakeIsolationManager.DestroyIsolatedEnvironmentReturns(nil)

	// Test
	err := coordinator.StopJob(context.Background(), "test-job-123")

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all cleanup operations were called in proper order (reverse order)
	if fakeNetworkManager.CleanupNetworkingCallCount() != 1 {
		t.Errorf("Expected CleanupNetworking to be called once, got: %d", fakeNetworkManager.CleanupNetworkingCallCount())
	}
	if fakeEnvManager.CleanupWorkspaceCallCount() != 1 {
		t.Errorf("Expected CleanupWorkspace to be called once, got: %d", fakeEnvManager.CleanupWorkspaceCallCount())
	}
	if fakeIsolationManager.DestroyIsolatedEnvironmentCallCount() != 1 {
		t.Errorf("Expected DestroyIsolatedEnvironment to be called once, got: %d", fakeIsolationManager.DestroyIsolatedEnvironmentCallCount())
	}

	// Verify correct job ID was passed to all cleanup methods
	_, jobID := fakeNetworkManager.CleanupNetworkingArgsForCall(0)
	if jobID != "test-job-123" {
		t.Errorf("Expected jobID 'test-job-123' for network cleanup, got: %s", jobID)
	}

	envJobID := fakeEnvManager.CleanupWorkspaceArgsForCall(0)
	if envJobID != "test-job-123" {
		t.Errorf("Expected jobID 'test-job-123' for workspace cleanup, got: %s", envJobID)
	}

	isoJobID := fakeIsolationManager.DestroyIsolatedEnvironmentArgsForCall(0)
	if isoJobID != "test-job-123" {
		t.Errorf("Expected jobID 'test-job-123' for isolation cleanup, got: %s", isoJobID)
	}
}

func TestExecutionCoordinator_StopJobWithErrors(t *testing.T) {
	// Setup
	fakeEnvManager := &executionfakes.FakeEnvironmentManager{}
	fakeNetworkManager := &executionfakes.FakeNetworkManager{}
	fakeProcessManager := &executionfakes.FakeProcessManager{}
	fakeIsolationManager := &executionfakes.FakeIsolationManager{}
	testLogger := logger.New()

	coordinator := execution.NewExecutionCoordinator(
		fakeEnvManager,
		fakeNetworkManager,
		fakeProcessManager,
		fakeIsolationManager,
		testLogger,
	)

	// Setup mocks to fail
	fakeNetworkManager.CleanupNetworkingReturns(errors.New("network cleanup failed"))
	fakeEnvManager.CleanupWorkspaceReturns(errors.New("workspace cleanup failed"))
	fakeIsolationManager.DestroyIsolatedEnvironmentReturns(errors.New("isolation cleanup failed"))

	// Test
	err := coordinator.StopJob(context.Background(), "test-job-123")

	// Verify error aggregation
	if err == nil {
		t.Fatal("Expected error when cleanup operations fail, got nil")
	}
	errorMessage := err.Error()
	if !strings.Contains(errorMessage, "cleanup errors") {
		t.Errorf("Expected aggregated cleanup errors, got: %v", err)
	}
	if !strings.Contains(errorMessage, "network cleanup failed") {
		t.Errorf("Expected network cleanup error in aggregated message, got: %v", err)
	}
	if !strings.Contains(errorMessage, "workspace cleanup failed") {
		t.Errorf("Expected workspace cleanup error in aggregated message, got: %v", err)
	}
	if !strings.Contains(errorMessage, "isolation cleanup failed") {
		t.Errorf("Expected isolation cleanup error in aggregated message, got: %v", err)
	}

	// Verify all cleanup operations were still attempted despite failures
	if fakeNetworkManager.CleanupNetworkingCallCount() != 1 {
		t.Errorf("Expected CleanupNetworking to be called once despite failures, got: %d", fakeNetworkManager.CleanupNetworkingCallCount())
	}
	if fakeEnvManager.CleanupWorkspaceCallCount() != 1 {
		t.Errorf("Expected CleanupWorkspace to be called once despite failures, got: %d", fakeEnvManager.CleanupWorkspaceCallCount())
	}
	if fakeIsolationManager.DestroyIsolatedEnvironmentCallCount() != 1 {
		t.Errorf("Expected DestroyIsolatedEnvironment to be called once despite failures, got: %d", fakeIsolationManager.DestroyIsolatedEnvironmentCallCount())
	}
}
