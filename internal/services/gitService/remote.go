package gitservice

import (
	"os"
	"os/exec"
)

func pruneRemotes() error {
	cmd := exec.Command("git", "remote", "update", "origin", "--prune")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}
