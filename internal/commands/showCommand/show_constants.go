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
			fmt.Printf("Family:         %s\n", consts.Family)
			fmt.Printf("Distribution:   %s\n", consts.Distribution)
			fmt.Printf("Release:        %s\n", consts.Release)
			fmt.Printf("Package Manager: %s\n", consts.PackageManager)
		},
	}
}
