package installers

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pb "joblet/api/gen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_Install(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "runtime-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create local runtime for testing
	localRuntimeDir := filepath.Join(tempDir, "local-runtime")
	require.NoError(t, os.MkdirAll(localRuntimeDir, 0755))
	setupScript := filepath.Join(localRuntimeDir, "setup.sh")
	require.NoError(t, os.WriteFile(setupScript, []byte("#!/bin/bash\necho 'test'"), 0755))

	manager := NewManager(tempDir)

	tests := []struct {
		name        string
		spec        *InstallSpec
		expectError bool
	}{
		{
			name: "github source",
			spec: &InstallSpec{
				RuntimeSpec: "test-runtime",
				Source: &pb.RuntimeSourceConfig{
					SourceType: &pb.RuntimeSourceConfig_Github{
						Github: &pb.GithubSource{
							Repository: "owner/repo",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "script source",
			spec: &InstallSpec{
				RuntimeSpec: "test-runtime",
				Source: &pb.RuntimeSourceConfig{
					SourceType: &pb.RuntimeSourceConfig_Script{
						Script: &pb.ScriptSource{
							SetupScript: "echo 'test script'",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "local runtime without source",
			spec: &InstallSpec{
				RuntimeSpec: "local-runtime",
				Source:      nil,
			},
			expectError: false,
		},
		{
			name: "no source and no local runtime",
			spec: &InstallSpec{
				RuntimeSpec: "nonexistent",
				Source:      nil,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := manager.Install(context.Background(), tt.spec)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, result.Success)
				assert.Equal(t, "bash", result.Command)
			}
		})
	}
}

func TestManager_GetAvailableInstallers(t *testing.T) {
	manager := NewManager("/tmp")
	installers := manager.GetAvailableInstallers()

	// Should include all installer types
	assert.Contains(t, installers, "github")
	assert.Contains(t, installers, "script")
	assert.Contains(t, installers, "local")
	assert.Contains(t, installers, "local_runtime")
	assert.Len(t, installers, 4)
}

func TestManager_ValidateSource(t *testing.T) {
	manager := NewManager("/tmp")

	validSource := &pb.RuntimeSourceConfig{
		SourceType: &pb.RuntimeSourceConfig_Github{
			Github: &pb.GithubSource{
				Repository: "owner/repo",
			},
		},
	}

	invalidSource := &pb.RuntimeSourceConfig{
		SourceType: &pb.RuntimeSourceConfig_Github{
			Github: &pb.GithubSource{
				Repository: "invalid-format",
			},
		},
	}

	assert.NoError(t, manager.ValidateSource(validSource))
	assert.Error(t, manager.ValidateSource(invalidSource))
}
