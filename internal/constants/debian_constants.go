// internal/constants/debian_constants.go
//go:build linux
// +build linux

package constants

func debianConstants(release string) LinuxConstants {
	return LinuxConstants{
		Family:         "Debian",
		Distribution:   "Debian",
		Release:        release,
		PackageManager: "apt",
	}
}
