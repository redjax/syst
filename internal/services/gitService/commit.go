package gitservice

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type CommitInfo struct {
	Hash    string
	Author  string
	Date    string
	Message string
}

func getCommitCount(branch string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", branch)
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("could not count commits on branch %s: %w", branch, err)
	}
	countStr := strings.TrimSpace(string(out))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, fmt.Errorf("invalid commit count output: %w", err)
	}
	return count, nil
}

func getLastCommit() (*CommitInfo, error) {
	// Custom pretty format for ease of parsing
	format := "%H%n%an%n%ad%n%B"

	cmd := exec.Command("git", "log", "-1", "--pretty=format:"+format, "--date=iso")

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not get last commit: %w", err)
	}

	lines := strings.SplitN(string(out), "\n", 4)

	if len(lines) < 4 {
		return nil, fmt.Errorf("unexpected output from git log")
	}

	return &CommitInfo{
		Hash:    lines[0],
		Author:  lines[1],
		Date:    lines[2],
		Message: strings.TrimSpace(lines[3]),
	}, nil
}
