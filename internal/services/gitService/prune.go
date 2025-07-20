package gitservice

import (
	"bufio"
	"errors"
	"fmt"
	"os"
)

// PruneBranches deletes local branches that were deleted on the remote.
// If --confirm is passed, prompt before each deletion.
func PruneBranches(mainBranch string, confirm bool, force bool, dryRun bool) error {
	ok, err := IsGitRepo()
	if errors.Is(err, ErrNotGitRepo) || !ok {
		return ErrNotGitRepo
	}

	if !CheckGitInstalled() {
		fmt.Printf("Error: git is not installed")
		return ErrGitNotInstalled
	}

	currentBranch, err := GetCurrentBranch()
	if err != nil {
		return err
	}

	var deleted []string
	var skipped []string

	defer func() {
		_ = checkoutBranch(currentBranch)
		printPruneResults(deleted, skipped, dryRun)
	}()

	if err := checkoutBranch(mainBranch); err != nil {
		return fmt.Errorf("could not checkout branch %s: %w", mainBranch, err)
	}

	if err := pruneRemotes(); err != nil {
		return err
	}

	branchesToDelete, err := getBranchesToDelete(mainBranch, currentBranch)
	if err != nil {
		return err
	}

	if len(branchesToDelete) == 0 {
		return nil
	}

	if dryRun {
		deleted = branchesToDelete
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	for _, branch := range branchesToDelete {
		if confirm {
			if !promptUser(reader, branch) {
				fmt.Printf("Skipping %s\n", branch)
				skipped = append(skipped, branch)
				continue
			}
		}

		if err := deleteBranch(branch, force); err != nil {
			fmt.Printf("Failed to delete %s: %v\n", branch, err)
		} else {
			deleted = append(deleted, branch)
		}
	}

	return nil
}

func printPruneResults(deleted, skipped []string, dryRun bool) {
	if dryRun {
		fmt.Println("\nDry run – branches that would be deleted:")
		for _, b := range deleted {
			fmt.Printf("  - %s\n", b)
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
}
