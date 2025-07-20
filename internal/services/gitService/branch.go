package gitservice

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// GetCurrentBranch returns the name of the current Git branch.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not detect current branch: %w", err)
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
