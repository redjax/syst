package scan

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/redjax/syst/internal/services/pathScanService/tbl"
)

// ScanDirectory scans a path with options and returns a list of files
func ScanDirectory(path string, limit int, sortColumn, sortOrder string, filterString string, recursive bool) error {
	if recursive {
		return scanDirectoryRecursive(path, limit, sortColumn, sortOrder, filterString)
	}
	return scanDirectoryShallow(path, limit, sortColumn, sortOrder, filterString)
}

// scanDirectoryShallow performs the original non-recursive scan
func scanDirectoryShallow(path string, limit int, sortColumn, sortOrder string, filterString string) error {
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

// scanDirectoryRecursive performs recursive directory traversal
func scanDirectoryRecursive(rootPath string, limit int, sortColumn, sortOrder string, filterString string) error {
	var results [][]string
	count := 0

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			return nil
		}

		// Skip .git directories to avoid deep traversal into git internals
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Check if we've hit our limit
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit_reached")
		}

		ctime, owner := getMeta(info, path)
		size := info.Size()
		sizeParsed := tbl.ByteCountIEC(size)

		// Use relative path from the root for display
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			relPath = path
		}

		// Add repo indicator for git repositories
		name := relPath
		if info.IsDir() && isGitRepository(path) {
			name = relPath + " [GIT]"
		}

		row := []string{
			name,
			fmt.Sprintf("%d", size),
			sizeParsed,
			ctime,
			info.ModTime().Format("2006-01-02 15:04:05"),
			owner,
			info.Mode().String(),
		}

		results = append(results, row)
		count++

		return nil
	})

	// Handle early termination due to limit
	if err != nil && err.Error() == "limit_reached" {
		err = nil
	}

	if err != nil {
		return err
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

// isGitRepository checks if a directory is a git repository
func isGitRepository(path string) bool {
	// Check if .git directory exists
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return true
	}

	// Also check using git command for more thorough detection
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	result := strings.TrimSpace(string(output))
	return result == "true"
}
