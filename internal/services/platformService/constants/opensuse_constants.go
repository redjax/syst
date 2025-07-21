//go:build linux
// +build linux

package constants

func openSUSEConstants(release string) PlatformConstants {
	return PlatformConstants{
		Family:         "SUSE",
		Distribution:   "OpenSUSE",
		Release:        release,
		PackageManager: "zypper",
	}
}
