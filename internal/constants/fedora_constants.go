// internal/constants/fedora_constants.go
//go:build linux
// +build linux

package constants

func fedoraConstants(release string) LinuxConstants {
	return LinuxConstants{
		Family:         "RedHat",
		Distribution:   "Fedora",
		Release:        release,
		PackageManager: "dnf",
	}
}
