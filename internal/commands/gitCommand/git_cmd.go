package gitcommand

import (
	"github.com/spf13/cobra"
)

// NewGitCommand returns the git command with all subcommands attached.
func NewGitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Git helper commands for syst",
		Long:  "Enhanced git helper operations like prune, for use with syst CLI.",
	}

	// Add subcommands
	cmd.AddCommand(NewGitPruneCommand())

	return cmd
}
