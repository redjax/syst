// scanpath_cmd.go
package scanPathCommand

import (
	"github.com/redjax/syst/internal/services/pathScanService/scan"
	"github.com/spf13/cobra"
)

func NewScanPathCommand() *cobra.Command {
	var (
		// Target path to scan
		path string
		// Limit scanned items
		limit int
		// Sort by column (name, size, created, modified, owner, permissions)
		sortBy string
		// Sort order, 'asc' or 'desc'
		order string
		// Filter results by column, i.e. 'size <10MB', 'created >2022-01-01'
		filter string
		// Recursively traverse subdirectories including git repositories
		recursive bool
	)

	cmd := &cobra.Command{
		Use:   "scanpath",
		Short: "Scan a directory and list items with metadata",
		Long: `Scan a directory and list files with metadata like size, creation time, permissions, etc.

Control the scan & results using flags like --limit (to limit the number of items scanned), and --order (to control sorting order, asc/desc).

Use the --recursive flag to traverse subdirectories.

Run syst scanpath --help to see all options.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return scan.ScanDirectory(path, limit, sortBy, order, filter, recursive) // ✅ direct call
		},
	}

	// Add command flags
	cmd.Flags().StringVarP(&path, "path", "p", ".", "Directory path to scan")
	cmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of results (0 = unlimited)")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Column to sort by (name, size, created, modified, owner, permissions)")
	cmd.Flags().StringVarP(&order, "order", "o", "asc", "Sort order: asc or desc")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter results (e.g. 'size <10MB', 'created >2022-01-01')")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively traverse subdirectories")

	return cmd
}
