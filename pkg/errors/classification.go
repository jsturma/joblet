package errors

import (
	"context"
	"errors"
	"fmt"
)

// ErrorCategory helps us group errors by what kind of problem they represent.
// Makes it easier to handle different types of issues appropriately.
type ErrorCategory string

const (
	CategoryInfrastructure ErrorCategory = "infrastructure"
	CategoryConfiguration  ErrorCategory = "configuration"
	CategoryValidation     ErrorCategory = "validation"
	CategoryResource       ErrorCategory = "resource"
	CategoryRuntime        ErrorCategory = "runtime"
	CategoryNetwork        ErrorCategory = "network"
	CategoryFilesystem     ErrorCategory = "filesystem"
	CategoryPermission     ErrorCategory = "permission"
	CategoryTimeout        ErrorCategory = "timeout"
	CategoryNotFound       ErrorCategory = "not_found"
	CategoryConflict       ErrorCategory = "conflict"
	CategoryUnknown        ErrorCategory = "unknown"
)

// ErrorSeverity tells us how serious an error is - from "meh, no big deal"
// to "oh no, everything is on fire!"
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical"
	SeverityHigh     ErrorSeverity = "high"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityLow      ErrorSeverity = "low"
	SeverityInfo     ErrorSeverity = "info"
)

// ClassifiedError is like a regular error but with extra info attached.
// It tells us what kind of error it is, how serious it is, and whether we should try again.
type ClassifiedError struct {
	Err       error
	Category  ErrorCategory
	Severity  ErrorSeverity
	Retryable bool
	UserMsg   string // What we actually tell the user (without scary technical details)
}

func (e *ClassifiedError) Error() string {
	return e.Err.Error()
}

func (e *ClassifiedError) Unwrap() error {
	return e.Err
}

// ClassifyError automatically classifies an error based on its type and content
func ClassifyError(err error) *ClassifiedError {
	if err == nil {
		return nil
	}

	// Check for already classified errors
	var classified *ClassifiedError
	if errors.As(err, &classified) {
		return classified
	}

	// Classify based on error type
	switch {
	case IsJobError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryInfrastructure,
			Severity:  SeverityMedium,
			Retryable: true,
			UserMsg:   "Job operation failed. Please try again.",
		}

	case IsRuntimeError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryRuntime,
			Severity:  SeverityHigh,
			Retryable: false,
			UserMsg:   "Runtime error occurred. Please check your runtime configuration.",
		}

	case IsNetworkError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryNetwork,
			Severity:  SeverityMedium,
			Retryable: true,
			UserMsg:   "Network operation failed. Please check your network configuration.",
		}

	case IsVolumeError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryFilesystem,
			Severity:  SeverityMedium,
			Retryable: true,
			UserMsg:   "Volume operation failed. Please check your volume configuration.",
		}

	case IsFilesystemError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryFilesystem,
			Severity:  SeverityMedium,
			Retryable: false,
			UserMsg:   "Filesystem operation failed. Please check file permissions and paths.",
		}

	case IsConfigError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryConfiguration,
			Severity:  SeverityHigh,
			Retryable: false,
			UserMsg:   "Configuration error. Please check your configuration settings.",
		}

	case IsResourceError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryResource,
			Severity:  SeverityMedium,
			Retryable: true,
			UserMsg:   "Insufficient resources available. Please try again later.",
		}

	case IsTimeoutError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryTimeout,
			Severity:  SeverityMedium,
			Retryable: true,
			UserMsg:   "Operation timed out. Please try again.",
		}

	case IsNotFoundError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryNotFound,
			Severity:  SeverityLow,
			Retryable: false,
			UserMsg:   "Requested resource not found.",
		}

	case IsPermissionError(err):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryPermission,
			Severity:  SeverityHigh,
			Retryable: false,
			UserMsg:   "Permission denied. Please check your access rights.",
		}

	case errors.Is(err, context.Canceled):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryTimeout,
			Severity:  SeverityLow,
			Retryable: false,
			UserMsg:   "Operation was canceled.",
		}

	case errors.Is(err, context.DeadlineExceeded):
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryTimeout,
			Severity:  SeverityMedium,
			Retryable: true,
			UserMsg:   "Operation timed out. Please try again.",
		}

	default:
		return &ClassifiedError{
			Err:       err,
			Category:  CategoryUnknown,
			Severity:  SeverityMedium,
			Retryable: false,
			UserMsg:   "An unexpected error occurred. Please contact support.",
		}
	}
}

