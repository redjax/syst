// internal/constants/ubuntu_constants.go
//go:build linux
// +build linux

package constants

func ubuntuConstants(release string) LinuxConstants {
	return LinuxConstants{
		Family:         "Debian",
		Distribution:   "Ubuntu",
		Release:        release,
		PackageManager: "apt",
	}
}
