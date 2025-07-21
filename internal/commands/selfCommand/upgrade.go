package selfcommand

import (
	"github.com/redjax/syst/internal/version"

	"github.com/spf13/cobra"
)

// When --check is passed, don't do an upgrade, just check if one is available
var checkOnly bool

// NewUpgradeCommand creates the 'self upgrade' command
func NewUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade syst CLI to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version.UpgradeSelf(cmd, args, checkOnly)
		},
	}

	// Register flags
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for latest version, don't upgrade if one is found.")

	return cmd
}
