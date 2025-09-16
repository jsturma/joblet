package errors

import (
	"context"
	stderr "errors"
	"fmt"
	"testing"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name              string
		err               error
		expectedCategory  ErrorCategory
		expectedSeverity  ErrorSeverity
		expectedRetryable bool
	}{
		{
			name:              "JobError",
			err:               WrapJobError("job-123", "start", fmt.Errorf("failed")),
			expectedCategory:  CategoryInfrastructure,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: true,
		},
		{
			name:              "RuntimeError",
			err:               WrapRuntimeError("python:3.11", "install", fmt.Errorf("failed")),
			expectedCategory:  CategoryRuntime,
			expectedSeverity:  SeverityHigh,
			expectedRetryable: false,
		},
		{
			name:              "NetworkError",
			err:               WrapNetworkError("test-net", "setup", fmt.Errorf("failed")),
			expectedCategory:  CategoryNetwork,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: true,
		},
		{
			name:              "VolumeError",
			err:               WrapVolumeError("test-vol", "mount", fmt.Errorf("failed")),
			expectedCategory:  CategoryFilesystem,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: true,
		},
		{
			name:              "FilesystemError",
			err:               WrapFilesystemError("/tmp", "create", fmt.Errorf("failed")),
			expectedCategory:  CategoryFilesystem,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: false,
		},
		{
			name:              "ConfigError",
			err:               WrapConfigError("joblet", "runtime", fmt.Errorf("invalid")),
			expectedCategory:  CategoryConfiguration,
			expectedSeverity:  SeverityHigh,
			expectedRetryable: false,
		},
		{
			name:              "ResourceError",
			err:               ErrResourceExhausted,
			expectedCategory:  CategoryResource,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: true,
		},
		{
			name:              "TimeoutError",
			err:               ErrTimeout,
			expectedCategory:  CategoryTimeout,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: true,
		},
		{
			name:              "NotFoundError",
			err:               ErrJobNotFound,
			expectedCategory:  CategoryNotFound,
			expectedSeverity:  SeverityLow,
			expectedRetryable: false,
		},
		{
			name:              "PermissionError",
			err:               ErrPermissionDenied,
			expectedCategory:  CategoryPermission,
			expectedSeverity:  SeverityHigh,
			expectedRetryable: false,
		},
		{
			name:              "ContextCanceled",
			err:               context.Canceled,
			expectedCategory:  CategoryTimeout,
			expectedSeverity:  SeverityLow,
			expectedRetryable: false,
		},
		{
			name:              "ContextDeadlineExceeded",
			err:               context.DeadlineExceeded,
			expectedCategory:  CategoryTimeout,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: true,
		},
		{
			name:              "UnknownError",
			err:               fmt.Errorf("unknown error"),
			expectedCategory:  CategoryUnknown,
			expectedSeverity:  SeverityMedium,
			expectedRetryable: false,
		},
		{
			name:              "NilError",
			err:               nil,
			expectedCategory:  "",
			expectedSeverity:  "",
			expectedRetryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classified := ClassifyError(tt.err)

			if tt.err == nil {
				if classified != nil {
					t.Errorf("Expected nil for nil error, got %v", classified)
				}
				return
			}

			if classified == nil {
				t.Fatalf("Expected non-nil classification for error: %v", tt.err)
			}

			if classified.Category != tt.expectedCategory {
				t.Errorf("Expected category %v, got %v", tt.expectedCategory, classified.Category)
			}

			if classified.Severity != tt.expectedSeverity {
				t.Errorf("Expected severity %v, got %v", tt.expectedSeverity, classified.Severity)
			}

			if classified.Retryable != tt.expectedRetryable {
				t.Errorf("Expected retryable %v, got %v", tt.expectedRetryable, classified.Retryable)
			}

			// Test that the classified error still unwraps to the original
			if classified.Unwrap() != tt.err {
				t.Errorf("Expected unwrapped error to be original error")
			}

			// Test that the error message is preserved
			if classified.Error() != tt.err.Error() {
				t.Errorf("Expected error message %q, got %q", tt.err.Error(), classified.Error())
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		shouldRetry bool
	}{
		{"ResourceError", ErrResourceExhausted, true},
		{"TimeoutError", ErrTimeout, true},
		{"JobError", WrapJobError("job-123", "start", fmt.Errorf("failed")), true},
		{"NetworkError", WrapNetworkError("net", "setup", fmt.Errorf("failed")), true},
		{"PermissionError", ErrPermissionDenied, false},
		{"ConfigError", WrapConfigError("comp", "field", fmt.Errorf("invalid")), false},
		{"RuntimeError", WrapRuntimeError("python", "install", fmt.Errorf("failed")), false},
		{"UnknownError", fmt.Errorf("unknown"), false},
		{"NilError", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldRetry(tt.err)
			if result != tt.shouldRetry {
				t.Errorf("Expected ShouldRetry to return %v for %v, got %v", tt.shouldRetry, tt.err, result)
			}

			// Test IsRetryable alias
			aliasResult := IsRetryable(tt.err)
			if aliasResult != tt.shouldRetry {
				t.Errorf("Expected IsRetryable to return %v for %v, got %v", tt.shouldRetry, tt.err, aliasResult)
			}
		})
	}
}

