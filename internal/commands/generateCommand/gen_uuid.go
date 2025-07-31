package generatecommand

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewUUIDCommand() *cobra.Command {
	var hexOutput bool

	cmd := &cobra.Command{
		Use:   "uuid",
		Short: "Return a random UUID.",
		Run: func(cmd *cobra.Command, args []string) {
			_uuid := uuid.New()

			if hexOutput {
				// Print UUID as a continuous hex string without dashes
				fmt.Println(strings.ReplaceAll(_uuid.String(), "-", ""))
			} else {
				// Default string format
				fmt.Println(_uuid.String())
			}
		},
	}

	cmd.Flags().BoolVar(&hexOutput, "hex", false, "Output UUID as hex string without dashes")

	return cmd
}
