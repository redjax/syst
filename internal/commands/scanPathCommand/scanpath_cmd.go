// scanpath_cmd.go
package scanPathCommand

import (
	"github.com/redjax/syst/internal/services/pathScanService/scan"
	"github.com/spf13/cobra"
)

func NewScanPathCommand() *cobra.Command {
	var (
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
		Use:   "scanpath [path]",
		Short: "Scan a directory and list items with metadata in an interactive TUI",
		Long: `Scan a directory and list files with metadata like size, creation time, permissions, etc.

The results are displayed in an interactive terminal user interface where you can:
- Navigate with arrow keys
- Sort by different columns
- Filter results
- Open files/directories

Control the scan & results using flags like --limit (to limit the number of items scanned), and --order (to control sorting order, asc/desc).

Use the --recursive flag to traverse subdirectories.

Examples:
  syst scanpath                    # Scan current directory
  syst scanpath /path/to/dir       # Scan specific directory  
  syst scanpath . --recursive      # Scan current directory recursively
  syst scanpath /usr --limit 100   # Scan /usr but limit to 100 items
`,
		Args: cobra.MaximumNArgs(1), // Accept 0 or 1 arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to current directory if no path provided
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return scan.ScanDirectoryTUI(path, limit, sortBy, order, filter, recursive)
		},
	}

	// Add command flags
	cmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of results (0 = unlimited)")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Column to sort by (name, size, created, modified, owner, permissions)")
	cmd.Flags().StringVarP(&order, "order", "o", "asc", "Sort order: asc or desc")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter results (e.g. 'size <10MB', 'created >2022-01-01')")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively traverse subdirectories")

	return cmd
}
