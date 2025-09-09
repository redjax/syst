package statusService

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// StatusOptions configures the git status display
type StatusOptions struct {
	ShowAll    bool
	ShowColors bool
}

// FileStatus represents the status of a file in the repository
type FileStatus struct {
	Path    string
	Status  string // "tracked", "untracked", "modified", "staged", "deleted"
	IsDir   bool
	Size    int64
	ModTime string
}

// StatusInfo contains all file status information
type StatusInfo struct {
	CleanFiles     []FileStatus // Tracked files with no changes (show as normal text)
	UntrackedFiles []FileStatus // New files not tracked by git
	ModifiedFiles  []FileStatus // Files with changes in working directory
	StagedFiles    []FileStatus // Files staged for commit
	DeletedFiles   []FileStatus // Files deleted from working directory
}

// RunGitStatus displays the git status with tracked/untracked indicators
func RunGitStatus(opts StatusOptions) error {
	// Check if we're in a git repository
	if !isGitRepository() {
		return fmt.Errorf("not a git repository")
	}

	statusInfo, err := gatherStatusInfo()
	if err != nil {
		return fmt.Errorf("failed to gather git status: %w", err)
	}

	printStatusInfo(statusInfo, opts)
	return nil
}

// isGitRepository checks if we're in a git repository
func isGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}

// gatherStatusInfo collects all file status information
func gatherStatusInfo() (*StatusInfo, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	// Get git status
	status, err := worktree.Status()
	if err != nil {
		return nil, err
	}

	statusInfo := &StatusInfo{
		CleanFiles:     []FileStatus{},
		UntrackedFiles: []FileStatus{},
		ModifiedFiles:  []FileStatus{},
		StagedFiles:    []FileStatus{},
		DeletedFiles:   []FileStatus{},
	}

	// Process git status
	for filePath, fileStatus := range status {
		info, err := os.Stat(filePath)
		var size int64
		var modTime string
		var isDir bool

		if err == nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02 15:04")
			isDir = info.IsDir()
		}

		fs := FileStatus{
			Path:    filePath,
			IsDir:   isDir,
			Size:    size,
			ModTime: modTime,
		}

		// Categorize by status
		if fileStatus.Staging != git.Unmodified {
			fs.Status = "staged"
			statusInfo.StagedFiles = append(statusInfo.StagedFiles, fs)
		}

		if fileStatus.Worktree == git.Modified {
			fs.Status = "modified"
			statusInfo.ModifiedFiles = append(statusInfo.ModifiedFiles, fs)
		}

		if fileStatus.Worktree == git.Untracked {
			fs.Status = "untracked"
			statusInfo.UntrackedFiles = append(statusInfo.UntrackedFiles, fs)
		}

		if fileStatus.Worktree == git.Deleted {
			fs.Status = "deleted"
			statusInfo.DeletedFiles = append(statusInfo.DeletedFiles, fs)
		}
	}

	// Also get all tracked files
	cleanFiles, err := getAllTrackedFiles(repo)
	if err == nil {
		statusInfo.CleanFiles = cleanFiles
	}

	return statusInfo, nil
}

// getAllTrackedFiles gets all files tracked by git
func getAllTrackedFiles(repo *git.Repository) ([]FileStatus, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	var trackedFiles []FileStatus

	err = tree.Files().ForEach(func(file *object.File) error {
		info, err := os.Stat(file.Name)
		var size int64
		var modTime string
		var isDir bool

		if err == nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02 15:04")
			isDir = info.IsDir()
		} else {
			// File might have been deleted
			size = file.Size
		}

		trackedFiles = append(trackedFiles, FileStatus{
			Path:    file.Name,
			Status:  "tracked",
			IsDir:   isDir,
			Size:    size,
			ModTime: modTime,
		})
		return nil
	})

	return trackedFiles, err
}

