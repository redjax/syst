package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/statusService"
	"github.com/spf13/cobra"
)

func NewGitStatusCommand() *cobra.Command {
	var (
		showAll    bool
		showColors bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show tracked and untracked files",
		Long: `Display the status of files in the repository showing which files are:
- Tracked by Git (green dot •)
- Untracked/new files (gray dot •)
- Modified files
- Staged files

This provides a comprehensive view of your repository's file status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := statusService.StatusOptions{
				ShowAll:    showAll,
				ShowColors: showColors,
			}
			return statusService.RunGitStatus(opts)
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", true, "Show all files including untracked")
	cmd.Flags().BoolVarP(&showColors, "colors", "c", true, "Use colors to indicate file status")

	return cmd
}