func TestGetSeverity(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedSeverity ErrorSeverity
	}{
		{"CriticalError", NewCriticalError(CategoryInfrastructure, fmt.Errorf("critical"), "Critical error"), SeverityCritical},
		{"HighSeverityError", WrapRuntimeError("python", "install", fmt.Errorf("failed")), SeverityHigh},
		{"MediumSeverityError", WrapJobError("job-123", "start", fmt.Errorf("failed")), SeverityMedium},
		{"LowSeverityError", ErrJobNotFound, SeverityLow},
		{"NilError", nil, SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSeverity(tt.err)
			if result != tt.expectedSeverity {
				t.Errorf("Expected severity %v, got %v", tt.expectedSeverity, result)
			}
		})
	}
}

func TestGetCategory(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		expectedCategory ErrorCategory
	}{
		{"JobError", WrapJobError("job-123", "start", fmt.Errorf("failed")), CategoryInfrastructure},
		{"RuntimeError", WrapRuntimeError("python", "install", fmt.Errorf("failed")), CategoryRuntime},
		{"NetworkError", WrapNetworkError("net", "setup", fmt.Errorf("failed")), CategoryNetwork},
		{"VolumeError", WrapVolumeError("vol", "mount", fmt.Errorf("failed")), CategoryFilesystem},
		{"ConfigError", WrapConfigError("comp", "field", fmt.Errorf("invalid")), CategoryConfiguration},
		{"UnknownError", fmt.Errorf("unknown"), CategoryUnknown},
		{"NilError", nil, CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCategory(tt.err)
			if result != tt.expectedCategory {
				t.Errorf("Expected category %v, got %v", tt.expectedCategory, result)
			}
		})
	}
}

func TestGetUserMessage(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			"JobError",
			WrapJobError("job-123", "start", fmt.Errorf("failed")),
			"Job operation failed. Please try again.",
		},
		{
			"RuntimeError",
			WrapRuntimeError("python", "install", fmt.Errorf("failed")),
			"Runtime error occurred. Please check your runtime configuration.",
		},
		{
			"CustomUserMessage",
			NewUserError(fmt.Errorf("internal error"), "Custom user message"),
			"Custom user message",
		},
		{
			"NilError",
			nil,
			"An error occurred.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserMessage(tt.err)
			if result != tt.expectedMsg {
				t.Errorf("Expected user message %q, got %q", tt.expectedMsg, result)
			}
		})
	}
}

func TestIsCritical(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		isCritical bool
	}{
		{"CriticalError", NewCriticalError(CategoryInfrastructure, fmt.Errorf("critical"), "Critical"), true},
		{"NonCriticalError", WrapJobError("job-123", "start", fmt.Errorf("failed")), false},
		{"NilError", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCritical(tt.err)
			if result != tt.isCritical {
				t.Errorf("Expected IsCritical to return %v, got %v", tt.isCritical, result)
			}
		})
	}
}

func TestNewCriticalError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	criticalErr := NewCriticalError(CategoryInfrastructure, originalErr, "Critical system failure")

	if criticalErr.Category != CategoryInfrastructure {
		t.Errorf("Expected category %v, got %v", CategoryInfrastructure, criticalErr.Category)
	}

	if criticalErr.Severity != SeverityCritical {
		t.Errorf("Expected severity %v, got %v", SeverityCritical, criticalErr.Severity)
	}

	if criticalErr.Retryable {
		t.Error("Expected critical error to not be retryable")
	}

	if criticalErr.UserMsg != "Critical system failure" {
		t.Errorf("Expected user message %q, got %q", "Critical system failure", criticalErr.UserMsg)
	}

	if criticalErr.Unwrap() != originalErr {
		t.Error("Expected to unwrap to original error")
	}
}