// printStatusInfo displays the status information
func printStatusInfo(info *StatusInfo, opts StatusOptions) {
	fmt.Println("📋 Git Repository Status")
	fmt.Println(strings.Repeat("=", 50))

	// Only show summary for files that have changes
	totalChanges := len(info.ModifiedFiles) + len(info.StagedFiles) + len(info.UntrackedFiles) + len(info.DeletedFiles)

	if totalChanges == 0 {
		fmt.Println("✅ Working directory clean - no changes to commit")
		if opts.ShowAll && len(info.CleanFiles) > 0 {
			fmt.Printf("\n📁 Repository contains %d tracked files\n", len(info.CleanFiles))
		}
		return
	}

	// Print summary of changes only
	if len(info.StagedFiles) > 0 {
		fmt.Printf("📦 Staged files: %d\n", len(info.StagedFiles))
	}
	if len(info.ModifiedFiles) > 0 {
		fmt.Printf("✏️  Modified files: %d\n", len(info.ModifiedFiles))
	}
	if len(info.DeletedFiles) > 0 {
		fmt.Printf("🗑️  Deleted files: %d\n", len(info.DeletedFiles))
	}
	if len(info.UntrackedFiles) > 0 {
		fmt.Printf("❓ Untracked files: %d\n", len(info.UntrackedFiles))
	}
	fmt.Println()

	// Show staged files first (green)
	if len(info.StagedFiles) > 0 {
		fmt.Println("📦 Staged Files (ready to commit):")
		printFileList(info.StagedFiles, "🟢", "\033[32m", opts.ShowColors)
		fmt.Println()
	}

	// Show modified files (yellow)
	if len(info.ModifiedFiles) > 0 {
		fmt.Println("✏️  Modified Files:")
		printFileList(info.ModifiedFiles, "🟡", "\033[33m", opts.ShowColors)
		fmt.Println()
	}

	// Show deleted files (red)
	if len(info.DeletedFiles) > 0 {
		fmt.Println("🗑️  Deleted Files:")
		printFileList(info.DeletedFiles, "🔴", "\033[31m", opts.ShowColors)
		fmt.Println()
	}

	// Show untracked files (gray)
	if len(info.UntrackedFiles) > 0 {
		fmt.Println("❓ Untracked Files:")
		printFileList(info.UntrackedFiles, "⚫", "\033[37m", opts.ShowColors)
		fmt.Println()
	}

	// Show all tracked files if requested (normal text, no special indicators)
	if opts.ShowAll && len(info.CleanFiles) > 0 {
		fmt.Println("📁 All Tracked Files (no changes):")
		printCleanFileList(info.CleanFiles)
	}
}

// printFileList prints a list of files with indicators and colors
func printFileList(files []FileStatus, indicator string, colorCode string, useColors bool) {
	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	resetCode := ""
	if useColors {
		resetCode = "\033[0m"
	} else {
		colorCode = ""
	}

	for _, file := range files {
		dirIndicator := ""
		if file.IsDir {
			dirIndicator = "/"
		}

		sizeStr := ""
		if file.Size > 0 {
			sizeStr = fmt.Sprintf(" (%s)", formatSize(file.Size))
		}

		timeStr := ""
		if file.ModTime != "" {
			timeStr = " - " + file.ModTime
		}

		fmt.Printf("  %s%s %s%s%s%s%s\n",
			colorCode, indicator, file.Path, dirIndicator, sizeStr, resetCode, timeStr)
	}
}

// printCleanFileList prints tracked files with no changes (normal text)
func printCleanFileList(files []FileStatus) {
	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	for _, file := range files {
		dirIndicator := ""
		if file.IsDir {
			dirIndicator = "/"
		}

		sizeStr := ""
		if file.Size > 0 {
			sizeStr = fmt.Sprintf(" (%s)", formatSize(file.Size))
		}

		timeStr := ""
		if file.ModTime != "" {
			timeStr = " - " + file.ModTime
		}

		fmt.Printf("  %s%s%s%s\n", file.Path, dirIndicator, sizeStr, timeStr)
	}
}

// formatSize formats file size in human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
