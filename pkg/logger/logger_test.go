package logger

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    LogLevel
		expected string
	}{
		{"DEBUG level", DEBUG, "DEBUG"},
		{"INFO level", INFO, "INFO"},
		{"WARN level", WARN, "WARN"},
		{"ERROR level", ERROR, "ERROR"},
		{"Unknown level", LogLevel(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.level.String()
			if result != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  LogLevel
		wantError bool
	}{
		{"Parse DEBUG", "DEBUG", DEBUG, false},
		{"Parse debug lowercase", "debug", DEBUG, false},
		{"Parse INFO", "INFO", INFO, false},
		{"Parse WARN", "WARN", WARN, false},
		{"Parse WARNING", "WARNING", WARN, false},
		{"Parse ERROR", "ERROR", ERROR, false},
		{"Parse invalid", "INVALID", INFO, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseLevel(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseLevel() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && result != tt.expected {
				t.Errorf("ParseLevel() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	logger := New()

	if logger == nil {
		t.Fatal("New() returned nil")
	}
	if logger.level != INFO {
		t.Errorf("Default level = %v, want %v", logger.level, INFO)
	}
	if logger.mode != "" {
		t.Errorf("Default mode = %v, want empty string", logger.mode)
	}
	if logger.fields == nil {
		t.Error("Fields map not initialized")
	}
}

func TestNewWithConfig(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:  DEBUG,
		Output: &buf,
		Format: "text",
		Mode:   "test",
	}

	logger := NewWithConfig(config)

	if logger.level != DEBUG {
		t.Errorf("Level = %v, want %v", logger.level, DEBUG)
	}
	if logger.mode != "test" {
		t.Errorf("Mode = %v, want 'test'", logger.mode)
	}
}

func TestLogger_SetAndGetMode(t *testing.T) {
	logger := New()

	logger.SetMode("server")
	if logger.GetMode() != "server" {
		t.Errorf("GetMode() = %v, want 'server'", logger.GetMode())
	}

	logger.SetMode("client")
	if logger.GetMode() != "client" {
		t.Errorf("GetMode() = %v, want 'client'", logger.GetMode())
	}
}

func TestLogger_WithFields(t *testing.T) {
	logger := New()

	// Test with multiple fields
	newLogger := logger.WithFields("key1", "value1", "key2", 123, "key3", true)

	if newLogger == logger {
		t.Error("WithFields should return new logger instance")
	}

	if len(newLogger.fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(newLogger.fields))
	}

	if newLogger.fields["key1"] != "value1" {
		t.Errorf("Field key1 = %v, want 'value1'", newLogger.fields["key1"])
	}

	// Test with odd number of arguments (should handle gracefully)
	oddLogger := logger.WithFields("key1", "value1", "key2")
	if len(oddLogger.fields) != 1 {
		t.Errorf("Expected 1 field with odd args, got %d", len(oddLogger.fields))
	}
}

func TestLogger_WithField(t *testing.T) {
	logger := New()
	newLogger := logger.WithField("request_id", "12345")

	if newLogger == logger {
		t.Error("WithField should return new logger instance")
	}

	if newLogger.fields["request_id"] != "12345" {
		t.Errorf("Field request_id = %v, want '12345'", newLogger.fields["request_id"])
	}
}

func TestLogger_WithMode(t *testing.T) {
	logger := New()
	logger = logger.WithField("existing", "field")

	newLogger := logger.WithMode("worker")

	if newLogger == logger {
		t.Error("WithMode should return new logger instance")
	}

	if newLogger.mode != "worker" {
		t.Errorf("Mode = %v, want 'worker'", newLogger.mode)
	}

	// Check that existing fields are preserved
	if newLogger.fields["existing"] != "field" {
		t.Error("WithMode should preserve existing fields")
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithConfig(Config{
		Level:  INFO,
		Output: &buf,
	})

	// DEBUG should not be logged when level is INFO
	buf.Reset()
	logger.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("DEBUG message logged when level is INFO")
	}

	// INFO should be logged
	buf.Reset()
	logger.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("INFO message not logged when level is INFO")
	}

	// WARN should be logged
	buf.Reset()
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Error("WARN message not logged when level is INFO")
	}

	// ERROR should be logged
	buf.Reset()
	logger.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Error("ERROR message not logged when level is INFO")
	}
}

func TestLogger_LogOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithConfig(Config{
		Level:  DEBUG,
		Output: &buf,
		Mode:   "test",
	})

	// Test basic logging
	logger.Info("test message")
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Error("Log output missing level")
	}
	if !strings.Contains(output, "[test]") {
		t.Error("Log output missing mode")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Log output missing message")
	}

	// Test with fields
	buf.Reset()
	logger.Info("with fields", "key1", "value1", "key2", 42)
	output = buf.String()

	if !strings.Contains(output, "key1=value1") {
		t.Error("Log output missing field key1")
	}
	if !strings.Contains(output, "key2=42") {
		t.Error("Log output missing field key2")
	}
}

