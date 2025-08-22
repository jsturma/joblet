package network

import (
	"fmt"
	"net"
	"sync"
)

// IPPoolManager implements the IPPool interface
type IPPoolManager struct {
	pools map[string]*networkPool
	mutex sync.RWMutex
}

// networkPool manages IP allocation for a specific network
type networkPool struct {
	cidr      *net.IPNet
	allocated map[string]bool // IP string -> allocated
	mutex     sync.RWMutex
}

// NewIPPoolManager creates a new IP pool manager
func NewIPPoolManager() *IPPoolManager {
	return &IPPoolManager{
		pools: make(map[string]*networkPool),
	}
}

// InitializePool initializes an IP pool for a network
func (ipm *IPPoolManager) InitializePool(networkName, cidr string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %s: %w", cidr, err)
	}

	ipm.mutex.Lock()
	defer ipm.mutex.Unlock()

	ipm.pools[networkName] = &networkPool{
		cidr:      ipNet,
		allocated: make(map[string]bool),
	}

	return nil
}

// AllocateIP allocates an IP from the pool
func (ipm *IPPoolManager) AllocateIP(networkName string) (net.IP, error) {
	ipm.mutex.RLock()
	pool, exists := ipm.pools[networkName]
	ipm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no IP pool for network %s", networkName)
	}

	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Find next available IP
	for ip := pool.cidr.IP.Mask(pool.cidr.Mask); pool.cidr.Contains(ip); ipm.incrementIP(ip) {
		ipStr := ip.String()

		// Skip network and broadcast addresses
		if ip.Equal(pool.cidr.IP) || ipm.isBroadcast(ip, pool.cidr) {
			continue
		}

		if !pool.allocated[ipStr] {
			pool.allocated[ipStr] = true
			// Return a copy of the IP
			allocatedIP := make(net.IP, len(ip))
			copy(allocatedIP, ip)
			return allocatedIP, nil
		}
	}

	return nil, fmt.Errorf("no available IPs in network %s", networkName)
}

// ReleaseIP releases an IP back to the pool
func (ipm *IPPoolManager) ReleaseIP(networkName string, ip net.IP) error {
	ipm.mutex.RLock()
	pool, exists := ipm.pools[networkName]
	ipm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("no IP pool for network %s", networkName)
	}

	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	ipStr := ip.String()
	if !pool.allocated[ipStr] {
		return fmt.Errorf("IP %s is not allocated in network %s", ipStr, networkName)
	}

	delete(pool.allocated, ipStr)
	return nil
}

// IsIPAvailable checks if an IP is available
func (ipm *IPPoolManager) IsIPAvailable(networkName string, ip net.IP) bool {
	ipm.mutex.RLock()
	pool, exists := ipm.pools[networkName]
	ipm.mutex.RUnlock()

	if !exists {
		return false
	}

	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	if !pool.cidr.Contains(ip) {
		return false
	}

	// Skip network and broadcast addresses
	if ip.Equal(pool.cidr.IP) || ipm.isBroadcast(ip, pool.cidr) {
		return false
	}

	return !pool.allocated[ip.String()]
}

// GetAvailableIPs returns all available IPs
func (ipm *IPPoolManager) GetAvailableIPs(networkName string) ([]net.IP, error) {
	ipm.mutex.RLock()
	pool, exists := ipm.pools[networkName]
	ipm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no IP pool for network %s", networkName)
	}

	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	var available []net.IP
	for ip := pool.cidr.IP.Mask(pool.cidr.Mask); pool.cidr.Contains(ip); ipm.incrementIP(ip) {
		// Skip network and broadcast addresses
		if ip.Equal(pool.cidr.IP) || ipm.isBroadcast(ip, pool.cidr) {
			continue
		}

		ipStr := ip.String()
		if !pool.allocated[ipStr] {
			// Add a copy of the IP
			availableIP := make(net.IP, len(ip))
			copy(availableIP, ip)
			available = append(available, availableIP)
		}
	}

	return available, nil
}

// GetAllocatedIPs returns all allocated IPs
func (ipm *IPPoolManager) GetAllocatedIPs(networkName string) ([]net.IP, error) {
	ipm.mutex.RLock()
	pool, exists := ipm.pools[networkName]
	ipm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no IP pool for network %s", networkName)
	}

	pool.mutex.RLock()
	defer pool.mutex.RUnlock()

	var allocated []net.IP
	for ipStr := range pool.allocated {
		ip := net.ParseIP(ipStr)
		if ip != nil {
			allocated = append(allocated, ip)
		}
	}

	return allocated, nil
}

// incrementIP increments an IP address
func (ipm *IPPoolManager) incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// isBroadcast checks if IP is the broadcast address for the network
func (ipm *IPPoolManager) isBroadcast(ip net.IP, ipNet *net.IPNet) bool {
	// Convert to 4-byte representation for IPv4
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}

	// Create broadcast address
	broadcast := make(net.IP, len(ip))
	copy(broadcast, ip)

	// Set host bits to 1
	for i := 0; i < len(ip); i++ {
		broadcast[i] = ip[i] | ^ipNet.Mask[i]
	}

	return ip.Equal(broadcast)
}
