//go:build linux
// +build linux

package constants

func ubuntuConstants(release string) PlatformConstants {
	return PlatformConstants{
		Family:         "Debian",
		Distribution:   "Ubuntu",
		Release:        release,
		PackageManager: "apt",
	}
}
