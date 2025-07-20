package gitservice

import (
	"errors"
)

// NotARepoError is returned when path is not a git repository
var ErrNotGitRepo = errors.New("path is not a git repository")
var ErrGitNotInstalled = errors.New("git is not installed")
