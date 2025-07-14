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
			info, err := platformservice.GatherPlatformInfo()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			if len(properties) > 0 {
				for _, prop := range properties {
					switch strings.ToLower(prop) {
					case "hostname":
						hostname, _ := os.Hostname()
						fmt.Printf("hostname: %s\n", hostname)
					case "os":
						fmt.Printf("os: %s\n", info.OS)
					case "arch":
						fmt.Printf("arch: %s\n", info.Arch)
					case "osrelease", "release":
						fmt.Printf("osrelease: %s\n", info.OSRelease)
					case "shell", "defaultshell":
						fmt.Printf("defaultshell: %s\n", info.DefaultShell)
					case "homedir", "userhomedir", "userhome", "home":
						fmt.Printf("userhomedir: %s\n", info.UserHomeDir)
					case "uptime":
						fmt.Printf("uptime: %s\n", info.Uptime)
					case "totalram":
						fmt.Printf("totalram: %d\n", info.TotalRAM)
					case "cpucores":
						fmt.Printf("cpucores: %d\n", info.CPUCores)
					case "cputhreads":
						fmt.Printf("cputhreads: %d\n", info.CPUThreads)
					case "cpumodel":
						fmt.Printf("cpumodel: %s\n", info.CPUModel)
					case "cpuvendor":
						fmt.Printf("cpuvendor: %s\n", info.CPUVendor)
					default:
						fmt.Printf("Unknown property: %s\n", prop)
					}
				}
			} else {
				fmt.Printf("%+v\n", info)
			}
		},
	}
	cmd.Flags().StringSliceVar(&properties, "property", nil, "Show only specific properties (can be repeated)")
	return cmd
}
