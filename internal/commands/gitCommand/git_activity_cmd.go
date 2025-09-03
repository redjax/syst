package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/activity"
	"github.com/spf13/cobra"
)

// NewGitActivityCommand returns the git activity command.
func NewGitActivityCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activity",
		Short: "Repository activity dashboard",
		Long:  "Show recent commit activity, development patterns, and commit frequency analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			return activity.RunActivityDashboard()
		},
	}

	return cmd
}
