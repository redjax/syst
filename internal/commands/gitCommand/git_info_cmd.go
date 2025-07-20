package gitcommand

import (
	gitservice "github.com/redjax/syst/internal/services/gitService"
	"github.com/spf13/cobra"
)

func NewGitInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show information about the current Git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			gitservice.PrintRepoInfo()

			return nil
		},
	}

	return cmd
}
