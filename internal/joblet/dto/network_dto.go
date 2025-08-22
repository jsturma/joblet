package dto

// NetworkConfigDTO represents network configuration for data transfer
type NetworkConfigDTO struct {
	Name   string `json:"name"`
	CIDR   string `json:"cidr"`
	Bridge string `json:"bridge"`
}

// NetworkInfoDTO represents network information for listing
type NetworkInfoDTO struct {
	Name     string `json:"name"`
	CIDR     string `json:"cidr"`
	Bridge   string `json:"bridge"`
	JobCount int    `json:"job_count"`
	Status   string `json:"status,omitempty"` // active, inactive, etc.
}

// NetworkLimitsDTO defines bandwidth limits for a job
type NetworkLimitsDTO struct {
	IngressBPS int64 `json:"ingress_bps"` // Incoming bandwidth in bytes per second
	EgressBPS  int64 `json:"egress_bps"`  // Outgoing bandwidth in bytes per second
	BurstSize  int   `json:"burst_size"`  // Burst size in KB (optional)
}

// JobAllocationDTO represents a job's network allocation
type JobAllocationDTO struct {
	JobID    string `json:"job_id"`
	Network  string `json:"network"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
	VethHost string `json:"veth_host,omitempty"`
	VethPeer string `json:"veth_peer,omitempty"`
}

// BandwidthStatsDTO holds bandwidth usage statistics
type BandwidthStatsDTO struct {
	Interface       string `json:"interface"`
	BytesSent       uint64 `json:"bytes_sent"`
	BytesReceived   uint64 `json:"bytes_received"`
	PacketsSent     uint64 `json:"packets_sent"`
	PacketsReceived uint64 `json:"packets_received"`
}

// CreateNetworkRequestDTO for creating new networks
type CreateNetworkRequestDTO struct {
	Name string `json:"name"`
	CIDR string `json:"cidr"`
}

// DeleteNetworkRequestDTO for deleting networks
type DeleteNetworkRequestDTO struct {
	Name  string `json:"name"`
	Force bool   `json:"force,omitempty"` // Force delete even if jobs are using it
}
