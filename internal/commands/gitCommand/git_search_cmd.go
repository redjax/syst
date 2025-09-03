package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/searchService"
	"github.com/spf13/cobra"
)

// NewGitSearchCommand creates the git search command
func NewGitSearchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Advanced repository search",
		Long:  "Search commits, authors, files, and content across repository history with interactive results",
		RunE: func(cmd *cobra.Command, args []string) error {
			return searchService.RunAdvancedSearch(args)
		},
	}
}
