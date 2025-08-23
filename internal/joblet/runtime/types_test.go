package runtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestRuntimeConfig_YAMLUnmarshaling(t *testing.T) {
	yamlContent := `name: python-3.11-ml
language: python
version: "3.11.9" 
description: "Python 3.11 with ML packages"
mounts:
  - source: "isolated/bin"
    target: "/bin"
    readonly: true
  - source: "isolated/usr"
    target: "/usr"
    readonly: false
environment:
  PATH: "/opt/venv/bin:/usr/bin:/bin"
  PYTHONPATH: "/usr/local/lib/python3.11/site-packages"
requirements:
  architectures: ["x86_64", "amd64"]
packages:
  - numpy
  - pandas
  - scikit-learn`

	var config RuntimeConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	assert.NoError(t, err)

	// Test basic fields
	assert.Equal(t, "python-3.11-ml", config.Name)
	assert.Equal(t, "python", config.Language)
	assert.Equal(t, "3.11.9", config.Version)
	assert.Equal(t, "Python 3.11 with ML packages", config.Description)

	// Test mounts
	assert.Len(t, config.Mounts, 2)

	mount1 := config.Mounts[0]
	assert.Equal(t, "isolated/bin", mount1.Source)
	assert.Equal(t, "/bin", mount1.Target)
	assert.True(t, mount1.ReadOnly)

	mount2 := config.Mounts[1]
	assert.Equal(t, "isolated/usr", mount2.Source)
	assert.Equal(t, "/usr", mount2.Target)
	assert.False(t, mount2.ReadOnly)

	// Test environment variables
	assert.Len(t, config.Environment, 2)
	assert.Equal(t, "/opt/venv/bin:/usr/bin:/bin", config.Environment["PATH"])
	assert.Equal(t, "/usr/local/lib/python3.11/site-packages", config.Environment["PYTHONPATH"])

	// Test requirements
	assert.Equal(t, []string{"x86_64", "amd64"}, config.Requirements.Architectures)

	// Test packages
	assert.Equal(t, []string{"numpy", "pandas", "scikit-learn"}, config.Packages)
}

func TestRuntimeConfig_YAMLMarshaling(t *testing.T) {
	config := RuntimeConfig{
		Name:        "test-runtime",
		Language:    "python",
		Version:     "3.11",
		Description: "Test Python runtime",
		Mounts: []MountSpec{
			{
				Source:   "isolated/bin",
				Target:   "/bin",
				ReadOnly: true,
			},
			{
				Source:   "isolated/usr",
				Target:   "/usr",
				ReadOnly: false,
			},
		},
		Environment: map[string]string{
			"PATH":       "/usr/bin:/bin",
			"PYTHONPATH": "/usr/lib/python3.11",
		},
		Requirements: RuntimeRequirements{
			Architectures: []string{"x86_64"},
		},
		Packages: []string{"requests", "flask"},
	}

	data, err := yaml.Marshal(&config)
	assert.NoError(t, err)

	// Verify that the marshaled YAML contains expected fields
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "name: test-runtime")
	assert.Contains(t, yamlStr, "language: python")
	assert.Contains(t, yamlStr, "version: \"3.11\"")
	assert.Contains(t, yamlStr, "source: isolated/bin")
	assert.Contains(t, yamlStr, "target: /bin")
	assert.Contains(t, yamlStr, "readonly: true")
	assert.Contains(t, yamlStr, "PATH: /usr/bin:/bin")
	assert.Contains(t, yamlStr, "architectures:")
	assert.Contains(t, yamlStr, "- x86_64")
	assert.Contains(t, yamlStr, "- requests")
	assert.Contains(t, yamlStr, "- flask")
}

func TestRuntimeConfig_MinimalYAML(t *testing.T) {
	// Test with minimal required fields
	yamlContent := `name: minimal-runtime
language: test
version: "1.0"`

	var config RuntimeConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	assert.NoError(t, err)

	assert.Equal(t, "minimal-runtime", config.Name)
	assert.Equal(t, "test", config.Language)
	assert.Equal(t, "1.0", config.Version)
	assert.Empty(t, config.Description)
	assert.Empty(t, config.Mounts)
	assert.Empty(t, config.Environment)
	assert.Empty(t, config.Requirements.Architectures)
	assert.Empty(t, config.Packages)
}

func TestRuntimeConfig_InvalidYAML(t *testing.T) {
	// Test with invalid YAML structure
	invalidYaml := `name: test
invalid: [[[
language: python`

	var config RuntimeConfig
	err := yaml.Unmarshal([]byte(invalidYaml), &config)
	assert.Error(t, err)
}

