package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/diffService"
	"github.com/spf13/cobra"
)

func NewGitDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff [branch1] [branch2]",
		Short: "Interactive change analysis between refs",
		Long:  "Show changes between branches/commits/tags with interactive file-by-file diff viewer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return diffService.RunDiffExplorer(args)
		},
	}

	return cmd
}
