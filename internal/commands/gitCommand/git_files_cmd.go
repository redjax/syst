package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/filesService"
	"github.com/spf13/cobra"
)

func NewGitFilesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "File analysis and statistics",
		Long:  "Analyze repository files including size, frequency of changes, and type breakdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesService.RunFileAnalysis()
		},
	}

	return cmd
}
