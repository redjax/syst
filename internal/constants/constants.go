package constants

import "time"

// PlatformConstants holds platform-specific constant values.
type PlatformConstants struct {
	PackageManager string
	Family         string
	Distribution   string
	Release        string

	Architecture string
	Hostname     string
	CPUModel     string
	CPUCount     int
	TotalRAM     uint64 // bytes
	DefaultShell string
	HomeDir      string
	Uptime       time.Duration
	Filesystem   string
}

// GetPlatformConstants returns platform-specific constants.
// It calls the platform-specific implementation.
func GetPlatformConstants() PlatformConstants {
	return platformConstants()
}
