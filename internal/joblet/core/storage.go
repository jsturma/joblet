package core

import (
	"joblet/internal/joblet/adapters"
)

// Define type aliases to avoid importing concrete adapters directly
// This allows for interface-based dependency injection

// JobStore is an alias for the job storage interface
type JobStore = adapters.JobStoreAdapter

// NetworkStore interface to avoid direct adapter dependency
type NetworkStore interface {
	AllocateIP(networkName string) (string, error)
	ReleaseIP(networkName, ipAddress string) error
	AssignJobToNetwork(jobID, networkName string, allocation *JobNetworkAllocation) error
	RemoveJobFromNetwork(jobID string) error
}

// VolumeStore is an alias for the volume storage interface
type VolumeStore = adapters.VolumeStoreAdapter
