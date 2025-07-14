package constants

// PlatformConstants holds platform-specific constant values.
type PlatformConstants struct {
	PackageManager string
	Family         string
	Distribution   string
	Release        string
}

// GetPlatformConstants returns platform-specific constants.
// It calls the platform-specific implementation.
func GetPlatformConstants() PlatformConstants {
	return platformConstants()
}
