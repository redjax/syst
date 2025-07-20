//go:build linux
// +build linux

package platformservice

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func detectOSRelease() string {
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		lines := strings.Split(string(data), "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(line[len("PRETTY_NAME="):], "\"")
			}
		}
	}

	return ""
}

func detectUptime() time.Duration {
	data, err := os.ReadFile("/proc/uptime")
	if err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			if seconds, err := strconv.ParseFloat(fields[0], 64); err == nil {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	return 0
}

func detectTotalRAM() uint64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		lines := strings.Split(string(data), "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
						return kb * 1024 // Convert kB to bytes
					}
				}
			}
		}
	}

	return 0
}

// detectDefaultShell tries to find the user's default shell.
func detectDefaultShell() string {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("ComSpec") // Usually "C:\\Windows\\System32\\cmd.exe"
	default:
		shell := os.Getenv("SHELL")
		if shell != "" {
			return shell
		}

		return "/bin/sh"
	}
}

func detectDNSServers() []string {
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil
	}

	var servers []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "nameserver ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1])
			}
		}
	}

	return servers
}