// ShouldRetry determines if an operation should be retried based on the error
func ShouldRetry(err error) bool {
	classified := ClassifyError(err)
	if classified == nil {
		return false
	}
	return classified.Retryable
}

// GetSeverity tells you how serious an error is.
// If we can't figure it out, we just assume it's not too bad.
func GetSeverity(err error) ErrorSeverity {
	classified := ClassifyError(err)
	if classified == nil {
		return SeverityLow
	}
	return classified.Severity
}

// GetCategory figures out what type of error we're dealing with.
// When in doubt, it just says "unknown" rather than guessing.
func GetCategory(err error) ErrorCategory {
	classified := ClassifyError(err)
	if classified == nil {
		return CategoryUnknown
	}
	return classified.Category
}

// GetUserMessage gets a nice, user-friendly message that you can actually show to people
// without making them feel like they need a computer science degree to understand it.
func GetUserMessage(err error) string {
	classified := ClassifyError(err)
	if classified == nil {
		return "An error occurred."
	}
	return classified.UserMsg
}

// IsRetryable checks if we should give this error another shot.
// Same as ShouldRetry, just a different name because some people prefer it.
func IsRetryable(err error) bool {
	return ShouldRetry(err)
}

// IsCritical checks if an error is critical severity
func IsCritical(err error) bool {
	return GetSeverity(err) == SeverityCritical
}

// ErrorSample provides metrics about error types for monitoring
type ErrorSample struct {
	Category     ErrorCategory
	Severity     ErrorSeverity
	Count        int64
	LastOccurred int64
}

// NewCriticalError creates a critical error - these are the "drop everything and fix this"
// type of errors that usually mean something is seriously broken.
func NewCriticalError(category ErrorCategory, err error, userMsg string) *ClassifiedError {
	return &ClassifiedError{
		Err:       err,
		Category:  category,
		Severity:  SeverityCritical,
		Retryable: false,
		UserMsg:   userMsg,
	}
}

// NewRetryableError creates an error that we think might work if we try again.
// Perfect for those "network hiccup" or "resource temporarily busy" situations.
func NewRetryableError(category ErrorCategory, err error, userMsg string) *ClassifiedError {
	return &ClassifiedError{
		Err:       err,
		Category:  category,
		Severity:  SeverityMedium,
		Retryable: true,
		UserMsg:   userMsg,
	}
}

// NewUserError creates a new error with a user-friendly message
func NewUserError(err error, userMsg string) *ClassifiedError {
	classified := ClassifyError(err)
	if classified == nil {
		classified = &ClassifiedError{
			Err:      err,
			Category: CategoryUnknown,
			Severity: SeverityMedium,
		}
	}
	classified.UserMsg = userMsg
	return classified
}

// FormatErrorForLogging formats an error for structured logging
func FormatErrorForLogging(err error) map[string]interface{} {
	if err == nil {
		return nil
	}

	classified := ClassifyError(err)
	result := map[string]interface{}{
		"error":     err.Error(),
		"category":  string(classified.Category),
		"severity":  string(classified.Severity),
		"retryable": classified.Retryable,
	}

	// Add specific error type information
	if jobID, ok := GetJobID(err); ok {
		result["job_id"] = jobID
	}
	if runtime, ok := GetRuntime(err); ok {
		result["runtime"] = runtime
	}
	if network, ok := GetNetwork(err); ok {
		result["network"] = network
	}

	return result
}

// LogError logs an error with appropriate context and classification
func LogError(logger interface{ Error(string, ...interface{}) }, err error, msg string) {
	if err == nil {
		return
	}

	logData := FormatErrorForLogging(err)
	args := make([]interface{}, 0, len(logData)*2)
	for k, v := range logData {
		args = append(args, k, v)
	}

	logger.Error(msg, args...)
}

// WrapWithUserMessage wraps an error with a user-friendly message while preserving the original error
func WrapWithUserMessage(err error, userMsg string) error {
	if err == nil {
		return nil
	}

	classified := NewUserError(err, userMsg)
	return fmt.Errorf("%s: %w", userMsg, classified)
}
