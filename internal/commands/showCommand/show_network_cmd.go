package showCommand

import (
	"fmt"

	platformservice "github.com/redjax/syst/internal/services/platformService"
	"github.com/spf13/cobra"
)

func NewShowNetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "net",
		Short: "Show only network-related platform information",
		Long:  `Shows network interfaces, DNS servers, and gateway info from the current platform.`,
		Run: func(cmd *cobra.Command, args []string) {
			info, err := platformservice.GatherPlatformInfo()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Println(info.NetFormat())
		},
	}

	return cmd
}
