package installers

import (
	"context"
	"testing"

	pb "joblet/api/gen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubInstaller_Install(t *testing.T) {
	installer := NewGitHubInstaller()

	tests := []struct {
		name        string
		spec        *InstallSpec
		expectError bool
	}{
		{
			name: "valid github source",
			spec: &InstallSpec{
				RuntimeSpec: "test-runtime",
				Source: &pb.RuntimeSourceConfig{
					SourceType: &pb.RuntimeSourceConfig_Github{
						Github: &pb.GithubSource{
							Repository: "owner/repo",
							Branch:     "main",
							Path:       "runtimes",
						},
					},
				},
				BuildArgs: map[string]string{},
			},
			expectError: false,
		},
		{
			name: "invalid source type",
			spec: &InstallSpec{
				RuntimeSpec: "test-runtime",
				Source: &pb.RuntimeSourceConfig{
					SourceType: &pb.RuntimeSourceConfig_Script{
						Script: &pb.ScriptSource{},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := installer.Install(context.Background(), tt.spec)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, result.Success)
				assert.Equal(t, "bash", result.Command)
				assert.Len(t, result.Args, 2)
				assert.Equal(t, "-c", result.Args[0])
				assert.Contains(t, result.Args[1], "Installing runtime test-runtime")
			}
		})
	}
}

func TestGitHubInstaller_Validate(t *testing.T) {
	installer := NewGitHubInstaller()

	tests := []struct {
		name        string
		source      *pb.RuntimeSourceConfig
		expectError bool
	}{
		{
			name: "valid repository format",
			source: &pb.RuntimeSourceConfig{
				SourceType: &pb.RuntimeSourceConfig_Github{
					Github: &pb.GithubSource{
						Repository: "owner/repo",
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid repository format",
			source: &pb.RuntimeSourceConfig{
				SourceType: &pb.RuntimeSourceConfig_Github{
					Github: &pb.GithubSource{
						Repository: "invalid-format",
					},
				},
			},
			expectError: true,
		},
		{
			name: "wrong source type",
			source: &pb.RuntimeSourceConfig{
				SourceType: &pb.RuntimeSourceConfig_Script{
					Script: &pb.ScriptSource{},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := installer.Validate(tt.source)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGitHubInstaller_GetSourceType(t *testing.T) {
	installer := NewGitHubInstaller()
	assert.Equal(t, "github", installer.GetSourceType())
}
