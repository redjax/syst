package platformservice

import (
	"fmt"
	"net"
	"strings"

	"github.com/jackpal/gateway"
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
