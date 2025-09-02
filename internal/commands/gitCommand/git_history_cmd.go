package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/historyService"
	"github.com/spf13/cobra"
)

// NewGitHistoryCommand creates the git history command
func NewGitHistoryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "history",
		Short: "Advanced git history views",
		Long:  "Interactive timeline, commit frequency analysis, and tag/release history browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			return historyService.RunHistoryExplorer()
		},
	}
}
