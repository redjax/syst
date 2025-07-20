package gitservice

import (
	"errors"
	"os"
	"os/exec"
)

func pruneRemotes() error {
	ok, err := IsGitRepo()
	if errors.Is(err, ErrorNotAGitRepo) || !ok {
		return ErrorNotAGitRepo
	}

	cmd := exec.Command("git", "remote", "update", "origin", "--prune")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}
