package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSelfCommand creates the 'self' parent command, which adds some of the other
// commands in this file as subcommands.
//
// When adding this as a subcommand to another CLI, use:
//
//	cmd.AddCommand(version.NewSelfCommand())
func NewSelfCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self",
		Short: "Manage this syst CLI",
		Long:  "Self-management operations for syst, e.g. upgrade to latest version.",
	}

	// Attach 'upgrade' as a subcommand
	cmd.AddCommand(NewUpgradeCommand())
	// Attach 'info' as a subcommand
	cmd.AddCommand(NewPackageInfoCommand())
	// Attach 'version' as a subcommand
	cmd.AddCommand(NewVersionCommand())

	return cmd
}

// NewVersionCommand adds a 'version' subcommand, which prints the package's version.
//
// When adding this as a subcommand to another CLI, use:
//
//	cmd.AddCommand(version.NewSelfCommand())
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI's version",
		Run: func(cmd *cobra.Command, args []string) {
			pkgInfo := GetPackageInfo()
			fmt.Printf("package: %s version:%s commit:%s date:%s\n",
				pkgInfo.PackageName,
				pkgInfo.PackageVersion,
				pkgInfo.PackageCommit,
				pkgInfo.PackageReleaseDate,
			)
		},
	}
}

// NewPackageInfoCommand adds a subcommand 'info' and prints info about the package.
//
// When adding this as a subcommand to another CLI, use:
//
//	cmd.AddCommand(version.NewPackageInfoCommand())
func NewPackageInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show info about the current package",
		RunE:  showPackageInfo,
	}
}

// NewUpgradeCommand creates the 'self upgrade' command.
// When adding this as a subcommand to another CLI, use:
//
//	cmd.AddCommand(version.NewUpgradeCommand())
func NewUpgradeCommand() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade syst CLI to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return UpgradeSelf(cmd, args, checkOnly)
		},
	}

	// Register flags
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for latest version, don't upgrade if one is found.")

	return cmd
}
