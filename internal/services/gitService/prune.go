package gitservice

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/redjax/syst/internal/services/platformService/capabilities"
)

// PruneBranches deletes local branches that were deleted on the remote.
// If --confirm is passed, prompt before each deletion.
func PruneBranches(mainBranch string, confirm bool, force bool, dryRun bool) error {
	// Check if git is installed
	path, err := capabilities.Which("git")
	if err != nil || path == "" {
		fmt.Println("ERROR: git is not installed or not found in PATH.")
		return fmt.Errorf("git not found")
	}

	// Detect current branch
	currentBranchCMD := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	currentBranchOut, err := currentBranchCMD.Output()
	if err != nil {
		return fmt.Errorf("could not detect current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(currentBranchOut))

	// Defer restoring original branch on exit
	defer func() {
		if currentBranch != "" {
			cmd := exec.Command("git", "checkout", currentBranch)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Printf("Warning: failed to switch back to original branch %q: %v\n", currentBranch, err)
			} else {
				fmt.Printf("Switched back to original branch: %s\n", currentBranch)
			}
		}
	}()

	// Checkout main branch
	cmd := exec.Command("git", "checkout", mainBranch)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not checkout branch %s: %w", mainBranch, err)
	}

	// Prune remotes
	cmd = exec.Command("git", "remote", "update", "origin", "--prune")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not prune remote branches: %w", err)
	}

	// Find local branches that are gone on the remote
	out, err := exec.Command("git", "branch", "-vv").Output()
	if err != nil {
		return fmt.Errorf("could not list local branches: %w", err)
	}

	var branchesToDelete []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ": gone]") {
			branchName := strings.Fields(line)[0]
			if branchName != mainBranch && branchName != currentBranch {
				branchesToDelete = append(branchesToDelete, branchName)
			}
		}
	}

	if len(branchesToDelete) == 0 {
		fmt.Println("No local branches found that were deleted on the remote.")
		return nil
	}

	deleted := []string{}
	skipped := []string{}

	// If dry-run: only display what would be deleted
	if dryRun {
		fmt.Printf("Dry run — branches that would be deleted:\n")
		for _, branch := range branchesToDelete {
			fmt.Printf("  - %s\n", branch)
		}
		fmt.Printf("\nTotal: %d branches\n", len(branchesToDelete))
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	for _, branch := range branchesToDelete {
		shouldDelete := true
		if confirm {
			fmt.Printf("Delete branch %s? [y/N]: ", branch)
			answer, err := reader.ReadString('\n')
			if err != nil || (strings.ToLower(strings.TrimSpace(answer)) != "y" && strings.ToLower(strings.TrimSpace(answer)) != "yes") {
				fmt.Printf("Skipping %s\n", branch)
				shouldDelete = false
				skipped = append(skipped, branch)
			}
		}

		if shouldDelete {
			del := exec.Command("git", "branch", "-d", branch)
			del.Stdin, del.Stdout, del.Stderr = os.Stdin, os.Stdout, os.Stderr
			if err := del.Run(); err != nil {
				fmt.Printf("Failed to delete %s: %v\n", branch, err)
			} else {
				deleted = append(deleted, branch)
			}
		}
	}

	// Print results
	fmt.Printf("\nDeleted %d branch(es):\n", len(deleted))
	for _, b := range deleted {
		fmt.Printf("  - %s\n", b)
	}

	if len(skipped) > 0 {
		fmt.Printf("\nSkipped %d branch(es):\n", len(skipped))
		for _, b := range skipped {
			fmt.Printf("  - %s\n", b)
		}
	}

	return nil
}
