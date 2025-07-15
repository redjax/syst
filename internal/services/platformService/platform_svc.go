package platformservice

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/klauspost/cpuid/v2"
)

// PlatformInfo holds information about the current host system.
type PlatformInfo struct {
	// e.g. "windows", "linux", "darwin"
	OS string
	// e.g. "amd64", "arm64"
	Arch string
	// e.g. "10", "24.04", "12", "Sonoma"
	OSRelease    string
	CurrentUser  PlatformUser
	DefaultShell string
	UserHomeDir  string
	Uptime       time.Duration
	// bytes
	TotalRAM uint64
	// physical cores
	CPUCores int
	// logical cores (threads)
	CPUThreads int
	// set to 1 by default; see note below
	CPUSockets int
	CPUModel   string
	CPUVendor  string
}

type PlatformUser struct {
	Username string
	Uid      string
	Gid      string
	Name     string
	HomeDir  string
}

// Add a printstring method to the PlatformInfo class.
// Controls how the class displays when printed directly. Like Python's __repr__.
func (p PlatformInfo) String() string {
	return fmt.Sprintf(
		`Platform Information:
  OS:            %s
  Architecture:  %s
  OS Release:    %s
  User:          %s (%s)
  Default Shell: %s
  Home Dir:      %s
  Uptime:        %s
  Total RAM:     %.2f GB
  CPU Cores:     %d
  CPU Threads:   %d
  CPU Sockets:   %d
  CPU Model:     %s
  CPU Vendor:    %s`,
		p.OS,
		p.Arch,
		p.OSRelease,
		p.CurrentUser.Name,
		p.CurrentUser.Username,
		p.DefaultShell,
		p.UserHomeDir,
		p.Uptime.String(),
		float64(p.TotalRAM)/(1024*1024*1024),
		p.CPUCores,
		p.CPUThreads,
		p.CPUSockets,
		p.CPUModel,
		p.CPUVendor,
	)
}

// GatherPlatformInfo collects platform information in a cross-platform way.
func GatherPlatformInfo() (*PlatformInfo, error) {
	pi := &PlatformInfo{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		CPUCores:   cpuid.CPU.PhysicalCores,
		CPUThreads: cpuid.CPU.LogicalCores,
		// Most systems are single-socket; platform-specific detection needed for more
		CPUSockets: 1,
		CPUModel:   cpuid.CPU.BrandName,
		CPUVendor:  cpuid.CPU.VendorString,
	}

	// Get user home directory
	if u, err := user.Current(); err == nil {
		pi.UserHomeDir = u.HomeDir
	}

	// Get default shell (best effort, platform-specific)
	pi.DefaultShell = detectDefaultShell()

	// Get OS release/version (platform-specific)
	pi.OSRelease = detectOSRelease()

	// Get uptime (platform-specific)
	pi.Uptime = detectUptime()

	// Get total RAM (platform-specific)
	pi.TotalRAM = detectTotalRAM()

	// Get user home directory and user info
	if u, err := user.Current(); err == nil {
		pi.UserHomeDir = u.HomeDir
		pi.CurrentUser = PlatformUser{
			Username: u.Username,
			Uid:      u.Uid,
			Gid:      u.Gid,
			Name:     u.Name,
			HomeDir:  u.HomeDir,
		}
	}

	return pi, nil
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
