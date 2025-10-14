package network

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ehsaniara/joblet/pkg/logger"
)

// NetworkMonitor implements the Monitor interface
type NetworkMonitor struct {
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
	mutex   sync.RWMutex
	logger  *logger.Logger

	// Stats tracking
	jobStats     map[string]*BandwidthStats
	networkStats map[string]*BandwidthStats
	limits       map[string]*NetworkLimits

	// Monitoring interval
	interval time.Duration
}

// NewNetworkMonitor creates a new network monitor
func NewNetworkMonitor(interval time.Duration) *NetworkMonitor {
	if interval == 0 {
		interval = 30 * time.Second
	}

	return &NetworkMonitor{
		logger:       logger.WithField("component", "network-monitor"),
		jobStats:     make(map[string]*BandwidthStats),
		networkStats: make(map[string]*BandwidthStats),
		limits:       make(map[string]*NetworkLimits),
		interval:     interval,
	}
}

// StartMonitoring starts the network monitoring loop
func (nm *NetworkMonitor) StartMonitoring(ctx context.Context) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.running {
		return fmt.Errorf("monitoring is already running")
	}

	nm.ctx, nm.cancel = context.WithCancel(ctx)
	nm.running = true

	go nm.monitoringLoop()
	return nil
}

// StopMonitoring stops the network monitoring
func (nm *NetworkMonitor) StopMonitoring() error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if !nm.running {
		return fmt.Errorf("monitoring is not running")
	}

	nm.cancel()
	nm.running = false
	return nil
}

// GetBandwidthStats returns bandwidth statistics for a job
func (nm *NetworkMonitor) GetBandwidthStats(jobID string) (*BandwidthStats, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	stats, exists := nm.jobStats[jobID]
	if !exists {
		return nil, fmt.Errorf("no stats found for job %s", jobID)
	}

	// Return a copy
	statsCopy := *stats
	return &statsCopy, nil
}

// GetNetworkStats returns bandwidth statistics for a network
func (nm *NetworkMonitor) GetNetworkStats(networkName string) (*BandwidthStats, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	stats, exists := nm.networkStats[networkName]
	if !exists {
		return nil, fmt.Errorf("no stats found for network %s", networkName)
	}

	// Return a copy
	statsCopy := *stats
	return &statsCopy, nil
}

// SetBandwidthLimits sets bandwidth limits for a job
func (nm *NetworkMonitor) SetBandwidthLimits(jobID string, limits *NetworkLimits) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Validate limits
	if limits.IngressBPS < 0 || limits.EgressBPS < 0 {
		return fmt.Errorf("bandwidth limits cannot be negative")
	}

	if limits.BurstSize < 0 {
		return fmt.Errorf("burst size cannot be negative")
	}

	// Store limits
	nm.limits[jobID] = &NetworkLimits{
		IngressBPS: limits.IngressBPS,
		EgressBPS:  limits.EgressBPS,
		BurstSize:  limits.BurstSize,
	}

	// Apply limits using traffic control (tc)
	return nm.applyBandwidthLimits(jobID, limits)
}

// monitoringLoop runs the main monitoring loop
func (nm *NetworkMonitor) monitoringLoop() {
	ticker := time.NewTicker(nm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-nm.ctx.Done():
			return
		case <-ticker.C:
			nm.collectStats()
		}
	}
}

// collectStats collects bandwidth statistics from network interfaces
func (nm *NetworkMonitor) collectStats() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	// Read network interface statistics from /proc/net/dev
	interfaceStats, err := nm.readNetworkStats()
	if err != nil {
		nm.logger.Warn("failed to read network stats", "error", err)
		return
	}

	// Update job stats based on veth interfaces
	for jobID := range nm.limits {
		// Job veth interfaces are named veth-p-{jobID[:8]}
		interfaceName := fmt.Sprintf("veth-p-%s", jobID[:8])

		if stats, exists := interfaceStats[interfaceName]; exists {
			if nm.jobStats[jobID] == nil {
				nm.jobStats[jobID] = &BandwidthStats{
					Interface: interfaceName,
				}
			}

			// Update job stats with current values
			nm.jobStats[jobID].BytesSent = stats.BytesSent
			nm.jobStats[jobID].BytesReceived = stats.BytesReceived
			nm.jobStats[jobID].PacketsSent = stats.PacketsSent
			nm.jobStats[jobID].PacketsReceived = stats.PacketsReceived

			nm.logger.Debug("collected job network stats",
				"jobID", jobID,
				"interface", interfaceName,
				"tx_bytes", stats.BytesSent,
				"rx_bytes", stats.BytesReceived)
		} else {
			// Interface might not exist yet or job already cleaned up
			nm.logger.Debug("job interface not found", "jobID", jobID, "interface", interfaceName)
		}
	}

	// Update network bridge stats (joblet0, joblet1, etc.)
	for interfaceName, stats := range interfaceStats {
		if strings.HasPrefix(interfaceName, "joblet") {
			// Extract network name from bridge interface (joblet0 -> network name from config)
			// For now, we'll use the interface name as the key
			if nm.networkStats[interfaceName] == nil {
				nm.networkStats[interfaceName] = &BandwidthStats{
					Interface: interfaceName,
				}
			}

			nm.networkStats[interfaceName].BytesSent = stats.BytesSent
			nm.networkStats[interfaceName].BytesReceived = stats.BytesReceived
			nm.networkStats[interfaceName].PacketsSent = stats.PacketsSent
			nm.networkStats[interfaceName].PacketsReceived = stats.PacketsReceived
		}
	}
}

