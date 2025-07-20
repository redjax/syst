package platformservice

import (
	"fmt"
	"strings"
)

type NetworkInterface struct {
	Name            string
	HardwareAddress string
	Flags           []string
	IPAddresses     []string
}

func (p PlatformInfo) PrintNetFormat() string {
	var builder strings.Builder

	builder.WriteString("Network Information:\n")

	builder.WriteString("  +Network Interfaces:\n")
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

	if len(p.DNSServers) > 0 {
		builder.WriteString(fmt.Sprintf("  +DNS Servers:      %s\n", strings.Join(p.DNSServers, ", ")))
	}

	if len(p.GatewayIPs) > 0 {
		builder.WriteString(fmt.Sprintf("  +Default Gateways: %s\n", strings.Join(p.GatewayIPs, ", ")))
	}

	return builder.String()
}
