package versioncommand

import (
	"fmt"

	"github.com/redjax/syst/internal/version"
	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI's version",
		Run: func(cmd *cobra.Command, args []string) {
			// Print version string
			fmt.Printf("version:%s commit:%s date:%s", version.Version, version.Commit, version.Date)
		},
	}
}