func TestLogger_WithFieldsPersistence(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithConfig(Config{
		Level:  INFO,
		Output: &buf,
	})

	// Create logger with persistent fields
	contextLogger := logger.WithFields("request_id", "123", "user", "alice")

	// Log multiple messages
	buf.Reset()
	contextLogger.Info("first message")
	firstOutput := buf.String()

	buf.Reset()
	contextLogger.Info("second message")
	secondOutput := buf.String()

	// Both should contain the persistent fields
	for _, output := range []string{firstOutput, secondOutput} {
		if !strings.Contains(output, "request_id=123") {
			t.Error("Persistent field request_id missing")
		}
		if !strings.Contains(output, "user=alice") {
			t.Error("Persistent field user missing")
		}
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithConfig(Config{
		Level:  ERROR,
		Output: &buf,
	})

	// INFO should not be logged initially
	logger.Info("should not appear")
	if buf.Len() > 0 {
		t.Error("INFO logged when level is ERROR")
	}

	// Change level to INFO
	logger.SetLevel(INFO)
	buf.Reset()
	logger.Info("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Error("INFO not logged after level changed to INFO")
	}
}

func TestLogger_GetLevel(t *testing.T) {
	logger := New()

	if logger.GetLevel() != INFO {
		t.Errorf("Default level = %v, want INFO", logger.GetLevel())
	}

	logger.SetLevel(DEBUG)
	if logger.GetLevel() != DEBUG {
		t.Errorf("Level after SetLevel = %v, want DEBUG", logger.GetLevel())
	}
}

func TestLogger_IsLevelEnabled(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  LogLevel
		debugEnabled bool
		infoEnabled  bool
	}{
		{"Level DEBUG", DEBUG, true, true},
		{"Level INFO", INFO, false, true},
		{"Level WARN", WARN, false, false},
		{"Level ERROR", ERROR, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New()
			logger.SetLevel(tt.loggerLevel)

			if logger.IsDebugEnabled() != tt.debugEnabled {
				t.Errorf("IsDebugEnabled() = %v, want %v", logger.IsDebugEnabled(), tt.debugEnabled)
			}

			if logger.IsInfoEnabled() != tt.infoEnabled {
				t.Errorf("IsInfoEnabled() = %v, want %v", logger.IsInfoEnabled(), tt.infoEnabled)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"Simple string", "hello", "hello"},
		{"String with spaces", "hello world", `"hello world"`},
		{"Integer", 42, "42"},
		{"Boolean", true, "true"},
		{"Error", testError("test error"), `"test error"`},
		{"Duration", time.Second, "1s"},
		{"Nil", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test global logger functions
	SetGlobalMode("global-test")

	// These should not panic
	Debug("debug msg")
	Info("info msg")
	Warn("warn msg")
	Error("error msg")

	// Test WithFields on global logger
	loggerWithFields := WithFields("key", "value")
	if loggerWithFields == nil {
		t.Error("WithFields returned nil")
	}

	// Test WithField on global logger
	loggerWithField := WithField("id", "123")
	if loggerWithField == nil {
		t.Error("WithField returned nil")
	}

	// Test WithMode on global logger
	loggerWithMode := WithMode("test-mode")
	if loggerWithMode == nil {
		t.Error("WithMode returned nil")
	}

	// Test SetLevel on global logger
	SetLevel(DEBUG)
	// Should not panic
	Debug("debug after level change")
}

// Helper type for error testing
type testError string

func (e testError) Error() string {
	return string(e)
}

// Benchmark tests
func BenchmarkLogger_Info(b *testing.B) {
	logger := NewWithConfig(Config{
		Level:  INFO,
		Output: &bytes.Buffer{},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

func BenchmarkLogger_WithFields(b *testing.B) {
	logger := NewWithConfig(Config{
		Level:  INFO,
		Output: &bytes.Buffer{},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.WithFields("key1", "value1", "key2", "value2")
	}
}

func BenchmarkLogger_InfoWithFields(b *testing.B) {
	logger := NewWithConfig(Config{
		Level:  INFO,
		Output: &bytes.Buffer{},
	})

	contextLogger := logger.WithFields("request_id", "123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		contextLogger.Info("message", "additional", "field")
	}
}
