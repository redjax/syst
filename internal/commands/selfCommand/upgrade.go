package selfcommand

import (
	"github.com/redjax/syst/internal/version"

	"github.com/spf13/cobra"
)

// NewUpgradeCommand creates the 'self upgrade' command
func NewUpgradeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade syst CLI to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.UpgradeSelf(cmd, args)
		},
	}
}
