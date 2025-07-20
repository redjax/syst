package gitservice

import (
	"errors"
	"fmt"
)

func CloneNoCheckout(url, output string) error {
	ok, err := IsGitRepo()
	if errors.Is(err, ErrNotGitRepo) || !ok {
		return ErrNotGitRepo
	}

	if !CheckGitInstalled() {
		fmt.Printf("Error: git is not installed")
		return ErrGitNotInstalled
	}

	cmd := execCommand("git", "clone", "--no-checkout", url, output)

	return cmd.Run()
}
