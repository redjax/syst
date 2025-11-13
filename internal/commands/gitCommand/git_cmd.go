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
	cmd.AddCommand(NewGitSparseCloneCommand())
	cmd.AddCommand(NewGitInfoCommand())
	cmd.AddCommand(NewGitActivityCommand())
	cmd.AddCommand(NewGitBlameCommand())
	cmd.AddCommand(NewGitBranchesCommand())
	cmd.AddCommand(NewGitCompareCommand())
	cmd.AddCommand(NewGitContributorsCommand())
	cmd.AddCommand(NewGitDiffCommand())
	cmd.AddCommand(NewGitFilesCommand())
	cmd.AddCommand(NewGitHealthCommand())
	cmd.AddCommand(NewGitHistoryCommand())
	cmd.AddCommand(NewGitIgnoredCommand())
	cmd.AddCommand(NewGitSearchCommand())
	cmd.AddCommand(NewGitStatusCommand())
	cmd.AddCommand(NewGitWorktreeCommand())

	return cmd
}