func TestMountSpec_YAMLHandling(t *testing.T) {
	yamlContent := `- source: "isolated/bin"
  target: "/bin" 
  readonly: true
- source: "isolated/lib"
  target: "/lib"
  readonly: false
- source: "isolated/usr"
  target: "/usr"
  # readonly defaults to false when omitted`

	var mounts []MountSpec
	err := yaml.Unmarshal([]byte(yamlContent), &mounts)
	assert.NoError(t, err)
	assert.Len(t, mounts, 3)

	// Test first mount
	assert.Equal(t, "isolated/bin", mounts[0].Source)
	assert.Equal(t, "/bin", mounts[0].Target)
	assert.True(t, mounts[0].ReadOnly)

	// Test second mount
	assert.Equal(t, "isolated/lib", mounts[1].Source)
	assert.Equal(t, "/lib", mounts[1].Target)
	assert.False(t, mounts[1].ReadOnly)

	// Test third mount (readonly omitted, defaults to false)
	assert.Equal(t, "isolated/usr", mounts[2].Source)
	assert.Equal(t, "/usr", mounts[2].Target)
	assert.False(t, mounts[2].ReadOnly) // Should default to false
}

func TestRuntimeRequirements_YAMLHandling(t *testing.T) {
	yamlContent := `architectures: ["x86_64", "amd64", "arm64"]`

	var requirements RuntimeRequirements
	err := yaml.Unmarshal([]byte(yamlContent), &requirements)
	assert.NoError(t, err)

	expected := []string{"x86_64", "amd64", "arm64"}
	assert.Equal(t, expected, requirements.Architectures)
}

