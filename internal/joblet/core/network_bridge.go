package core

import (
	"joblet/internal/joblet/network"
)

// NetworkStoreBridge adapts NetworkStore to network.NetworkStoreInterface
type NetworkStoreBridge struct {
	store NetworkStore
}

// NewNetworkStoreBridge creates a bridge to adapt NetworkStore to NetworkStoreInterface
func NewNetworkStoreBridge(store NetworkStore) network.NetworkStoreInterface {
	return &NetworkStoreBridge{store: store}
}

// GetNetworkConfig implements network.NetworkStoreInterface
func (nsb *NetworkStoreBridge) GetNetworkConfig(name string) (*network.NetworkConfig, error) {
	// For now, return a basic config - this could be enhanced to actually
	// fetch from the store if needed
	return &network.NetworkConfig{
		CIDR:   "172.18.0.0/16", // Default CIDR
		Bridge: "br-" + name,    // Default bridge naming
	}, nil
}
