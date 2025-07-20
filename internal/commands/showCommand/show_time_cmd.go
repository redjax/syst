package showCommand

import (
	"fmt"

	platformservice "github.com/redjax/syst/internal/services/platformService"
	"github.com/spf13/cobra"
)

func NewShowTimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "time",
		Short: "Show host's time info",
		Long:  `Show information about host's time, like the current time, timezone, and offset seconds.`,
		Run: func(cmd *cobra.Command, args []string) {
			info, err := platformservice.GatherPlatformInfo(Verbose())
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Println(info.PrintTimeFormat())
		},
	}

	return cmd
}
