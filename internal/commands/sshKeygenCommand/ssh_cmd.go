package sshcommand

import (
	"github.com/spf13/cobra"
)

// NewSSHCommand is the parent command for SSH-related subcommands
func NewSSHCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "SSH-related utilities",
		Long:  "Parent command for SSH-related operations such as key generation and management.",
	}

	// Subcommands
	cmd.AddCommand(NewSSHKeygenCommand())

	return cmd
}
