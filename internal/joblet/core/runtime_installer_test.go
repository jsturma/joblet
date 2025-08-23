//go:build linux

package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"joblet/pkg/config"
	"joblet/pkg/logger"
	"joblet/pkg/platform"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRuntimeInstaller(t *testing.T) {
	config := &config.Config{
		Runtime: config.RuntimeConfig{
			BasePath:    "/opt/joblet/runtimes",
			CommonPaths: []string{"/usr/bin", "/bin"},
		},
	}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	assert.NotNil(t, installer)
	assert.Equal(t, config, installer.config)
	assert.Equal(t, testPlatform, installer.platform)
}

func TestRuntimeInstaller_buildPathFromConfig(t *testing.T) {
	config := &config.Config{
		Runtime: config.RuntimeConfig{
			CommonPaths: []string{"/usr/bin", "/usr/local/bin"},
		},
	}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)
	pathStr := installer.buildPathFromConfig()

	assert.True(t, strings.HasPrefix(pathStr, "PATH="))
	pathValue := strings.TrimPrefix(pathStr, "PATH=")

	// Should contain configured paths
	assert.Contains(t, pathValue, "/usr/bin")
	assert.Contains(t, pathValue, "/usr/local/bin")

	// Should contain essential system paths
	assert.Contains(t, pathValue, "/bin")
	assert.Contains(t, pathValue, "/sbin")
	assert.Contains(t, pathValue, "/usr/sbin")
}

func TestRuntimeInstaller_getRuntimePath(t *testing.T) {
	config := &config.Config{}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	tests := []struct {
		name     string
		spec     string
		expected string
	}{
		{
			name:     "simple runtime spec",
			spec:     "python-3.11",
			expected: "python-3.11",
		},
		{
			name:     "complex runtime spec",
			spec:     "openjdk-21-lts",
			expected: "openjdk-21-lts",
		},
		{
			name:     "single word spec",
			spec:     "golang",
			expected: "golang",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := installer.getRuntimePath(tt.spec)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRuntimeInstaller_autoDetectRuntimePath(t *testing.T) {
	config := &config.Config{}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	tests := []struct {
		name     string
		spec     string
		expected string
	}{
		{
			name:     "spec with colon",
			spec:     "python:3.11-ml",
			expected: "python-3.11-ml",
		},
		{
			name:     "spec without colon",
			spec:     "openjdk-21",
			expected: "openjdk-21",
		},
		{
			name:     "simple spec",
			spec:     "golang:1.21",
			expected: "golang-1.21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := installer.autoDetectRuntimePath(tt.spec)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRuntimeInstaller_writeFilesToChroot(t *testing.T) {
	config := &config.Config{}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	// Create temporary chroot directory
	tempDir := t.TempDir()
	chrootDir := filepath.Join(tempDir, "chroot")
	err := os.MkdirAll(chrootDir, 0755)
	require.NoError(t, err)

	// Test files to write
	files := []*RuntimeFile{
		{
			Path:       "setup.sh",
			Content:    []byte("#!/bin/bash\necho 'Setting up runtime'"),
			Executable: true,
		},
		{
			Path:       "config.yml",
			Content:    []byte("name: test-runtime\nversion: 1.0"),
			Executable: false,
		},
		{
			Path:       "scripts/install.sh",
			Content:    []byte("#!/bin/bash\necho 'Installing packages'"),
			Executable: true,
		},
	}

	targetDir := "/tmp/runtime-scripts"
	err = installer.writeFilesToChroot(chrootDir, targetDir, files)
	require.NoError(t, err)

	// Verify files were written correctly
	fullTargetDir := filepath.Join(chrootDir, "tmp/runtime-scripts")

	// Check setup.sh
	setupPath := filepath.Join(fullTargetDir, "setup.sh")
	setupStat, err := os.Stat(setupPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), setupStat.Mode())

	setupContent, err := os.ReadFile(setupPath)
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/bash\necho 'Setting up runtime'", string(setupContent))

	// Check config.yml
	configPath := filepath.Join(fullTargetDir, "config.yml")
	configStat, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0644), configStat.Mode())

	configContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, "name: test-runtime\nversion: 1.0", string(configContent))

	// Check scripts/install.sh (nested path)
	installPath := filepath.Join(fullTargetDir, "scripts/install.sh")
	installStat, err := os.Stat(installPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), installStat.Mode())

	installContent, err := os.ReadFile(installPath)
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/bash\necho 'Installing packages'", string(installContent))
}

