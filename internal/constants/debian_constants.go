//go:build linux
// +build linux

package constants

func debianConstants(release string) PlatformConstants {
	return PlatformConstants{
		Family:         "Debian",
		Distribution:   "Debian",
		Release:        release,
		PackageManager: "apt",
	}
}
