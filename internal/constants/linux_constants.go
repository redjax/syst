//go:build linux
// +build linux

package constants

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// platformConstants detects the distro and merges distro-specific and common info.
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
	release := getOSReleaseField(osRelease, "VERSION_ID")

	var base PlatformConstants
	switch id {
	case "ubuntu":
		base = ubuntuConstants(release)
	case "debian":
		base = debianConstants(release)
	case "fedora":
		base = fedoraConstants(release)
	case "opensuse", "opensuse-leap":
		base = openSUSEConstants(release)
	default:
		base = PlatformConstants{
			Family:         "unknown",
			Distribution:   id,
			Release:        release,
			PackageManager: "unknown",
		}
	}

	// Fill in common fields:
	base.Architecture = runtime.GOARCH
	base.Hostname = getHostname()
	base.CPUModel = getCPUModel()
	base.CPUCount = runtime.NumCPU()
	base.TotalRAM = getTotalRAM()
	base.DefaultShell = getDefaultShell()
	base.HomeDir = getHomeDir()
	base.Uptime = getUptime()
	base.Filesystem = getRootFilesystemType()

	return base
}

// Shared helpers

func getOSReleaseField(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, key+"=") {
			return strings.Trim(line[len(key)+1:], "\"")
		}
	}
	return ""
}

func getHostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hn
}

func getCPUModel() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "unknown"
}

func getTotalRAM() uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseUint(fields[1], 10, 64)
				if err == nil {
					return kb * 1024
				}
			}
		}
	}
	return 0
}

func getDefaultShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}
	return shell
}

func getHomeDir() string {
	usr, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return usr
}

func getUptime() time.Duration {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}
	return time.Duration(seconds) * time.Second
}

func getRootFilesystemType() string {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return "unknown"
	}

	switch stat.Type {
	case 0xEF53:
		return "ext4"
	case 0x9123683E:
		return "btrfs"
	case 0x58465342:
		return "xfs"
	case 0x6969:
		return "nfs"
	case 0x52654973:
		return "reiserfs"
	default:
		return "unknown"
	}
}