// readNetworkStats reads network interface statistics from /proc/net/dev
func (nm *NetworkMonitor) readNetworkStats() (map[string]*BandwidthStats, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/net/dev: %w", err)
	}
	defer file.Close()

	stats := make(map[string]*BandwidthStats)
	scanner := bufio.NewScanner(file)

	// Skip the first two header lines
	for i := 0; i < 2 && scanner.Scan(); i++ {
		// Skip headers
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Find the colon that separates interface name from stats
		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		interfaceName := strings.TrimSpace(line[:colonIndex])
		statsLine := strings.TrimSpace(line[colonIndex+1:])

		fields := strings.Fields(statsLine)
		if len(fields) < 16 {
			nm.logger.Debug("insufficient fields in network stats",
				"interface", interfaceName,
				"fields", len(fields))
			continue
		}

		// Parse receive stats (fields 0-7)
		bytesReceived, _ := strconv.ParseUint(fields[0], 10, 64)
		packetsReceived, _ := strconv.ParseUint(fields[1], 10, 64)

		// Parse transmit stats (fields 8-15)
		bytesSent, _ := strconv.ParseUint(fields[8], 10, 64)
		packetsSent, _ := strconv.ParseUint(fields[9], 10, 64)

		stats[interfaceName] = &BandwidthStats{
			Interface:       interfaceName,
			BytesReceived:   bytesReceived,
			BytesSent:       bytesSent,
			PacketsReceived: packetsReceived,
			PacketsSent:     packetsSent,
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /proc/net/dev: %w", err)
	}

	return stats, nil
}

// applyBandwidthLimits applies traffic control limits using Linux tc (traffic control)
func (nm *NetworkMonitor) applyBandwidthLimits(jobID string, limits *NetworkLimits) error {
	// Job veth interfaces are named veth-p-{jobID[:8]}
	interfaceName := fmt.Sprintf("veth-p-%s", jobID[:8])

	nm.logger.Info("applying bandwidth limits",
		"jobID", jobID,
		"interface", interfaceName,
		"ingress_bps", limits.IngressBPS,
		"egress_bps", limits.EgressBPS,
		"burst_kb", limits.BurstSize)

	// Remove any existing qdisc first (cleanup)
	_ = nm.removeTC(interfaceName) // Ignore error - qdisc may not exist

	var lastErr error

	// Apply egress (outgoing) limit using HTB (Hierarchical Token Bucket)
	if limits.EgressBPS > 0 {
		// Add root qdisc with HTB
		cmd := exec.Command("tc", "qdisc", "add", "dev", interfaceName, "root", "handle", "1:", "htb", "default", "10")
		if output, err := cmd.CombinedOutput(); err != nil {
			lastErr = fmt.Errorf("failed to add root qdisc: %w, output: %s", err, string(output))
			nm.logger.Error("failed to add root qdisc", "error", lastErr, "output", string(output))
		} else {
			// Add HTB class with rate limit
			burstBytes := limits.BurstSize * 1024 // Convert KB to bytes
			if burstBytes == 0 {
				burstBytes = int(limits.EgressBPS / 8) // Default burst = 1 second of data
			}

			cmd = exec.Command("tc", "class", "add", "dev", interfaceName,
				"parent", "1:", "classid", "1:10", "htb",
				"rate", fmt.Sprintf("%dbps", limits.EgressBPS),
				"burst", fmt.Sprintf("%db", burstBytes))

			if output, err := cmd.CombinedOutput(); err != nil {
				lastErr = fmt.Errorf("failed to add HTB class: %w, output: %s", err, string(output))
				nm.logger.Error("failed to add HTB class", "error", lastErr, "output", string(output))
			} else {
				nm.logger.Debug("egress limit applied successfully",
					"interface", interfaceName,
					"rate_bps", limits.EgressBPS)
			}
		}
	}

	// Apply ingress (incoming) limit using police action with IFB (Intermediate Functional Block)
	// Note: Ingress limiting is more complex because tc doesn't directly support rate limiting on ingress
	// We use policing which drops packets exceeding the rate
	if limits.IngressBPS > 0 {
		// Add ingress qdisc
		cmd := exec.Command("tc", "qdisc", "add", "dev", interfaceName, "handle", "ffff:", "ingress")
		if output, err := cmd.CombinedOutput(); err != nil {
			lastErr = fmt.Errorf("failed to add ingress qdisc: %w, output: %s", err, string(output))
			nm.logger.Error("failed to add ingress qdisc", "error", lastErr, "output", string(output))
		} else {
			// Add police filter for rate limiting
			burstBytes := limits.BurstSize * 1024
			if burstBytes == 0 {
				burstBytes = int(limits.IngressBPS / 8) // Default burst = 1 second of data
			}

			cmd = exec.Command("tc", "filter", "add", "dev", interfaceName,
				"parent", "ffff:", "protocol", "ip", "prio", "1", "u32",
				"match", "u32", "0", "0",
				"police", "rate", fmt.Sprintf("%dbps", limits.IngressBPS),
				"burst", fmt.Sprintf("%db", burstBytes),
				"drop")

			if output, err := cmd.CombinedOutput(); err != nil {
				lastErr = fmt.Errorf("failed to add ingress filter: %w, output: %s", err, string(output))
				nm.logger.Error("failed to add ingress filter", "error", lastErr, "output", string(output))
			} else {
				nm.logger.Debug("ingress limit applied successfully",
					"interface", interfaceName,
					"rate_bps", limits.IngressBPS)
			}
		}
	}

	if lastErr != nil {
		return fmt.Errorf("bandwidth limit application had errors: %w", lastErr)
	}

	nm.logger.Info("bandwidth limits applied successfully",
		"jobID", jobID,
		"interface", interfaceName)

	return nil
}

