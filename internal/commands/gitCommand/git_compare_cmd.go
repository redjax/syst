package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/compareService"
	"github.com/spf13/cobra"
)

func NewGitCompareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare [ref1] [ref2]",
		Short: "Comparison tools for refs",
		Long:  "Compare different branches/tags/commits showing divergence and shared history",
		RunE: func(cmd *cobra.Command, args []string) error {
			return compareService.RunComparison(args)
		},
	}

	return cmd
}
