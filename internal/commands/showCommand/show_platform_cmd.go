package showCommand

import (
	"fmt"
	"os"
	"strings"

	platformservice "github.com/redjax/syst/internal/services/platformService"

	"github.com/spf13/cobra"
)

func NewPlatformCmd() *cobra.Command {
	var properties []string
	var includeNet bool

	cmd := &cobra.Command{
		Use:   "platform",
		Short: "Show platform information. You can pass multiple --property <propertyname> flags.",
		Long: `Show detailed platform information.

Available properties for --property:
  - hostname
  - os
  - arch
  - osrelease (alias: release)
  - defaultshell
  - userhomedir
  - uptime
  - totalram
  - cpucores
  - cputhreads
  - cpumodel
  - cpuvendor
`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get PlatformInfo object
			info, err := platformservice.GatherPlatformInfo()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			// Filter properties if any --property flags detected
			if len(properties) > 0 {
				// Match each property to a PlatformInfo value
				for _, prop := range properties {
					switch strings.ToLower(prop) {
					// Hostname property requested
					case "hostname":
						hostname, _ := os.Hostname()
						fmt.Printf("hostname: %s\n", hostname)
					// OS name property requested
					case "os":
						fmt.Printf("os: %s\n", info.OS)
					// CPU architecture property requested
					case "arch":
						fmt.Printf("arch: %s\n", info.Arch)
					// OS release property requested
					case "osrelease", "release":
						fmt.Printf("osrelease: %s\n", info.OSRelease)
					// Default shell property requested
					case "shell", "defaultshell":
						fmt.Printf("defaultshell: %s\n", info.DefaultShell)
					// $HOME dir property requested
					case "homedir", "userhomedir", "userhome", "home":
						fmt.Printf("userhomedir: %s\n", info.UserHomeDir)
					// Uptime property requested
					case "uptime":
						fmt.Printf("uptime: %s\n", info.Uptime)
					// Total RAM property requested
					case "totalram":
						fmt.Printf("totalram: %d\n", info.TotalRAM)
					// CPU core count property requested
					case "cpucores":
						fmt.Printf("cpucores: %d\n", info.CPUCores)
					// CPU thread property requested
					case "cputhreads":
						fmt.Printf("cputhreads: %d\n", info.CPUThreads)
					// CPU model info property requested
					case "cpumodel":
						fmt.Printf("cpumodel: %s\n", info.CPUModel)
					// CPU vendor info property requested
					case "cpuvendor":
						fmt.Printf("cpuvendor: %s\n", info.CPUVendor)
					// Unmatched/unknown property requested
					default:
						fmt.Printf("Unknown property: %s\n", prop)
					}
				}
			} else {
				// Print all properties
				// fmt.Printf("%+v\n", info)
				fmt.Println(info.Format(includeNet))
			}
		},
	}

	// Accept multiple --property flags
	cmd.Flags().StringSliceVar(&properties, "property", nil, "Show only specific properties (can be repeated)")

	cmd.Flags().BoolVar(&includeNet, "net", false, "Include network interfaces, DNS, and gateway information")

	return cmd
}
