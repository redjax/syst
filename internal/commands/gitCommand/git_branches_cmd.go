package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/branchesService"
	"github.com/spf13/cobra"
)

func NewGitBranchesCommand() *cobra.Command {
	var branchName string

	cmd := &cobra.Command{
		Use:   "branches",
		Short: "Interactive branch explorer",
		Long:  "Show all local/remote branches with interactive navigation and analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			return branchesService.RunBranchesExplorer(branchName)
		},
	}

	cmd.Flags().StringVarP(&branchName, "branch", "b", "", "Open specific branch directly")

	return cmd
}
