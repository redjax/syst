// internal/constants/linux_constants.go
//go:build linux
// +build linux

package constants

import (
	"os"
	"strings"
)

type LinuxConstants struct {
	// e.g. "Debian", "RedHat"
	Family string
	// e.g. "Ubuntu", "Fedora"
	Distribution string
	// e.g. "24.04", "42"
	Release string
	// e.g. "apt", "dnf"
	PackageManager string
}

// GetLinuxConstants detects distro and returns the appropriate constants.
func GetLinuxConstants() LinuxConstants {
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		// fallback generic Linux constants
		return LinuxConstants{
			Family:         "Unknown",
			Distribution:   "Unknown",
			Release:        "Unknown",
			PackageManager: "unknown",
		}
	}

	osRelease := string(content)

	// Detect distro by ID or ID_LIKE fields
	id := getOSReleaseField(osRelease, "ID")
	idLike := getOSReleaseField(osRelease, "ID_LIKE")
	release := getOSReleaseField(osRelease, "VERSION_ID")

	// Normalize strings to lowercase
	id = strings.ToLower(id)
	idLike = strings.ToLower(idLike)

	// Check per-distro overrides
	switch id {
	case "ubuntu":
		return ubuntuConstants(release)
	case "debian":
		return debianConstants(release)
	case "fedora":
		return fedoraConstants(release)
	case "opensuse", "opensuse-leap":
		return openSUSEConstants(release)
		// Add more distros here
	}

	// If distro not matched, check family via ID_LIKE
	if strings.Contains(idLike, "debian") {
		return debianConstants(release)
	}
	if strings.Contains(idLike, "rhel") || strings.Contains(idLike, "fedora") {
		return fedoraConstants(release)
	}

	// Default fallback
	return LinuxConstants{
		Family:         "Unknown",
		Distribution:   id,
		Release:        release,
		PackageManager: "unknown",
	}
}

// Helper to parse a field from /etc/os-release content
func getOSReleaseField(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, key+"=") {
			return strings.Trim(line[len(key)+1:], "\"")
		}
	}
	return ""
}

func platformConstants() PlatformConstants {
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return PlatformConstants{
			PackageManager: "unknown",
			Family:         "unknown",
			Distribution:   "unknown",
			Release:        "unknown",
		}
	}

	osRelease := string(content)
	id := getOSReleaseField(osRelease, "ID")
	idLike := getOSReleaseField(osRelease, "ID_LIKE")
	release := getOSReleaseField(osRelease, "VERSION_ID")

	id = strings.ToLower(id)
	idLike = strings.ToLower(idLike)

	switch id {
	case "ubuntu":
		return PlatformConstants{
			PackageManager: "apt",
			Family:         "debian",
			Distribution:   "ubuntu",
			Release:        release,
		}
	case "debian":
		return PlatformConstants{
			PackageManager: "apt",
			Family:         "debian",
			Distribution:   "debian",
			Release:        release,
		}
	case "fedora":
		return PlatformConstants{
			PackageManager: "dnf",
			Family:         "redhat",
			Distribution:   "fedora",
			Release:        release,
		}
	case "opensuse":
		return PlatformConstants{
			PackageManager: "zypper",
			Family:         "suse",
			Distribution:   "opensuse",
			Release:        release,
		}
	default:
		if strings.Contains(idLike, "debian") {
			return PlatformConstants{
				PackageManager: "apt",
				Family:         "debian",
				Distribution:   id,
				Release:        release,
			}
		}
		if strings.Contains(idLike, "rhel") || strings.Contains(idLike, "fedora") {
			return PlatformConstants{
				PackageManager: "dnf",
				Family:         "redhat",
				Distribution:   id,
				Release:        release,
			}
		}
		return PlatformConstants{
			PackageManager: "unknown",
			Family:         "unknown",
			Distribution:   id,
			Release:        release,
		}
	}
}
