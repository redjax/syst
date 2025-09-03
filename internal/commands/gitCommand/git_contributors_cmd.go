package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/contributorsService"
	"github.com/spf13/cobra"
)

func NewGitContributorsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contributors",
		Short: "Developer statistics and analysis",
		Long:  "Show commit counts, line changes, and activity by author with interactive exploration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return contributorsService.RunContributorsAnalysis()
		},
	}

	return cmd
}
