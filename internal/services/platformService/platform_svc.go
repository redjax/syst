package platformservice

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/klauspost/cpuid/v2"
)

// PlatformInfo holds information about the current host system.
type PlatformInfo struct {
	Hostname string

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

	Interfaces []NetworkInterface
	DNSServers []string
	GatewayIPs []string

	Disks []DiskInfo

	Time TimeInfo
}

type PlatformUser struct {
	Username string
	Uid      string
	Gid      string
	Name     string
	HomeDir  string
}

func (p PlatformInfo) PrintFormat(includeNet bool, includeDisks bool) string {
	var builder strings.Builder

	builder.WriteString("Platform Information:\n")
	builder.WriteString(fmt.Sprintf("  Hostname:      %s\n", p.Hostname))
	builder.WriteString(fmt.Sprintf("  OS:            %s\n", p.OS))
	builder.WriteString(fmt.Sprintf("  Architecture:  %s\n", p.Arch))
	builder.WriteString(fmt.Sprintf("  OS Release:    %s\n", p.OSRelease))
	builder.WriteString(fmt.Sprintf("  User:          %s (%s)\n", p.CurrentUser.Name, p.CurrentUser.Username))
	builder.WriteString(fmt.Sprintf("  Default Shell: %s\n", p.DefaultShell))
	builder.WriteString(fmt.Sprintf("  Home Dir:      %s\n", p.UserHomeDir))
	builder.WriteString(fmt.Sprintf("  Uptime:        %s\n", p.Uptime.String()))
	builder.WriteString(fmt.Sprintf("  Total RAM:     %.2f GB\n", float64(p.TotalRAM)/(1024*1024*1024)))
	builder.WriteString(fmt.Sprintf("  CPU Cores:     %d\n", p.CPUCores))
	builder.WriteString(fmt.Sprintf("  CPU Threads:   %d\n", p.CPUThreads))
	builder.WriteString(fmt.Sprintf("  CPU Sockets:   %d\n", p.CPUSockets))
	builder.WriteString(fmt.Sprintf("  CPU Model:     %s\n", p.CPUModel))
	builder.WriteString(fmt.Sprintf("  CPU Vendor:    %s\n", p.CPUVendor))
	builder.WriteString("  Time Info:\n")
	builder.WriteString(fmt.Sprintf("    Current Time: %s\n", p.Time.CurrentTime))
	builder.WriteString(fmt.Sprintf("    Timezone: %s (%s)\n", p.Time.TimezoneLong, p.Time.Timezone))
	builder.WriteString(fmt.Sprintf("    Offset: %vs\n", p.Time.OffsetSeconds))

	if includeNet {
		builder.WriteString("\n" + p.PrintNetFormat())
	}

	if includeDisks {
		builder.WriteString("\n" + p.PrintDiskFormat())
	}

	return builder.String()
}

// GatherPlatformInfo collects platform information in a cross-platform way.
func GatherPlatformInfo(verbose bool) (*PlatformInfo, error) {
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

	// Get time info
	pi.Time = getTimeInfo()

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

	// Get hostname
	hostname, _ := os.Hostname()
	pi.Hostname = hostname

	// Get network interfaces
	pi.Interfaces = detectNetworkInterfaces()

	// Get DNS (platform-specific)
	pi.DNSServers = detectDNSServers()

	// Default gateway(s)
	pi.GatewayIPs = detectDefaultGateways()

	// Disk info
	pi.Disks = detectDisks(verbose)

	return pi, nil
}
