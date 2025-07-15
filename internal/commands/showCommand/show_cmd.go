package showCommand

import (
	"github.com/spf13/cobra"
)

func NewShowCmd() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show commands print information in the selected domain, i.e. show platform.",
		Long: `Print configuration/debug data.

Show information about the platform, & dynamic constants created on the fly based on platform.
	
Run syst show --help to see all options.
`,
	}

	// Attach subcommands
	showCmd.AddCommand(NewPlatformCmd())
	showCmd.AddCommand(NewConstantsCmd())

	return showCmd
}
