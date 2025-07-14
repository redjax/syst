package showCommand

import (
	"fmt"
	"os"

	platformservice "github.com/redjax/syst/internal/services/platformService"

	"github.com/spf13/cobra"
)

func NewPlatformCmd() *cobra.Command {
	var property string

	cmd := &cobra.Command{
		Use:   "platform",
		Short: "Show platform information",
		Run: func(cmd *cobra.Command, args []string) {
			info, err := platformservice.GatherPlatformInfo()
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			if property != "" {
				// Print only the requested property (simple example)
				switch property {
				case "hostname":
					hostname, _ := os.Hostname()
					fmt.Println(hostname)
				case "os":
					fmt.Println(info.OS)
				// Add more cases as needed
				default:
					fmt.Printf("Unknown property: %s\n", property)
				}
			} else {
				fmt.Printf("%+v\n", info)
			}
		},
	}
	cmd.Flags().StringVar(&property, "property", "", "Show only a specific property (e.g., hostname)")
	return cmd
}
