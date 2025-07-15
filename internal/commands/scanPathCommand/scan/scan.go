package scan

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/redjax/syst/internal/commands/scanPathCommand/tbl"
)

func ScanDirectory(path string, limit int, sortColumn, sortOrder string, filterString string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var results [][]string
	count := 0
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ctime, owner := getMeta(info, filepath.Join(path, entry.Name()))
		size := info.Size()
		sizeParsed := tbl.ByteCountIEC(size)

		row := []string{
			info.Name(),
			fmt.Sprintf("%d", size),
			sizeParsed,
			ctime,
			info.ModTime().Format("2006-01-02 15:04:05"),
			owner,
			info.Mode().String(),
		}
		results = append(results, row)
		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	var filterExpr *tbl.FilterExpr
	if filterString != "" {
		filterExpr, err = tbl.ParseFilter(filterString)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid filter: %v\n", err)
		}
	}

	results = tbl.ApplyFilter(results, filterExpr)
	tbl.SortResults(results, sortColumn, sortOrder)
	tbl.PrintScanResultsTable(results)

	return nil
}
