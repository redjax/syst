package gitcommand

import (
	gitservice "github.com/redjax/syst/internal/services/gitService"
	"github.com/spf13/cobra"
)

func NewGitPruneCommand() *cobra.Command {
	var mainBranch string
	var confirm bool
	var force bool
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete local branches that were deleted on the remote.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return gitservice.PruneBranches(mainBranch, confirm, force, dryRun)
		},
	}

	cmd.Flags().StringVar(&mainBranch, "main-branch", "main", "The name of your main branch")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Prompt before deleting each branch")
	cmd.Flags().BoolVar(&force, "force", false, "Force delete branches using 'git branch -D'")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "List branches that would be deleted but take no action")

	return cmd
}
