//go:build darwin
// +build darwin

package constants

import (
	"os/exec"
	"strings"

	"github.com/redjax/syst/internal/utils"
)

// platformConstants returns macOS platform constants including detected package manager.
func platformConstants() PlatformConstants {
	return PlatformConstants{
		PackageManager: detectMacPackageManager(),
		Family:         "darwin",
		Distribution:   "macos",
		Release:        darwinRelease(),
	}
}

// detectMacPackageManager checks if Homebrew is installed; returns "brew" or "".
func detectMacPackageManager() string {
	if utils.IsCommandAvailable("brew") {
		return "brew"
	}
	return ""
}

// darwinRelease returns macOS version string, e.g. "14.4".
func darwinRelease() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}
