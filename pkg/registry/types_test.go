package registry

import (
	"testing"
	"time"
)

func TestRegistry_GetLatestVersion(t *testing.T) {
	registry := &Registry{
		Version:   "1",
		UpdatedAt: time.Now(),
		Runtimes: map[string]map[string]*RuntimeEntry{
			"python-3.11-ml": {
				"1.0.0": {Version: "1.0.0"},
				"1.0.1": {Version: "1.0.1"},
				"1.0.2": {Version: "1.0.2"},
			},
			"openjdk-21": {
				"2.0.0": {Version: "2.0.0"},
			},
			// Test semantic version comparison: 1.10.0 > 1.9.0 (not by string comparison)
			"test-semver": {
				"1.9.0":  {Version: "1.9.0"},
				"1.10.0": {Version: "1.10.0"},
				"1.2.0":  {Version: "1.2.0"},
			},
			// Test multi-digit version numbers
			"test-complex": {
				"1.3.1":  {Version: "1.3.1"},
				"1.3.2":  {Version: "1.3.2"},
				"1.3.10": {Version: "1.3.10"},
				"1.3.9":  {Version: "1.3.9"},
			},
		},
	}

	tests := []struct {
		name        string
		runtimeName string
		want        string
	}{
		{
			name:        "python-3.11-ml latest",
			runtimeName: "python-3.11-ml",
			want:        "1.0.2", // Highest version
		},
		{
			name:        "openjdk-21 latest",
			runtimeName: "openjdk-21",
			want:        "2.0.0",
		},
		{
			name:        "semantic version comparison (1.10.0 > 1.9.0)",
			runtimeName: "test-semver",
			want:        "1.10.0", // Should be 1.10.0, not 1.9.0 (string comparison would give 1.9.0)
		},
		{
			name:        "complex version numbers",
			runtimeName: "test-complex",
			want:        "1.3.10", // Should be 1.3.10, not 1.3.9 (string comparison would give 1.3.9)
		},
		{
			name:        "non-existent runtime",
			runtimeName: "nonexistent",
			want:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.GetLatestVersion(tt.runtimeName)
			if got != tt.want {
				t.Errorf("GetLatestVersion(%q) = %q, want %q", tt.runtimeName, got, tt.want)
			}
		})
	}
}

func TestRegistry_GetRuntimeEntry(t *testing.T) {
	registry := &Registry{
		Version:   "1",
		UpdatedAt: time.Now(),
		Runtimes: map[string]map[string]*RuntimeEntry{
			"python-3.11-ml": {
				"1.0.0": {
					Version:     "1.0.0",
					DownloadURL: "https://example.com/python-3.11-ml-1.0.0.tar.gz",
					Checksum:    "sha256:abc123",
				},
			},
		},
	}

	tests := []struct {
		name        string
		runtimeName string
		version     string
		wantNil     bool
		wantURL     string
	}{
		{
			name:        "existing runtime and version",
			runtimeName: "python-3.11-ml",
			version:     "1.0.0",
			wantNil:     false,
			wantURL:     "https://example.com/python-3.11-ml-1.0.0.tar.gz",
		},
		{
			name:        "non-existent runtime",
			runtimeName: "nonexistent",
			version:     "1.0.0",
			wantNil:     true,
		},
		{
			name:        "non-existent version",
			runtimeName: "python-3.11-ml",
			version:     "999.0.0",
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := registry.GetRuntimeEntry(tt.runtimeName, tt.version)
			if tt.wantNil {
				if entry != nil {
					t.Errorf("Expected nil entry, got %+v", entry)
				}
				return
			}

			if entry == nil {
				t.Fatal("Expected non-nil entry, got nil")
			}

			if entry.DownloadURL != tt.wantURL {
				t.Errorf("Expected URL %s, got %s", tt.wantURL, entry.DownloadURL)
			}
		})
	}
}

func TestRegistry_ListVersions(t *testing.T) {
	registry := &Registry{
		Version:   "1",
		UpdatedAt: time.Now(),
		Runtimes: map[string]map[string]*RuntimeEntry{
			"python-3.11-ml": {
				"1.0.0": {Version: "1.0.0"},
				"1.0.1": {Version: "1.0.1"},
				"1.0.2": {Version: "1.0.2"},
			},
		},
	}

	tests := []struct {
		name        string
		runtimeName string
		wantCount   int
	}{
		{
			name:        "python-3.11-ml has 3 versions",
			runtimeName: "python-3.11-ml",
			wantCount:   3,
		},
		{
			name:        "non-existent runtime",
			runtimeName: "nonexistent",
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			versions := registry.ListVersions(tt.runtimeName)
			if len(versions) != tt.wantCount {
				t.Errorf("Expected %d versions, got %d", tt.wantCount, len(versions))
			}
		})
	}
}

func TestRegistry_HasRuntime(t *testing.T) {
	registry := &Registry{
		Version:   "1",
		UpdatedAt: time.Now(),
		Runtimes: map[string]map[string]*RuntimeEntry{
			"python-3.11-ml": {
				"1.0.0": {Version: "1.0.0"},
			},
		},
	}

	tests := []struct {
		name        string
		runtimeName string
		want        bool
	}{
		{
			name:        "existing runtime",
			runtimeName: "python-3.11-ml",
			want:        true,
		},
		{
			name:        "non-existent runtime",
			runtimeName: "nonexistent",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.HasRuntime(tt.runtimeName)
			if got != tt.want {
				t.Errorf("HasRuntime(%q) = %v, want %v", tt.runtimeName, got, tt.want)
			}
		})
	}
}

func TestRuntimeEntry_SupportsPlatform(t *testing.T) {
	entry := &RuntimeEntry{
		Version:   "1.0.0",
		Platforms: []string{"ubuntu-amd64", "ubuntu-arm64", "rhel-amd64"},
	}

	tests := []struct {
		name     string
		platform string
		want     bool
	}{
		{
			name:     "supported platform ubuntu-amd64",
			platform: "ubuntu-amd64",
			want:     true,
		},
		{
			name:     "supported platform ubuntu-arm64",
			platform: "ubuntu-arm64",
			want:     true,
		},
		{
			name:     "unsupported platform",
			platform: "macos-amd64",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entry.SupportsPlatform(tt.platform)
			if got != tt.want {
				t.Errorf("SupportsPlatform(%q) = %v, want %v", tt.platform, got, tt.want)
			}
		})
	}
}
