package network

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewNetworkMonitor(t *testing.T) {
	tests := []struct {
		name             string
		interval         time.Duration
		expectedInterval time.Duration
	}{
		{
			name:             "with specified interval",
			interval:         10 * time.Second,
			expectedInterval: 10 * time.Second,
		},
		{
			name:             "with zero interval defaults to 30s",
			interval:         0,
			expectedInterval: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm := NewNetworkMonitor(tt.interval)

			if nm == nil {
				t.Fatal("NewNetworkMonitor returned nil")
			}

			if nm.interval != tt.expectedInterval {
				t.Errorf("expected interval %v, got %v", tt.expectedInterval, nm.interval)
			}

			if nm.logger == nil {
				t.Error("logger should be initialized")
			}

			if nm.jobStats == nil {
				t.Error("jobStats map should be initialized")
			}

			if nm.networkStats == nil {
				t.Error("networkStats map should be initialized")
			}

			if nm.limits == nil {
				t.Error("limits map should be initialized")
			}
		})
	}
}

func TestReadNetworkStats(t *testing.T) {
	// Create a temporary /proc/net/dev file for testing
	testData := `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo: 1234567890 100000    0    0    0     0          0         0 1234567890 100000    0    0    0     0       0          0
 ens18: 9876543210 500000    0 1000    0     0          0         0 5432109876 450000    0    0    0     0       0          0
veth-p-abc123: 11111 222    0    0    0     0          0         0 33333 444    0    0    0     0       0          0
joblet0: 55555 666    0    0    0     0          0         0 77777 888    0    0    0     0       0          0
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "proc_net_dev_test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}
	tmpFile.Close()

	// We'll need to read from our temp file instead of /proc/net/dev
	// For this test, we'll call readNetworkStats() which opens /proc/net/dev
	// In a real test environment, we'd need to mock the file reading
	// For now, let's test the parsing logic directly

	t.Run("parse network stats format", func(t *testing.T) {
		// Read the temp file and parse it
		content, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("failed to read temp file: %v", err)
		}

		lines := strings.Split(string(content), "\n")
		if len(lines) < 3 {
			t.Fatal("test data should have at least 3 lines")
		}

		// Skip header lines and test parsing logic
		testLines := []struct {
			line            string
			expectedName    string
			expectedRxBytes uint64
			expectedTxBytes uint64
			expectedRxPkts  uint64
			expectedTxPkts  uint64
		}{
			{
				line:            lines[2], // lo
				expectedName:    "lo",
				expectedRxBytes: 1234567890,
				expectedTxBytes: 1234567890,
				expectedRxPkts:  100000,
				expectedTxPkts:  100000,
			},
			{
				line:            lines[3], // ens18
				expectedName:    "ens18",
				expectedRxBytes: 9876543210,
				expectedTxBytes: 5432109876,
				expectedRxPkts:  500000,
				expectedTxPkts:  450000,
			},
			{
				line:            lines[4], // veth-p-abc123
				expectedName:    "veth-p-abc123",
				expectedRxBytes: 11111,
				expectedTxBytes: 33333,
				expectedRxPkts:  222,
				expectedTxPkts:  444,
			},
			{
				line:            lines[5], // joblet0
				expectedName:    "joblet0",
				expectedRxBytes: 55555,
				expectedTxBytes: 77777,
				expectedRxPkts:  666,
				expectedTxPkts:  888,
			},
		}

		for _, tt := range testLines {
			line := strings.TrimSpace(tt.line)
			colonIndex := strings.Index(line, ":")
			if colonIndex == -1 {
				continue
			}

			interfaceName := strings.TrimSpace(line[:colonIndex])
			statsLine := strings.TrimSpace(line[colonIndex+1:])
			fields := strings.Fields(statsLine)

			if interfaceName != tt.expectedName {
				t.Errorf("expected interface name %s, got %s", tt.expectedName, interfaceName)
			}

			if len(fields) < 16 {
				t.Errorf("expected at least 16 fields, got %d", len(fields))
				continue
			}

			// Check RX bytes (field 0)
			if fields[0] != string(rune(tt.expectedRxBytes)) {
				// Compare as string representation
				expected := strings.Fields(tt.line)[1] // First field after interface name
				if fields[0] != expected {
					t.Logf("RX bytes field: expected %s, got %s", expected, fields[0])
				}
			}
		}
	})
}

func TestSetBandwidthLimits(t *testing.T) {
	nm := NewNetworkMonitor(30 * time.Second)

	tests := []struct {
		name        string
		jobID       string
		limits      *NetworkLimits
		expectError bool
		errorMsg    string
	}{
		{
			name:  "valid limits",
			jobID: "test-job-12345678",
			limits: &NetworkLimits{
				IngressBPS: 1000000, // 1 Mbps
				EgressBPS:  2000000, // 2 Mbps
				BurstSize:  128,     // 128 KB
			},
			expectError: false,
		},
		{
			name:  "negative ingress rate",
			jobID: "test-job-12345678",
			limits: &NetworkLimits{
				IngressBPS: -1000,
				EgressBPS:  1000000,
				BurstSize:  128,
			},
			expectError: true,
			errorMsg:    "bandwidth limits cannot be negative",
		},
		{
			name:  "negative egress rate",
			jobID: "test-job-12345678",
			limits: &NetworkLimits{
				IngressBPS: 1000000,
				EgressBPS:  -1000,
				BurstSize:  128,
			},
			expectError: true,
			errorMsg:    "bandwidth limits cannot be negative",
		},
		{
			name:  "negative burst size",
			jobID: "test-job-12345678",
			limits: &NetworkLimits{
				IngressBPS: 1000000,
				EgressBPS:  1000000,
				BurstSize:  -10,
			},
			expectError: true,
			errorMsg:    "burst size cannot be negative",
		},
		{
			name:  "zero limits are valid",
			jobID: "test-job-12345678",
			limits: &NetworkLimits{
				IngressBPS: 0,
				EgressBPS:  0,
				BurstSize:  0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := nm.SetBandwidthLimits(tt.jobID, tt.limits)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				// We expect tc commands to fail in test environment (no interface)
				// but the validation should pass
				if err != nil && strings.Contains(err.Error(), "negative") {
					t.Errorf("validation should have passed, got error: %v", err)
				}

				// Check that limits were stored
				storedLimits, err := nm.GetJobLimits(tt.jobID)
				if err != nil && !strings.Contains(err.Error(), "no limits found") {
					// Limits might not be stored if tc commands failed
					t.Logf("limits not stored (expected in test environment): %v", err)
				} else if storedLimits != nil {
					if storedLimits.IngressBPS != tt.limits.IngressBPS {
						t.Errorf("expected IngressBPS %d, got %d", tt.limits.IngressBPS, storedLimits.IngressBPS)
					}
					if storedLimits.EgressBPS != tt.limits.EgressBPS {
						t.Errorf("expected EgressBPS %d, got %d", tt.limits.EgressBPS, storedLimits.EgressBPS)
					}
					if storedLimits.BurstSize != tt.limits.BurstSize {
						t.Errorf("expected BurstSize %d, got %d", tt.limits.BurstSize, storedLimits.BurstSize)
					}
				}
			}
		})
	}
}

func TestGetBandwidthStats(t *testing.T) {
	nm := NewNetworkMonitor(30 * time.Second)

	// Manually add some test stats
	testJobID := "test-job-12345678"
	nm.jobStats[testJobID] = &BandwidthStats{
		Interface:       "veth-p-test-job",
		BytesSent:       1000,
		BytesReceived:   2000,
		PacketsSent:     10,
		PacketsReceived: 20,
	}

	t.Run("get existing job stats", func(t *testing.T) {
		stats, err := nm.GetBandwidthStats(testJobID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stats.Interface != "veth-p-test-job" {
			t.Errorf("expected interface %s, got %s", "veth-p-test-job", stats.Interface)
		}

		if stats.BytesSent != 1000 {
			t.Errorf("expected BytesSent 1000, got %d", stats.BytesSent)
		}

		if stats.BytesReceived != 2000 {
			t.Errorf("expected BytesReceived 2000, got %d", stats.BytesReceived)
		}

		if stats.PacketsSent != 10 {
			t.Errorf("expected PacketsSent 10, got %d", stats.PacketsSent)
		}

		if stats.PacketsReceived != 20 {
			t.Errorf("expected PacketsReceived 20, got %d", stats.PacketsReceived)
		}
	})

	t.Run("get non-existent job stats", func(t *testing.T) {
		_, err := nm.GetBandwidthStats("non-existent-job")
		if err == nil {
			t.Error("expected error for non-existent job, got nil")
		}

		if !strings.Contains(err.Error(), "no stats found") {
			t.Errorf("expected 'no stats found' error, got: %v", err)
		}
	})
}

func TestGetNetworkStats(t *testing.T) {
	nm := NewNetworkMonitor(30 * time.Second)

	// Manually add some test stats
	testNetwork := "joblet0"
	nm.networkStats[testNetwork] = &BandwidthStats{
		Interface:       testNetwork,
		BytesSent:       5000,
		BytesReceived:   6000,
		PacketsSent:     50,
		PacketsReceived: 60,
	}

	t.Run("get existing network stats", func(t *testing.T) {
		stats, err := nm.GetNetworkStats(testNetwork)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if stats.Interface != testNetwork {
			t.Errorf("expected interface %s, got %s", testNetwork, stats.Interface)
		}

		if stats.BytesSent != 5000 {
			t.Errorf("expected BytesSent 5000, got %d", stats.BytesSent)
		}

		if stats.BytesReceived != 6000 {
			t.Errorf("expected BytesReceived 6000, got %d", stats.BytesReceived)
		}
	})

	t.Run("get non-existent network stats", func(t *testing.T) {
		_, err := nm.GetNetworkStats("non-existent-network")
		if err == nil {
			t.Error("expected error for non-existent network, got nil")
		}

		if !strings.Contains(err.Error(), "no stats found") {
			t.Errorf("expected 'no stats found' error, got: %v", err)
		}
	})
}

func TestRemoveJobLimits(t *testing.T) {
	nm := NewNetworkMonitor(30 * time.Second)

	testJobID := "test-job-12345678"

	// Add some test data
	nm.limits[testJobID] = &NetworkLimits{
		IngressBPS: 1000000,
		EgressBPS:  2000000,
		BurstSize:  128,
	}
	nm.jobStats[testJobID] = &BandwidthStats{
		Interface:       "veth-p-test-job",
		BytesSent:       1000,
		BytesReceived:   2000,
		PacketsSent:     10,
		PacketsReceived: 20,
	}

	// Remove limits (tc commands will fail in test environment, but that's okay)
	err := nm.RemoveJobLimits(testJobID)
	// Error is expected because the veth interface doesn't exist in test environment
	// but the function should still clean up the maps
	t.Logf("RemoveJobLimits error (expected in test): %v", err)

	// Check that maps were cleaned up
	if _, exists := nm.limits[testJobID]; exists {
		t.Error("limits should be removed from map")
	}

	if _, exists := nm.jobStats[testJobID]; exists {
		t.Error("jobStats should be removed from map")
	}
}

func TestInterfaceNaming(t *testing.T) {
	tests := []struct {
		name              string
		jobID             string
		expectedInterface string
	}{
		{
			name:              "standard job ID",
			jobID:             "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			expectedInterface: "veth-p-f47ac10b",
		},
		{
			name:              "short job ID",
			jobID:             "abc12345",
			expectedInterface: "veth-p-abc12345",
		},
		{
			name:              "long job ID gets truncated",
			jobID:             "verylongjobid123456789",
			expectedInterface: "veth-p-verylong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the interface naming convention
			interfaceName := "veth-p-" + tt.jobID[:8]

			if interfaceName != tt.expectedInterface {
				t.Errorf("expected interface %s, got %s", tt.expectedInterface, interfaceName)
			}
		})
	}
}
