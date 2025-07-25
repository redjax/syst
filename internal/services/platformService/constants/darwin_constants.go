//go:build darwin
// +build darwin

package constants

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"syscall"

	"github.com/redjax/syst/internal/utils"
)

// platformConstants assembles mac-specific platform info
func platformConstants() PlatformConstants {
	return PlatformConstants{
		PackageManager: detectMacPackageManager(),
		Family:         "darwin",
		Distribution:   "macos",
		Release:        darwinRelease(),

		Architecture: runtime.GOARCH,
		Hostname:     getHostname(),
		CPUModel:     getCPUModel(),
		CPUCount:     runtime.NumCPU(),
		TotalRAM:     getTotalRAM(),
		DefaultShell: getDefaultShell(),
		HomeDir:      getHomeDir(),
		Uptime:       getUptime(),
		Filesystem:   getRootFilesystemType(),
	}
}

// detectMacPackageManager detects if a package manager (i.e. brew) is installed
func detectMacPackageManager() string {
	if utils.IsCommandAvailable("brew") {
		return "brew"
	}
	return ""
}

// darwinRelease gets the Mac OS release info
func darwinRelease() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(out))
}

// getHostname returns the hostname of the current machine
func getHostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	return hn
}

// getCPUModel returns the CPU make/model info
func getCPUModel() string {
	out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(out))
}

// getTotalRam returns a byte representation of the total memory on the current machine
func getTotalRAM() uint64 {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0
	}

	memStr := strings.TrimSpace(string(out))
	mem, err := strconv.ParseUint(memStr, 10, 64)
	if err != nil {
		return 0
	}

	return mem
}

// getDefaultShell returns the user's shell
func getDefaultShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}

	return shell
}

// getHomeDir returns the user's $HOME directory
func getHomeDir() string {
	usr, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return usr
}

// getUptime returns the machine's uptime
func getUptime() time.Duration {
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err != nil {
		return 0
	}

	// Example output: { sec = 1689253200, usec = 0 } Sat Jul 13 17:00:00 2024
	parts := strings.Split(string(out), " ")

	for i, part := range parts {
		if part == "sec" && i+2 < len(parts) {
			secStr := strings.Trim(parts[i+2], ",")
			sec, err := strconv.ParseInt(secStr, 10, 64)
			if err == nil {
				bootTime := time.Unix(sec, 0)
				return time.Since(bootTime)
			}
		}
	}

	return 0
}

// getRootFilesystem returns the filesystem type
func getRootFilesystemType() string {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return "unknown"
	}

	switch stat.Type {
	case 0x42465331:
		return "ufs"
	case 0x00011954:
		return "hfs"
	case 0x2FC12FC1:
		return "apfs"
	default:
		return "unknown"
	}
}