func TestRuntimeInstaller_findLocalRuntime(t *testing.T) {
	config := &config.Config{}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	// Create test runtime directories
	tempDir := t.TempDir()

	// Test with JOBLET_DEV_PATH environment variable
	devPath := filepath.Join(tempDir, "dev")
	runtimesDir := filepath.Join(devPath, "runtimes")
	pythonRuntimeDir := filepath.Join(runtimesDir, "python-3.11")

	err := os.MkdirAll(pythonRuntimeDir, 0755)
	require.NoError(t, err)

	// Create setup.sh
	setupScript := filepath.Join(pythonRuntimeDir, "setup.sh")
	err = os.WriteFile(setupScript, []byte("#!/bin/bash\necho 'Python setup'"), 0755)
	require.NoError(t, err)

	// Set environment variable
	originalDevPath := os.Getenv("JOBLET_DEV_PATH")
	defer func() {
		if originalDevPath != "" {
			os.Setenv("JOBLET_DEV_PATH", originalDevPath)
		} else {
			os.Unsetenv("JOBLET_DEV_PATH")
		}
	}()
	os.Setenv("JOBLET_DEV_PATH", devPath)

	// Test finding runtime
	foundPath := installer.findLocalRuntime("python:3.11")
	assert.Equal(t, pythonRuntimeDir, foundPath)

	// Test non-existent runtime
	notFoundPath := installer.findLocalRuntime("java:21")
	assert.Empty(t, notFoundPath)
}

