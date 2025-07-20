package platformservice

import (
	"fmt"
	"net"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/jackpal/gateway"
	"github.com/klauspost/cpuid/v2"
)

type NetworkInterface struct {
	Name            string
	HardwareAddress string
	Flags           []string
	IPAddresses     []string
}

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

	Interfaces []NetworkInterface
	Hostname   string
	DNSServers []string
	GatewayIPs []string
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
// func (p PlatformInfo) String() string {
// 	return fmt.Sprintf(
// 		`Platform Information:
//   OS:            %s
//   Architecture:  %s
//   OS Release:    %s
//   User:          %s (%s)
//   Default Shell: %s
//   Home Dir:      %s
//   Uptime:        %s
//   Total RAM:     %.2f GB
//   CPU Cores:     %d
//   CPU Threads:   %d
//   CPU Sockets:   %d
//   CPU Model:     %s
//   CPU Vendor:    %s`,
// 		p.OS,
// 		p.Arch,
// 		p.OSRelease,
// 		p.CurrentUser.Name,
// 		p.CurrentUser.Username,
// 		p.DefaultShell,
// 		p.UserHomeDir,
// 		p.Uptime.String(),
// 		float64(p.TotalRAM)/(1024*1024*1024),
// 		p.CPUCores,
// 		p.CPUThreads,
// 		p.CPUSockets,
// 		p.CPUModel,
// 		p.CPUVendor,
// 	)
// }

func (p PlatformInfo) Format(includeNet bool) string {
	var builder strings.Builder

	builder.WriteString("Platform Information:\n")
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

	if includeNet {
		// Append interface info
		builder.WriteString("\nNetwork Information:")

		builder.WriteString("\n  +NetworkInterfaces:\n")
		for _, iface := range p.Interfaces {
			builder.WriteString(fmt.Sprintf("    - %s (%s)\n", iface.Name, iface.HardwareAddress))
			if len(iface.Flags) > 0 {
				builder.WriteString(fmt.Sprintf("      Flags: %s\n", strings.Join(iface.Flags, ", ")))
			}
			if len(iface.IPAddresses) > 0 {
				builder.WriteString(fmt.Sprintf("      IPs:   %s\n", strings.Join(iface.IPAddresses, ", ")))
			}

			builder.WriteString("\n")
		}

		// DNS / Gateways
		if len(p.DNSServers) > 0 {
			builder.WriteString(fmt.Sprintf("  +DNS Servers:      %s", strings.Join(p.DNSServers, ", ")))
		}

		if len(p.GatewayIPs) > 0 {
			builder.WriteString(fmt.Sprintf("\n  +Default Gateways: %s", strings.Join(p.GatewayIPs, ", ")))
		}
	}

	return builder.String()
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

	// Get hostname
	hostname, _ := os.Hostname()
	pi.Hostname = hostname

	// Get network interfaces
	pi.Interfaces = detectNetworkInterfaces()

	// Get DNS (platform-specific)
	pi.DNSServers = detectDNSServers()

	// Optional: default gateway(s)
	pi.GatewayIPs = detectDefaultGateways()

	return pi, nil
}

// detectNetworkInterfaces gathers network interface details.
func detectNetworkInterfaces() []NetworkInterface {
	var result []NetworkInterface

	ifaces, err := net.Interfaces()
	if err != nil {
		return result
	}

	for _, iface := range ifaces {
		ni := NetworkInterface{
			Name:            iface.Name,
			HardwareAddress: iface.HardwareAddr.String(),
		}

		// Add flags (e.g. up, loopback)
		for _, f := range []net.Flags{
			net.FlagUp, net.FlagLoopback, net.FlagBroadcast,
			net.FlagMulticast, net.FlagPointToPoint,
		} {
			if iface.Flags&f != 0 {
				ni.Flags = append(ni.Flags, f.String())
			}
		}

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				ni.IPAddresses = append(ni.IPAddresses, addr.String())
			}
		}

		result = append(result, ni)
	}

	return result
}

func detectDefaultGateways() []string {
	var gateways []string
	gw, err := gateway.DiscoverGateway()

	if err == nil && gw != nil && !gw.Equal(net.IPv4zero) {
		gateways = append(gateways, gw.String())
	}

	return gateways
}
