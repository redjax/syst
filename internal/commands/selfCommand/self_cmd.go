package selfcommand

import (
	"github.com/spf13/cobra"

	"github.com/redjax/syst/internal/version"
)

// NewSelfCommand creates the 'self' parent command
func NewSelfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self",
		Short: "Manage this syst CLI",
		Long:  "Self-management operations for syst, e.g. upgrade to latest version.",
	}

	// Attach 'upgrade' as a subcommand
	cmd.AddCommand(version.NewUpgradeCommand())
	// Attach 'info' as a subcommand
	cmd.AddCommand(NewPackageInfoCommand())

	return cmd
}
