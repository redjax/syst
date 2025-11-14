package gitservice

import (
	"fmt"
)

func CloneNoCheckout(url, output string) error {
	if !CheckGitInstalled() {
		fmt.Printf("Error: git is not installed")
		return ErrGitNotInstalled
	}

	cmd := execCommand("git", "clone", "--no-checkout", url, output)

	return cmd.Run()
}
