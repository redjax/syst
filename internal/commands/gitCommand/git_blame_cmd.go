package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/blameService"
	"github.com/spf13/cobra"
)

func NewGitBlameCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blame [file]",
		Short: "Interactive file investigation",
		Long:  "Interactive blame viewer with line-by-line author information and historical changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return blameService.RunBlameViewer(args)
		},
	}

	return cmd
}
