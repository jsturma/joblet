package core

import "joblet/internal/joblet/adapters"

// NetworkStoreAdapter adapts adapters.NetworkStoreAdapter to work with core types
type NetworkStoreAdapter struct {
	store adapters.NetworkStoreAdapter
}

// NewNetworkStoreAdapter creates a bridge adapter
func NewNetworkStoreAdapter(store adapters.NetworkStoreAdapter) *NetworkStoreAdapter {
	return &NetworkStoreAdapter{store: store}
}

// AllocateIP allocates an IP address
func (nsa *NetworkStoreAdapter) AllocateIP(networkName string) (string, error) {
	return nsa.store.AllocateIP(networkName)
}

// ReleaseIP releases an IP address
func (nsa *NetworkStoreAdapter) ReleaseIP(networkName, ipAddress string) error {
	return nsa.store.ReleaseIP(networkName, ipAddress)
}

// AssignJobToNetwork assigns a job to network with type conversion
func (nsa *NetworkStoreAdapter) AssignJobToNetwork(jobID, networkName string, allocation *JobNetworkAllocation) error {
	// Convert core.JobNetworkAllocation to adapters.JobNetworkAllocation
	adapterAlloc := &adapters.JobNetworkAllocation{
		JobID:       allocation.JobID,
		NetworkName: allocation.NetworkName,
		IPAddress:   allocation.IPAddress,
		MACAddress:  allocation.MACAddress,
		Hostname:    allocation.Hostname,
		Metadata:    allocation.Metadata,
		AssignedAt:  allocation.AssignedAt,
	}

	return nsa.store.AssignJobToNetwork(jobID, networkName, adapterAlloc)
}

// RemoveJobFromNetwork removes a job from network
func (nsa *NetworkStoreAdapter) RemoveJobFromNetwork(jobID string) error {
	return nsa.store.RemoveJobFromNetwork(jobID)
}

// GetUnderlyingStore returns the underlying adapter store for components that need it
func (nsa *NetworkStoreAdapter) GetUnderlyingStore() adapters.NetworkStoreAdapter {
	return nsa.store
}
