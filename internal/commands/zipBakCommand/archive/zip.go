package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redjax/syst/internal/commands/zipBakCommand/config"
)

// shouldIgnore determines if a path should be ignored during zip backup
func shouldIgnore(path, src string, ignore []string) bool {
	rel, err := filepath.Rel(src, path)
	if err != nil {
		return false
	}

	for _, pattern := range ignore {
		if rel == pattern || filepath.Base(path) == pattern || strings.HasPrefix(rel, pattern+string(os.PathSeparator)) {
			return true
		}
	}

	return false
}

// getTimestamp returns a formatted timestamp
func getTimestamp() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

// makeZip creates a .zip archive of a path
func makeZip(src, dest string, ignore []string, dryRun bool) error {
	// Get absolute path for source
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}

	// Ensure dest ends with .zip (case-insensitive)
	zipPath := dest
	if !strings.HasSuffix(strings.ToLower(zipPath), ".zip") {
		zipPath += ".zip"
	}

	// Check if dry run
	if dryRun {
		fmt.Printf("[DRY RUN] Would create zip archive at '%s'\n", zipPath)
		return nil
	}

	// Create .zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Write data to .zip
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	return filepath.Walk(srcAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == srcAbs {
			return nil
		}

		if shouldIgnore(path, srcAbs, ignore) {
			return nil
		}

		relPath, err := filepath.Rel(srcAbs, path)
		if err != nil {
			return err
		}

		arcname := relPath
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := zipWriter.Create(arcname)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)

		return err
	})
}

// StartBackup is exposed and calls the makeZip command
func StartBackup(cfg *config.BackupConfig) error {
	// Load src from config
	src := cfg.BackupSrc
	// Get a timestamp
	ts := getTimestamp()
	// Create path to destination
	dest := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_%s", ts, cfg.BackupName))

	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source directory does not exist: %v", err)
	}

	// Create destination parent path if not exists
	if _, err := os.Stat(cfg.OutputDir); os.IsNotExist(err) {
		if cfg.DryRun {
			fmt.Printf("[DRY RUN] Would create output directory '%s'\n", cfg.OutputDir)
		} else {
			// 0750 = owner rwx, group rx, others none
			if err := os.MkdirAll(cfg.OutputDir, 0750); err != nil {
				return fmt.Errorf("failed to create output dir: %v", err)
			}
		}
	}

	// Archive
	err := makeZip(src, dest, cfg.IgnorePaths, cfg.DryRun)
	if err != nil {
		return fmt.Errorf("failed to create zip: %v", err)
	}

	// Cleanup
	if cfg.DoCleanup {
		err := CleanupBackups(cfg.OutputDir, cfg.KeepBackups, cfg.DryRun, dest+".zip")
		if err != nil {
			return fmt.Errorf("cleanup failed: %v", err)
		}
	}

	return nil
}
