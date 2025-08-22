package execution

import (
	"context"
	"fmt"
	"net"
	"time"

	"joblet/internal/joblet/network"
	"joblet/pkg/logger"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// NetworkService handles job networking setup and management
type NetworkService struct {
	networkSetup *network.NetworkSetup
	networkStore NetworkStoreInterface
	logger       *logger.Logger
}

// NetworkStoreInterface defines the contract for network storage operations
//
//counterfeiter:generate . NetworkStoreInterface
type NetworkStoreInterface interface {
	AllocateIP(networkName string) (string, error)
	ReleaseIP(networkName, ipAddress string) error
	AssignJobToNetwork(jobID, networkName string, allocation *JobNetworkAllocation) error
	RemoveJobFromNetwork(jobID string) error
}

// JobNetworkAllocation represents a job's network allocation
type JobNetworkAllocation struct {
	JobID       string
	NetworkName string
	IPAddress   string
	Hostname    string
	AssignedAt  int64
}

// NewNetworkService creates a new network service
func NewNetworkService(
	networkSetup *network.NetworkSetup,
	networkStore NetworkStoreInterface,
	logger *logger.Logger,
) *NetworkService {
	return &NetworkService{
		networkSetup: networkSetup,
		networkStore: networkStore,
		logger:       logger.WithField("component", "network-service"),
	}
}

// SetupNetworking sets up networking for a job
func (ns *NetworkService) SetupNetworking(ctx context.Context, jobID, networkName string) (*NetworkAllocation, error) {
	log := ns.logger.WithField("jobID", jobID).WithField("network", networkName)
	log.Debug("setting up job networking")

	// Handle isolated network case
	if networkName == "isolated" {
		return &NetworkAllocation{
			JobID:   jobID,
			Network: "isolated",
		}, nil
	}

	// Allocate IP from store
	ipAddress, err := ns.networkStore.AllocateIP(networkName)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}

	// Create network allocation record
	hostname := fmt.Sprintf("job_%s", jobID[:8])
	allocation := &JobNetworkAllocation{
		JobID:       jobID,
		NetworkName: networkName,
		IPAddress:   ipAddress,
		Hostname:    hostname,
		AssignedAt:  time.Now().Unix(),
	}

	// Assign job to network
	if err := ns.networkStore.AssignJobToNetwork(jobID, networkName, allocation); err != nil {
		// Release IP on error
		if releaseErr := ns.networkStore.ReleaseIP(networkName, ipAddress); releaseErr != nil {
			log.Warn("failed to release IP during cleanup", "ip", ipAddress, "error", releaseErr)
		}
		return nil, fmt.Errorf("failed to assign job to network: %w", err)
	}

	// Convert to network service allocation format
	netAlloc := &NetworkAllocation{
		JobID:    jobID,
		Network:  networkName,
		IP:       ipAddress,
		Hostname: hostname,
	}

	// Setup network in namespace using network setup service
	if ns.networkSetup != nil {
		// Convert to network.JobAllocation for network setup
		ip := net.ParseIP(ipAddress)
		if ip == nil {
			ns.cleanup(jobID, networkName, ipAddress)
			return nil, fmt.Errorf("invalid IP address: %s", ipAddress)
		}

		// Note: Network setup would happen after process launch with the actual PID
		// This is a simplified version - in practice, network setup happens after process launch
		log.Debug("network allocation prepared", "ip", ipAddress, "hostname", hostname)
	}

	log.Info("job networking setup completed", "ip", ipAddress, "hostname", hostname)
	return netAlloc, nil
}

// CleanupNetworking cleans up networking for a job
func (ns *NetworkService) CleanupNetworking(ctx context.Context, jobID string) error {
	log := ns.logger.WithField("jobID", jobID)
	log.Debug("cleaning up job networking")

	// Remove job from network store
	if err := ns.networkStore.RemoveJobFromNetwork(jobID); err != nil {
		log.Warn("failed to remove job from network", "error", err)
		return err
	}

	log.Debug("job networking cleanup completed")
	return nil
}

// cleanup performs cleanup on error
func (ns *NetworkService) cleanup(jobID, networkName, ipAddress string) {
	if err := ns.networkStore.ReleaseIP(networkName, ipAddress); err != nil {
		ns.logger.Warn("failed to release IP during cleanup",
			"jobID", jobID, "network", networkName, "ip", ipAddress, "error", err)
	}

	if err := ns.networkStore.RemoveJobFromNetwork(jobID); err != nil {
		ns.logger.Warn("failed to remove job from network during cleanup",
			"jobID", jobID, "error", err)
	}
}
