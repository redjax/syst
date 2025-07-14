package showCommand

import (
	"fmt"

	"github.com/redjax/syst/internal/constants"

	"github.com/spf13/cobra"
)

func NewConstantsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "constants",
		Short: "Show platform constants",
		Run: func(cmd *cobra.Command, args []string) {
			consts := constants.GetPlatformConstants()
			fmt.Printf("Uptime: %s\n", consts.Uptime)
			fmt.Printf("Hostname: %s\n", consts.Hostname)
			fmt.Printf("Default Shell: %s\n", consts.DefaultShell)
			fmt.Printf("$HOME path: %s\n", consts.HomeDir)
			fmt.Printf("Family:         %s\n", consts.Family)
			fmt.Printf("Distribution:   %s\n", consts.Distribution)
			fmt.Printf("Release:        %s\n", consts.Release)
			fmt.Printf("Package Manager: %s\n", consts.PackageManager)
			fmt.Printf("CPU Architecture: %s\n", consts.Architecture)
			fmt.Printf("CPU Model: %s\n", consts.CPUModel)
			fmt.Printf("CPU Count: %d\n", consts.CPUCount)
			fmt.Printf("TotalRAM (bytes): %d\n", consts.TotalRAM)
			fmt.Printf("Filesystem: %s\n", consts.Filesystem)
		},
	}
}
