package validation

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// VolumeInfo represents volume information for validation
type VolumeInfo struct {
	Name      string
	Type      string
	Size      string
	Available bool
}

// RuntimeInfo represents runtime information for validation
type RuntimeInfo struct {
	Name      string
	Version   string
	Available bool
}

// NetworkInfo represents network information for validation
type NetworkInfo struct {
	Name      string
	CIDR      string
	Available bool
}

// VolumeValidator interface for volume validation operations
//
//counterfeiter:generate . VolumeValidator
type VolumeValidator interface {
	VolumeExists(name string) bool
	ValidateVolumeAccess(name string, readOnly bool) error
	ListVolumes() []VolumeInfo
}

// RuntimeValidator interface for runtime validation operations
//
//counterfeiter:generate . RuntimeValidator
type RuntimeValidator interface {
	RuntimeExists(name string) bool
	ValidateRuntime(name string) error
	ListRuntimes() []RuntimeInfo
}

// NetworkValidator interface for network validation operations
//
//counterfeiter:generate . NetworkValidator
type NetworkValidator interface {
	NetworkExists(name string) bool
	ValidateNetworkAccess(name string) error
	ListNetworks() []NetworkInfo
}

// VolumeManagerInterface defines the interface for volume operations
//
//counterfeiter:generate . VolumeManagerInterface
type VolumeManagerInterface interface {
	VolumeExists(volumeName string) bool
}

// RuntimeManagerInterface defines the interface for runtime operations
//
//counterfeiter:generate . RuntimeManagerInterface
type RuntimeManagerInterface interface {
	RuntimeExists(runtimeName string) bool
	ListRuntimes() []RuntimeInfo
}
