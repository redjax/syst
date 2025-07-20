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
	CommitCount   int
	LastCommit    *CommitInfo
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

	// Count number of commits on current branch
	commitCount, err := getCommitCount(branch)
	if err != nil {
		fmt.Printf("Warning: could not get commit count: %v\n", err)
	} else {
		info.CommitCount = commitCount
	}

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

		lastCommit, err := getLastCommit()
		if err != nil {
			fmt.Printf("Warning: could not get last commit: %v\n", err)
		} else {
			info.LastCommit = lastCommit
		}
	}

	return info, nil
}

func PrintRepoInfo() error {
	info, err := GetRepoInfo()
	if err != nil {
		return err
	}

	// Repo Path
	fmt.Printf("%-18s %s\n", "Repository Path:", info.Path)

	// Repo Size as human readable
	fmt.Printf("%-18s %s\n", "Repo Size:", BytesToHumanReadable(uint64(info.SizeBytes)))

	if info.IsRepo {
		fmt.Printf("%-18s %s\n", "Current Branch:", info.CurrentBranch)
		fmt.Printf("%-18s %d\n", "Total Commits:", info.CommitCount)

		if info.LastCommit != nil {
			fmt.Println("Last Commit:")
			fmt.Printf("  %-10s %s\n", "Hash:", info.LastCommit.Hash)
			fmt.Printf("  %-10s %s\n", "Author:", info.LastCommit.Author)
			fmt.Printf("  %-10s %s\n", "Date:", info.LastCommit.Date)
		}

		if len(info.Remotes) == 0 {
			fmt.Println("Remotes:           No remotes found")
		} else {
			fmt.Println("Remotes:")
			for _, remote := range info.Remotes {
				fmt.Printf("  +%s:\n", remote.Name)
				fmt.Printf("    %-6s %s\n", "Fetch:", remote.FetchURL)
				fmt.Printf("    %-6s %s\n", "Push:", remote.PushURL)
			}
		}
	}

	return nil
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
