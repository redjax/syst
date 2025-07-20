package gitservice

import (
	"fmt"
	"os"
	"path/filepath"
)

type RemoteInfo struct {
	Name     string
	FetchURL string
	PushURL  string
}

// RepoInfo contains metadata about a Git repository.
type RepoInfo struct {
	Path          string // Absolute path
	CurrentBranch string
	SizeBytes     int64
	IsRepo        bool
	Remotes       []RemoteInfo
}

// GetRepoInfo returns metadata about the current repo.
func GetRepoInfo() (*RepoInfo, error) {
	// Resolve working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("could not get current directory: %w", err)
	}

	info := &RepoInfo{
		Path: wd,
	}

	ok, err := IsGitRepo()
	if err != nil || !ok {
		info.IsRepo = false
		return info, nil // Still return partial info
	}

	info.IsRepo = true

	// Get current branch
	branch, err := GetCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("could not determine current branch: %w", err)
	}
	info.CurrentBranch = branch

	// Estimate size of .git directory
	gitDir := filepath.Join(wd, ".git")
	size, err := dirSize(gitDir)
	if err != nil {
		// Non-fatal
		fmt.Printf("Warning: could not calculate repo size: %v\n", err)
	} else {
		info.SizeBytes = size
	}

	if info.IsRepo {
		remotes, err := getRemotes()
		if err != nil {
			fmt.Printf("Warning: could not list remotes: %v\n", err)
		} else {
			info.Remotes = remotes
		}
	}

	return info, nil
}

// dirSize recursively calculates the size of a directory in bytes.
func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
