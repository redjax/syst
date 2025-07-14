//go:build linux
// +build linux

package constants

func fedoraConstants(release string) PlatformConstants {
	return PlatformConstants{
		Family:         "RedHat",
		Distribution:   "Fedora",
		Release:        release,
		PackageManager: "dnf",
	}
}
