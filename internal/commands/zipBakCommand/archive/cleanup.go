package archive

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// fileInfo holds the path and modTime of a file.
type fileInfo struct {
	path    string
	modTime int64
}

// CleanupBackups deletes the oldest zip files in `dir`, keeping only the most recent `keep`.
// It ensures the current backup (excludePath) is never deleted.
func CleanupBackups(dir string, keep int, dryRun bool, excludePath string) error {
	// Find all .zip files in path
	matches, err := filepath.Glob(filepath.Join(dir, "*.zip"))
	if err != nil {
		return err
	}

	var allFiles []fileInfo
	for _, path := range matches {
		// Get info about path
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}

		// Append to allFiles array
		allFiles = append(allFiles, fileInfo{path: path, modTime: fi.ModTime().Unix()})
	}

	// Check if there are more values than given --keep value
	if len(allFiles) <= keep {
		fmt.Printf("No cleanup needed. There are %d backup file(s), and the keep threshold is %d.\n", len(allFiles), keep)
		return nil
	}

	// Sort by mod time (newest first)
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].modTime > allFiles[j].modTime
	})

	// Build list of files to keep, including the excludePath
	var keepFiles []fileInfo
	for _, fi := range allFiles {
		if fi.path == excludePath || len(keepFiles) < keep {
			keepFiles = append(keepFiles, fi)
		}
	}

	// Build a set of paths to keep
	keepSet := make(map[string]struct{})
	for _, fi := range keepFiles {
		keepSet[fi.path] = struct{}{}
	}

	// Files not in the keepSet should be deleted
	fmt.Println("Will keep:")
	for _, fi := range keepFiles {
		fmt.Printf("  %s\n", fi.path)
	}

	// Remove files
	fmt.Println("Will delete:")
	for _, fi := range allFiles {
		if _, keep := keepSet[fi.path]; keep {
			continue
		}

		fmt.Printf("  %s\n", fi.path)

		if dryRun {
			fmt.Printf("[DRY RUN] Would delete old backup: %s\n", fi.path)
		} else {
			if err := os.Remove(fi.path); err != nil {
				fmt.Printf("Error deleting %s: %v\n", fi.path, err)
			} else {
				fmt.Printf("Deleted old backup: %s\n", fi.path)
			}
		}
	}

	return nil
}
