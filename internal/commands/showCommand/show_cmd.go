package showCommand

import (
	"github.com/spf13/cobra"
)

var verbose bool

func Verbose() bool {
	return verbose
}

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
	showCmd.AddCommand(NewShowNetCmd())
	showCmd.AddCommand(NewDiskInfoCmd())
	showCmd.AddCommand(NewShowTimeCmd())

	showCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output (includes all system/virtual disks, more detail, etc)")

	return showCmd
}
