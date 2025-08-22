package core

// JobNetworkAllocation represents job network allocation for core layer
// This mirrors adapters.JobNetworkAllocation to avoid circular dependencies
type JobNetworkAllocation struct {
	JobID       string            `json:"job_id"`
	NetworkName string            `json:"network_name"`
	IPAddress   string            `json:"ip_address,omitempty"`
	MACAddress  string            `json:"mac_address,omitempty"`
	Hostname    string            `json:"hostname,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	AssignedAt  int64             `json:"assigned_at"`
}
