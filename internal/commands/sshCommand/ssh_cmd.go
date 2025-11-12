package sshcommand

import (
	"github.com/spf13/cobra"

	"github.com/redjax/syst/internal/commands/sshCommand/ui"
)

// NewSSHCommand is the parent command for SSH-related subcommands
func NewSSHCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH-related utilities",
		Long:  "Parent command for SSH-related operations such as key generation and management.",
		Run: func(cmd *cobra.Command, args []string) {
			// Launch Bubbletea UI if no subcommand is given
			ui.RunSSHUI()
		},
	}

	// Subcommands
	cmd.AddCommand(NewSSHKeygenCommand())

	return cmd
}