// GetJobLimits returns the bandwidth limits for a job
func (nm *NetworkMonitor) GetJobLimits(jobID string) (*NetworkLimits, error) {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	limits, exists := nm.limits[jobID]
	if !exists {
		return nil, fmt.Errorf("no limits found for job %s", jobID)
	}

	// Return a copy
	limitsCopy := *limits
	return &limitsCopy, nil
}

// RemoveJobLimits removes bandwidth limits and stats for a job
func (nm *NetworkMonitor) RemoveJobLimits(jobID string) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	delete(nm.limits, jobID)
	delete(nm.jobStats, jobID)

	// Remove tc rules for the job
	return nm.removeBandwidthLimits(jobID)
}

// removeBandwidthLimits removes traffic control limits from a job's interface
func (nm *NetworkMonitor) removeBandwidthLimits(jobID string) error {
	// Job veth interfaces are named veth-p-{jobID[:8]}
	interfaceName := fmt.Sprintf("veth-p-%s", jobID[:8])

	nm.logger.Info("removing bandwidth limits", "jobID", jobID, "interface", interfaceName)

	return nm.removeTC(interfaceName)
}

// removeTC removes all traffic control rules from an interface
func (nm *NetworkMonitor) removeTC(interfaceName string) error {
	var lastErr error

	// Remove root qdisc (egress)
	cmd := exec.Command("tc", "qdisc", "del", "dev", interfaceName, "root")
	if output, err := cmd.CombinedOutput(); err != nil {
		// It's okay if this fails - interface might not exist or qdisc might not be set
		nm.logger.Debug("failed to remove root qdisc (may not exist)",
			"interface", interfaceName,
			"error", err,
			"output", string(output))
		if !strings.Contains(string(output), "Cannot find device") &&
			!strings.Contains(string(output), "No such file or directory") &&
			!strings.Contains(string(output), "RTNETLINK answers: No such device") {
			lastErr = fmt.Errorf("failed to remove root qdisc: %w", err)
		}
	} else {
		nm.logger.Debug("removed root qdisc successfully", "interface", interfaceName)
	}

	// Remove ingress qdisc
	cmd = exec.Command("tc", "qdisc", "del", "dev", interfaceName, "ingress")
	if output, err := cmd.CombinedOutput(); err != nil {
		// It's okay if this fails - interface might not exist or qdisc might not be set
		nm.logger.Debug("failed to remove ingress qdisc (may not exist)",
			"interface", interfaceName,
			"error", err,
			"output", string(output))
		if !strings.Contains(string(output), "Cannot find device") &&
			!strings.Contains(string(output), "No such file or directory") &&
			!strings.Contains(string(output), "RTNETLINK answers: No such device") &&
			!strings.Contains(string(output), "RTNETLINK answers: Invalid argument") {
			lastErr = fmt.Errorf("failed to remove ingress qdisc: %w", err)
		}
	} else {
		nm.logger.Debug("removed ingress qdisc successfully", "interface", interfaceName)
	}

	if lastErr != nil {
		nm.logger.Warn("errors occurred while removing tc rules",
			"interface", interfaceName,
			"error", lastErr)
	}

	return lastErr
}
