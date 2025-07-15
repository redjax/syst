// scanpath_cmd.go
package scanPathCommand

import (
	"github.com/redjax/syst/internal/commands/scanPathCommand/scan"
	"github.com/spf13/cobra"
)

func NewScanPathCommand() *cobra.Command {
	var (
		path   string
		limit  int
		sortBy string
		order  string
		filter string
	)

	cmd := &cobra.Command{
		Use:   "scanpath",
		Short: "Scan a directory and list items with metadata",
		Long:  "Scan a directory and list files with metadata like size, creation time, permissions, etc.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return scan.ScanDirectory(path, limit, sortBy, order, filter) // ✅ direct call
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", ".", "Directory path to scan")
	cmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of results (0 = unlimited)")
	cmd.Flags().StringVarP(&sortBy, "sort", "s", "name", "Column to sort by (name, size, created, modified, owner, permissions)")
	cmd.Flags().StringVarP(&order, "order", "o", "asc", "Sort order: asc or desc")
	cmd.Flags().StringVarP(&filter, "filter", "f", "", "Filter results (e.g. 'size <10MB', 'created >2022-01-01')")

	return cmd
}
