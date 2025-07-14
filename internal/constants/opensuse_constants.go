// internal/constants/opensuse_constants.go
//go:build linux
// +build linux

package constants

func openSUSEConstants(release string) LinuxConstants {
	return LinuxConstants{
		Family:         "SUSE",
		Distribution:   "OpenSUSE",
		Release:        release,
		PackageManager: "zypper",
	}
}
