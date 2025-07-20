package whichcommand

import (
	"fmt"

	"github.com/redjax/syst/internal/services/platformService/capabilities"
	"github.com/spf13/cobra"
)

func NewWhichCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "which [binary]",
		Short: "A cross-platform wrapper that functions like UNIX 'which'.",
		Long:  `Test a CLI command to see if it's installed/available, and if so return the path.`,
		Args:  cobra.MinimumNArgs(1), // Require at least one argument
		RunE: func(cmd *cobra.Command, args []string) error {
			bin := args[0] // Get the first argument

			path, err := capabilities.Which(bin)
			if err != nil {
				fmt.Printf("Command '%s' not found in PATH\n", bin)
				return nil
			}

			// Print the full path
			fmt.Printf("Found command '%s' at path: %s\n", bin, path)

			return nil
		},
	}

	return cmd
}
