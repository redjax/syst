package selfcommand

import (
	"fmt"

	"github.com/redjax/syst/internal/version"

	"github.com/spf13/cobra"
)

// NewUpgradeCommand creates the 'self upgrade' command
func NewPackageInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show info about the current package",
		RunE:  showPackageInfo,
	}
}

func showPackageInfo(cmd *cobra.Command, args []string) error {
	pkgInfo := version.GetPackageInfo()

	fmt.Printf("Program: %s\nOwner: %s\nRepository Name: %s\nRepository URL: %s\n", pkgInfo.PackageName, pkgInfo.RepoUser, pkgInfo.RepoName, pkgInfo.RepoUrl)

	return nil
}