func TestNewRetryableError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	retryableErr := NewRetryableError(CategoryResource, originalErr, "Resource temporarily unavailable")

	if retryableErr.Category != CategoryResource {
		t.Errorf("Expected category %v, got %v", CategoryResource, retryableErr.Category)
	}

	if retryableErr.Severity != SeverityMedium {
		t.Errorf("Expected severity %v, got %v", SeverityMedium, retryableErr.Severity)
	}

	if !retryableErr.Retryable {
		t.Error("Expected retryable error to be retryable")
	}

	if retryableErr.UserMsg != "Resource temporarily unavailable" {
		t.Errorf("Expected user message %q, got %q", "Resource temporarily unavailable", retryableErr.UserMsg)
	}
}

func TestFormatErrorForLogging(t *testing.T) {
	jobErr := WrapJobError("job-123", "start", fmt.Errorf("job failed"))
	runtimeErr := WrapRuntimeError("python:3.11", "install", fmt.Errorf("install failed"))
	networkErr := WrapNetworkError("test-net", "setup", fmt.Errorf("network failed"))

	tests := []struct {
		name          string
		err           error
		expectJobID   bool
		expectRuntime bool
		expectNetwork bool
	}{
		{"JobError", jobErr, true, false, false},
		{"RuntimeError", runtimeErr, false, true, false},
		{"NetworkError", networkErr, false, false, true},
		{"NilError", nil, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorForLogging(tt.err)

			if tt.err == nil {
				if result != nil {
					t.Errorf("Expected nil result for nil error, got %v", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("Expected non-nil result for error: %v", tt.err)
			}

			// Check required fields
			if _, ok := result["error"]; !ok {
				t.Error("Expected 'error' field in result")
			}
			if _, ok := result["category"]; !ok {
				t.Error("Expected 'category' field in result")
			}
			if _, ok := result["severity"]; !ok {
				t.Error("Expected 'severity' field in result")
			}
			if _, ok := result["retryable"]; !ok {
				t.Error("Expected 'retryable' field in result")
			}

			// Check optional fields
			if tt.expectJobID {
				if _, ok := result["job_id"]; !ok {
					t.Error("Expected 'job_id' field for JobError")
				}
			}
			if tt.expectRuntime {
				if _, ok := result["runtime"]; !ok {
					t.Error("Expected 'runtime' field for RuntimeError")
				}
			}
			if tt.expectNetwork {
				if _, ok := result["network"]; !ok {
					t.Error("Expected 'network' field for NetworkError")
				}
			}
		})
	}
}

func TestWrapWithUserMessage(t *testing.T) {
	originalErr := fmt.Errorf("internal database error")
	userMsg := "Unable to save your data. Please try again."

	wrappedErr := WrapWithUserMessage(originalErr, userMsg)

	if wrappedErr == nil {
		t.Fatal("Expected non-nil wrapped error")
	}

	// Should contain user message in error string
	if !contains(wrappedErr.Error(), userMsg) {
		t.Errorf("Expected wrapped error to contain user message %q, got %q", userMsg, wrappedErr.Error())
	}

	// Should be able to unwrap to a ClassifiedError
	var classified *ClassifiedError
	if !stderr.As(wrappedErr, &classified) {
		t.Error("Expected to be able to unwrap to ClassifiedError")
	}

	if classified.UserMsg != userMsg {
		t.Errorf("Expected user message %q in classified error, got %q", userMsg, classified.UserMsg)
	}

	// Test with nil error
	nilWrapped := WrapWithUserMessage(nil, "test message")
	if nilWrapped != nil {
		t.Errorf("Expected nil when wrapping nil error, got %v", nilWrapped)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

// Benchmark tests
func BenchmarkClassifyError(b *testing.B) {
	err := WrapJobError("job-123", "start", fmt.Errorf("job failed"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ClassifyError(err)
	}
}

func BenchmarkFormatErrorForLogging(b *testing.B) {
	err := WrapJobError("job-123", "start", fmt.Errorf("job failed"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FormatErrorForLogging(err)
	}
}
