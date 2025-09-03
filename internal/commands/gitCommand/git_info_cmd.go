package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/infoService"
	"github.com/spf13/cobra"
)

func NewGitInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Interactive TUI for exploring repository statistics",
		Long:  "Launch an interactive terminal user interface for exploring detailed repository information including remotes, branches, contributors, and more.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return infoService.RunRepoInfoTUI()
		},
	}

	return cmd
}
