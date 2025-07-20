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
	path, err := capabilities.Which("git")
	if err != nil || path == "" {
		fmt.Println("ERROR: git is not installed or not found in PATH.")
		return fmt.Errorf("git not found")
	}

	currentBranchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	currentBranchOut, err := currentBranchCmd.Output()
	if err != nil {
		return fmt.Errorf("could not detect current branch: %w", err)
	}
	currentBranch := strings.TrimSpace(string(currentBranchOut))

	var deleted []string
	var skipped []string

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

		if dryRun {
			fmt.Println("\nDry run – branches that would be deleted:")
			for _, branch := range deleted {
				fmt.Printf("  - %s\n", branch)
			}
			fmt.Printf("\nTotal: %d branch(es)\n", len(deleted))
			return
		}

		if len(deleted) > 0 {
			fmt.Printf("\nDeleted %d branch(es):\n", len(deleted))
			for _, b := range deleted {
				fmt.Printf("  - %s\n", b)
			}
		} else {
			fmt.Println("\nNo branches were deleted.")
		}

		if len(skipped) > 0 {
			fmt.Printf("\nSkipped %d branch(es):\n", len(skipped))
			for _, b := range skipped {
				fmt.Printf("  - %s\n", b)
			}
		}
	}()

	cmd := exec.Command("git", "checkout", mainBranch)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not checkout branch %s: %w", mainBranch, err)
	}

	cmd = exec.Command("git", "remote", "update", "origin", "--prune")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not prune remote branches: %w", err)
	}

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
		return nil // No branches to delete; defer will still restore branch and print "none deleted"
	}

	deleted = branchesToDelete

	if dryRun {
		return nil // Skip deletion, defer block will handle printing
	}

	reader := bufio.NewReader(os.Stdin)

	for _, branch := range branchesToDelete {
		if confirm {
			fmt.Printf("Delete branch %s? [y/N]: ", branch)
			answer, err := reader.ReadString('\n')
			if err != nil || (strings.ToLower(strings.TrimSpace(answer)) != "y" && strings.ToLower(strings.TrimSpace(answer)) != "yes") {
				fmt.Printf("Skipping %s\n", branch)
				skipped = append(skipped, branch)
				continue
			}
		}

		args := []string{"branch"}
		if force {
			args = append(args, "-D", branch)
		} else {
			args = append(args, "-d", branch)
		}

		del := exec.Command("git", args...)
		del.Stdin, del.Stdout, del.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := del.Run(); err != nil {
			fmt.Printf("Failed to delete %s: %v\n", branch, err)
		} else {
			deleted = append(deleted, branch)
		}
	}

	return nil
}
