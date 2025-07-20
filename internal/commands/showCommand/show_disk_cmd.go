package showCommand

import (
	"fmt"

	platformservice "github.com/redjax/syst/internal/services/platformService"
	"github.com/spf13/cobra"
)

func NewDiskInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disks",
		Short: "Show disk info",
		Long:  `Shows all disks, their mountpoint, total size, used size, and used %%`,
		Run: func(cmd *cobra.Command, args []string) {
			info, err := platformservice.GatherPlatformInfo(Verbose())
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			fmt.Println(info.PrintDiskFormat())
		},
	}

	return cmd
}