func TestRuntimeSpec_Validation(t *testing.T) {
	tests := []struct {
		name    string
		spec    RuntimeSpec
		isValid bool
	}{
		{
			name: "valid python spec",
			spec: RuntimeSpec{
				Language: "python",
				Version:  "3.11",
				Tags:     []string{"ml", "gpu"},
			},
			isValid: true,
		},
		{
			name: "valid java spec",
			spec: RuntimeSpec{
				Language: "java",
				Version:  "21",
				Tags:     []string{},
			},
			isValid: true,
		},
		{
			name: "empty language",
			spec: RuntimeSpec{
				Language: "",
				Version:  "1.0",
				Tags:     []string{},
			},
			isValid: false,
		},
		{
			name: "empty version",
			spec: RuntimeSpec{
				Language: "python",
				Version:  "",
				Tags:     []string{},
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := tt.spec.Language != "" && tt.spec.Version != ""
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

func TestRuntimeInfo_Fields(t *testing.T) {
	now := time.Now()
	info := RuntimeInfo{
		Name:        "python-3.11-ml",
		Language:    "python",
		Version:     "3.11.9",
		Description: "Python with ML packages",
		Path:        "/opt/joblet/runtimes/python-3.11-ml",
		Size:        1024 * 1024 * 500, // 500MB
		LastUsed:    &now,
		Available:   true,
	}

	// Test all fields are set correctly
	assert.Equal(t, "python-3.11-ml", info.Name)
	assert.Equal(t, "python", info.Language)
	assert.Equal(t, "3.11.9", info.Version)
	assert.Equal(t, "Python with ML packages", info.Description)
	assert.Equal(t, "/opt/joblet/runtimes/python-3.11-ml", info.Path)
	assert.Equal(t, int64(1024*1024*500), info.Size)
	assert.Equal(t, &now, info.LastUsed)
	assert.True(t, info.Available)
}

func TestRuntimeConfig_DefaultValues(t *testing.T) {
	config := RuntimeConfig{}

	// Test zero values
	assert.Empty(t, config.Name)
	assert.Empty(t, config.Language)
	assert.Empty(t, config.Version)
	assert.Empty(t, config.Description)
	assert.Nil(t, config.Mounts)
	assert.Nil(t, config.Environment)
	assert.Empty(t, config.Requirements.Architectures)
	assert.Nil(t, config.Packages)
}

func TestRuntimeConfig_PartiallyPopulated(t *testing.T) {
	yamlContent := `name: partial-runtime
language: golang
version: "1.21"
mounts:
  - source: "isolated/bin"
    target: "/bin"`

	var config RuntimeConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	assert.NoError(t, err)

	// Test populated fields
	assert.Equal(t, "partial-runtime", config.Name)
	assert.Equal(t, "golang", config.Language)
	assert.Equal(t, "1.21", config.Version)

	// Test unpopulated fields have zero values
	assert.Empty(t, config.Description)
	assert.Empty(t, config.Environment)
	assert.Empty(t, config.Requirements.Architectures)
	assert.Empty(t, config.Packages)

	// Test mounts are populated
	assert.Len(t, config.Mounts, 1)
	assert.Equal(t, "isolated/bin", config.Mounts[0].Source)
	assert.Equal(t, "/bin", config.Mounts[0].Target)
	assert.False(t, config.Mounts[0].ReadOnly) // Should default to false
}

func TestRuntimeConfig_ComplexEnvironment(t *testing.T) {
	yamlContent := `name: env-test
language: python
version: "3.11"
environment:
  PATH: "/opt/venv/bin:/usr/local/bin:/usr/bin:/bin"
  PYTHONPATH: "/usr/local/lib/python3.11/site-packages:/opt/custom/lib"
  LD_LIBRARY_PATH: "/usr/local/lib:/opt/custom/lib"
  VIRTUAL_ENV: "/opt/venv"
  PYTHONDONTWRITEBYTECODE: "1"
  PYTHONUNBUFFERED: "1"`

	var config RuntimeConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	assert.NoError(t, err)

	assert.Len(t, config.Environment, 6)
	assert.Equal(t, "/opt/venv/bin:/usr/local/bin:/usr/bin:/bin", config.Environment["PATH"])
	assert.Equal(t, "/usr/local/lib/python3.11/site-packages:/opt/custom/lib", config.Environment["PYTHONPATH"])
	assert.Equal(t, "/usr/local/lib:/opt/custom/lib", config.Environment["LD_LIBRARY_PATH"])
	assert.Equal(t, "/opt/venv", config.Environment["VIRTUAL_ENV"])
	assert.Equal(t, "1", config.Environment["PYTHONDONTWRITEBYTECODE"])
	assert.Equal(t, "1", config.Environment["PYTHONUNBUFFERED"])
}

func TestRuntimeConfig_ComplexMounts(t *testing.T) {
	yamlContent := `name: mount-test
language: test
version: "1.0"
mounts:
  - source: "isolated/bin"
    target: "/bin"
    readonly: true
  - source: "isolated/lib"
    target: "/lib" 
    readonly: true
  - source: "isolated/lib64"
    target: "/lib64"
    readonly: true
  - source: "isolated/usr/bin"
    target: "/usr/bin"
    readonly: true
  - source: "isolated/usr/lib"
    target: "/usr/lib"
    readonly: true
  - source: "isolated/opt"
    target: "/opt"
    readonly: false
  - source: "isolated/etc"
    target: "/etc"
    readonly: false`

	var config RuntimeConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	assert.NoError(t, err)

	assert.Len(t, config.Mounts, 7)

	// Test all mounts have correct structure
	expectedMounts := []struct {
		source   string
		target   string
		readonly bool
	}{
		{"isolated/bin", "/bin", true},
		{"isolated/lib", "/lib", true},
		{"isolated/lib64", "/lib64", true},
		{"isolated/usr/bin", "/usr/bin", true},
		{"isolated/usr/lib", "/usr/lib", true},
		{"isolated/opt", "/opt", false},
		{"isolated/etc", "/etc", false},
	}

	for i, expected := range expectedMounts {
		mount := config.Mounts[i]
		assert.Equal(t, expected.source, mount.Source)
		assert.Equal(t, expected.target, mount.Target)
		assert.Equal(t, expected.readonly, mount.ReadOnly)
	}
}

// Test edge cases for YAML parsing
func TestRuntimeConfig_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
	}{
		{
			name: "empty mounts array",
			yamlContent: `name: test
language: python
version: "3.11"
mounts: []`,
			expectError: false,
		},
		{
			name: "empty environment map",
			yamlContent: `name: test
language: python
version: "3.11"
environment: {}`,
			expectError: false,
		},
		{
			name: "null mounts",
			yamlContent: `name: test
language: python  
version: "3.11"
mounts: null`,
			expectError: false,
		},
		{
			name: "null environment",
			yamlContent: `name: test
language: python
version: "3.11"
environment: null`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config RuntimeConfig
			err := yaml.Unmarshal([]byte(tt.yamlContent), &config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Basic validation that required fields are present
				assert.NotEmpty(t, config.Name)
				assert.NotEmpty(t, config.Language)
				assert.NotEmpty(t, config.Version)
			}
		})
	}
}
