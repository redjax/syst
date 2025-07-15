package showCommand

import (
	"fmt"

	"github.com/redjax/syst/internal/constants"
	utils "github.com/redjax/syst/internal/utils/convert"

	"github.com/spf13/cobra"
)

func NewConstantsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "constants",
		Short: "Show platform constants",
		Long: `syst generates a set of global constants based on the current platform. The show constants command displays those constants to the user.

Constant data includes: system hostname & uptime, the $HOME path and default shell, OS/release info, CPU info, total RAM, and the filesystem.
		`,
		Run: func(cmd *cobra.Command, args []string) {
			// Get constants
			consts := constants.GetPlatformConstants()
			// Convert TotalRAM bytes to human-readable string
			totalRam := utils.BytesToHumanReadable(consts.TotalRAM)

			// Print constants
			fmt.Printf("Uptime=%s\n", consts.Uptime)
			fmt.Printf("Hostname=%s\n", consts.Hostname)
			fmt.Printf("Default Shell=%s\n", consts.DefaultShell)
			fmt.Printf("$HOME path=%s\n", consts.HomeDir)
			fmt.Printf("Family=%s\n", consts.Family)
			fmt.Printf("Distribution=%s\n", consts.Distribution)
			fmt.Printf("Release=%s\n", consts.Release)
			fmt.Printf("Package Manager=%s\n", consts.PackageManager)
			fmt.Printf("CPU Architecture=%s\n", consts.Architecture)
			fmt.Printf("CPU Model=%s\n", consts.CPUModel)
			fmt.Printf("CPU Count=%d\n", consts.CPUCount)
			fmt.Printf("TotalRAM=%s\n", totalRam)
			fmt.Printf("Filesystem=%s\n", consts.Filesystem)
		},
	}
}
