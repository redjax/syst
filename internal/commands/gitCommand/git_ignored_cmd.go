package gitcommand

import (
	"github.com/redjax/syst/internal/services/gitService/ignoredService"
	"github.com/spf13/cobra"
)

func NewGitIgnoredCommand() *cobra.Command {
	var (
		showAll    bool
		showSizes  bool
		outputPath string
	)

	cmd := &cobra.Command{
		Use:   "ignored",
		Short: "List files and directories ignored by Git",
		Long: `Display all files and directories that are ignored by Git based on .gitignore rules.
		
This command shows:
- Files ignored by .gitignore patterns
- Directories ignored by .gitignore patterns  
- File sizes (optional)
- Export to file (optional)

The output includes both tracked ignored files and untracked ignored files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ignoredService.IgnoredOptions{
				ShowAll:    showAll,
				ShowSizes:  showSizes,
				OutputPath: outputPath,
			}
			return ignoredService.RunIgnoredFiles(opts)
		},
	}

	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all ignored files including hidden files")
	cmd.Flags().BoolVarP(&showSizes, "sizes", "s", false, "Show file sizes for ignored files")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Export ignored files list to a file")

	return cmd
}
