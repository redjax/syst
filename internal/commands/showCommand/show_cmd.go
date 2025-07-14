package showCommand

import (
	"github.com/spf13/cobra"
)

func NewShowCmd() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show commands print information in the selected domain, i.e. show platform.",
	}

	// Attach subcommands
	showCmd.AddCommand(NewPlatformCmd())
	showCmd.AddCommand(NewConstantsCmd())

	return showCmd
}
