package gitservice

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/redjax/syst/internal/services/platformService/capabilities"
)

func ensureGitInstalled() error {
	path, err := capabilities.Which("git")

	if err != nil || path == "" {
		fmt.Println("ERROR: git is not installed or not found in PATH.")
		return fmt.Errorf("git not found")
	}

	return nil
}

// Prompt user for confirmation
func promptUser(reader *bufio.Reader, branch string) bool {
	fmt.Printf("Delete branch %s? [y/N]: ", branch)

	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	answer = strings.ToLower(strings.TrimSpace(answer))

	return answer == "y" || answer == "yes"
}

// IsGitRepo checks if the current working directory is part of a Git repository.
func IsGitRepo() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	output, err := cmd.Output()
	if err != nil {
		return false, ErrNotGitRepo
	}

	result := strings.TrimSpace(string(output))

	return result == "true", nil
}

func CheckGitInstalled() bool {
	_, err := capabilities.Which("git")
	return err == nil
}

// execCommand allows mocking for tests later if needed
var execCommand = func(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}
