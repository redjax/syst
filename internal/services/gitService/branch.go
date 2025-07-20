package gitservice

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type BranchSyncStatus struct {
	CurrentBranch  string
	TrackingBranch string
	Ahead          int // commits to push
	Behind         int // commits to pull
	HasUpstream    bool
	Error          error
}

// GetCurrentBranch returns the name of the current Git branch.
func GetCurrentBranch() (string, error) {
	if !CheckGitInstalled() {
		fmt.Printf("Error: git is not installed")
		return "", ErrGitNotInstalled
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("could not determine current branch: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// Git checkout a branch
func checkoutBranch(branch string) error {
	if branch == "" {
		return nil
	}

	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: failed to switch to branch %q: %v\n", branch, err)
		return err
	}

	fmt.Printf("Switched to branch: %s\n", branch)

	return nil
}

func getBranchesToDelete(mainBranch, currentBranch string) ([]string, error) {
	out, err := exec.Command("git", "branch", "-vv").Output()
	if err != nil {
		return nil, fmt.Errorf("could not list local branches: %w", err)
	}

	var toDelete []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ": gone]") {
			branch := strings.Fields(line)[0]

			if branch != mainBranch && branch != currentBranch {
				toDelete = append(toDelete, branch)
			}
		}
	}

	return toDelete, nil
}

func deleteBranch(name string, force bool) error {
	args := []string{"branch"}

	if force {
		args = append(args, "-D", name)
	} else {
		args = append(args, "-d", name)
	}

	cmd := exec.Command("git", args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}

func getBranchSyncStatus(branch string) (*BranchSyncStatus, error) {
	status := &BranchSyncStatus{}

	// Try to resolve upstream reference
	upstreamRef := branch + "@{upstream}"
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", upstreamRef)
	out, err := cmd.Output()
	if err != nil {
		status.HasUpstream = false
		return status, nil // Not fatal
	}

	status.HasUpstream = true
	status.TrackingBranch = strings.TrimSpace(string(out))

	// Run git fetch to update remote refs (non-blocking)
	exec.Command("git", "fetch").Run()

	// Compare local and upstream
	cmd = exec.Command("git", "rev-list", "--left-right", "--count", branch+"..."+status.TrackingBranch)
	out, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not get ahead/behind status: %w", err)
	}
	parts := strings.Fields(string(out))
	if len(parts) == 2 {
		fmt.Sscanf(parts[0], "%d", &status.Behind)
		fmt.Sscanf(parts[1], "%d", &status.Ahead)
	}
	return status, nil
}

func CheckoutBranch(branch string) error {
	if !CheckGitInstalled() {
		fmt.Printf("Error: git is not installed")
		return ErrGitNotInstalled
	}

	ok, err := IsGitRepo()

	if errors.Is(err, ErrNotGitRepo) || !ok {
		return ErrNotGitRepo
	}

	cmd := execCommand("git", "checkout", branch)

	return cmd.Run()
}

func GetBranchSyncStatus() BranchSyncStatus {
	status := BranchSyncStatus{}

	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchBytes, err := cmd.Output()
	if err != nil {
		status.Error = fmt.Errorf("could not determine current branch: %w", err)
		return status
	}

	branch := strings.TrimSpace(string(branchBytes))
	status.CurrentBranch = branch

	// Call internal logic
	internalStatus, err := getBranchSyncStatus(branch)
	if err != nil {
		status.Error = err
		return status
	}

	// Merge internal status results into public one
	status.TrackingBranch = internalStatus.TrackingBranch
	status.HasUpstream = internalStatus.HasUpstream
	status.Ahead = internalStatus.Ahead
	status.Behind = internalStatus.Behind

	return status
}
