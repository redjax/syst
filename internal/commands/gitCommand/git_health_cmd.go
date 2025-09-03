package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/healthService"
	"github.com/spf13/cobra"
)

// NewGitHealthCommand creates the git health command
func NewGitHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Repository health check",
		Long:  "Analyze repository health including large files, potential issues, security concerns, and quality metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return healthService.RunHealthCheck()
		},
	}
}