func TestRuntimeInstaller_makedev(t *testing.T) {
	config := &config.Config{}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	tests := []struct {
		major    uint32
		minor    uint32
		expected uint64
	}{
		{1, 3, 259}, // /dev/null
		{1, 5, 261}, // /dev/zero
		{1, 8, 264}, // /dev/random
		{1, 9, 265}, // /dev/urandom
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("major_%d_minor_%d", tt.major, tt.minor), func(t *testing.T) {
			result := installer.makedev(tt.major, tt.minor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRuntimeInstallRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request RuntimeInstallRequest
		isValid bool
	}{
		{
			name: "valid request",
			request: RuntimeInstallRequest{
				RuntimeSpec:    "python-3.11",
				Repository:     "test/repo",
				Branch:         "main",
				Path:           "runtimes/python",
				ForceReinstall: false,
			},
			isValid: true,
		},
		{
			name: "empty runtime spec",
			request: RuntimeInstallRequest{
				RuntimeSpec: "",
				Repository:  "test/repo",
			},
			isValid: false,
		},
		{
			name: "minimal valid request",
			request: RuntimeInstallRequest{
				RuntimeSpec: "golang-1.21",
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.request.RuntimeSpec != ""
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestRuntimeInstallResult_Fields(t *testing.T) {
	duration := 5 * time.Minute
	result := RuntimeInstallResult{
		RuntimeSpec: "python-3.11-ml",
		Success:     true,
		Message:     "Installation completed successfully",
		InstallPath: "/opt/joblet/runtimes/python-3.11-ml/runtime.yml",
		Duration:    duration,
		LogOutput:   "Setup completed\nRuntime installed",
	}

	// Verify all fields
	assert.Equal(t, "python-3.11-ml", result.RuntimeSpec)
	assert.True(t, result.Success)
	assert.Equal(t, "Installation completed successfully", result.Message)
	assert.Equal(t, "/opt/joblet/runtimes/python-3.11-ml/runtime.yml", result.InstallPath)
	assert.Equal(t, duration, result.Duration)
	assert.Equal(t, "Setup completed\nRuntime installed", result.LogOutput)
}

func TestRuntimeInstallFromLocalRequest_Fields(t *testing.T) {
	files := []*RuntimeFile{
		{Path: "setup.sh", Content: []byte("setup"), Executable: true},
		{Path: "config.yml", Content: []byte("config"), Executable: false},
	}

	request := RuntimeInstallFromLocalRequest{
		RuntimeSpec:    "test-runtime",
		Files:          files,
		ForceReinstall: true,
	}

	assert.Equal(t, "test-runtime", request.RuntimeSpec)
	assert.Len(t, request.Files, 2)
	assert.True(t, request.ForceReinstall)

	// Check files
	assert.Equal(t, "setup.sh", request.Files[0].Path)
	assert.Equal(t, []byte("setup"), request.Files[0].Content)
	assert.True(t, request.Files[0].Executable)

	assert.Equal(t, "config.yml", request.Files[1].Path)
	assert.Equal(t, []byte("config"), request.Files[1].Content)
	assert.False(t, request.Files[1].Executable)
}

func TestRuntimeFile_Fields(t *testing.T) {
	content := []byte("#!/bin/bash\necho 'test'")
	file := RuntimeFile{
		Path:       "scripts/setup.sh",
		Content:    content,
		Executable: true,
	}

	assert.Equal(t, "scripts/setup.sh", file.Path)
	assert.Equal(t, content, file.Content)
	assert.True(t, file.Executable)
}

// Mock implementation for testing streaming interface
type mockRuntimeStreamer struct {
	progressMessages []string
	logData          [][]byte
}

func (m *mockRuntimeStreamer) SendProgress(message string) error {
	m.progressMessages = append(m.progressMessages, message)
	return nil
}

func (m *mockRuntimeStreamer) SendLog(data []byte) error {
	m.logData = append(m.logData, data)
	return nil
}

func TestRuntimeInstaller_StreamingInterface(t *testing.T) {
	// Test that the streaming interface works correctly
	streamer := &mockRuntimeStreamer{}

	// Test sending progress
	err := streamer.SendProgress("Starting installation")
	assert.NoError(t, err)
	assert.Len(t, streamer.progressMessages, 1)
	assert.Equal(t, "Starting installation", streamer.progressMessages[0])

	// Test sending log data
	logData := []byte("Installing packages...")
	err = streamer.SendLog(logData)
	assert.NoError(t, err)
	assert.Len(t, streamer.logData, 1)
	assert.Equal(t, logData, streamer.logData[0])

	// Test multiple messages
	err = streamer.SendProgress("Configuring runtime")
	assert.NoError(t, err)
	err = streamer.SendLog([]byte("Configuration complete"))
	assert.NoError(t, err)

	assert.Len(t, streamer.progressMessages, 2)
	assert.Len(t, streamer.logData, 2)
	assert.Equal(t, "Configuring runtime", streamer.progressMessages[1])
	assert.Equal(t, []byte("Configuration complete"), streamer.logData[1])
}

// Test helper methods that can be tested in isolation
func TestRuntimeInstaller_pathHandling(t *testing.T) {
	config := &config.Config{
		Runtime: config.RuntimeConfig{
			BasePath: "/opt/joblet/runtimes",
		},
	}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	// Test getRuntimePath with various inputs
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "python-3.11"},
		{"complex", "openjdk-21-lts-alpine"},
		{"with dots", "node-18.17.0"},
		{"single word", "redis"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := installer.getRuntimePath(tt.input)
			// Should return the input as-is for direct mapping
			assert.Equal(t, tt.input, result)
		})
	}
}

func TestRuntimeInstaller_environmentVariables(t *testing.T) {
	config := &config.Config{
		Runtime: config.RuntimeConfig{
			CommonPaths: []string{"/custom/bin", "/usr/local/bin"},
		},
	}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()

	installer := NewRuntimeInstaller(config, testLogger, testPlatform)
	pathEnv := installer.buildPathFromConfig()

	// Should start with PATH=
	assert.True(t, strings.HasPrefix(pathEnv, "PATH="))

	// Should contain configured paths
	assert.Contains(t, pathEnv, "/custom/bin")
	assert.Contains(t, pathEnv, "/usr/local/bin")

	// Should contain system essentials
	assert.Contains(t, pathEnv, "/bin")
	assert.Contains(t, pathEnv, "/sbin")
	assert.Contains(t, pathEnv, "/usr/sbin")

	// Should not duplicate paths
	pathValue := strings.TrimPrefix(pathEnv, "PATH=")
	paths := strings.Split(pathValue, ":")
	uniquePaths := make(map[string]bool)
	for _, path := range paths {
		assert.False(t, uniquePaths[path], "Path %s is duplicated", path)
		uniquePaths[path] = true
	}
}

// Benchmark for performance-critical methods
func BenchmarkGetRuntimePath(b *testing.B) {
	config := &config.Config{}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()
	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		installer.getRuntimePath("python-3.11-ml-gpu")
	}
}

func BenchmarkBuildPathFromConfig(b *testing.B) {
	config := &config.Config{
		Runtime: config.RuntimeConfig{
			CommonPaths: []string{"/usr/bin", "/usr/local/bin", "/opt/bin", "/custom/bin"},
		},
	}
	testLogger := logger.New()
	testPlatform := platform.NewPlatform()
	installer := NewRuntimeInstaller(config, testLogger, testPlatform)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		installer.buildPathFromConfig()
	}
}
