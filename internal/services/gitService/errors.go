package gitservice

import (
	"errors"
)

// NotARepoError is returned when path is not a git repository
var ErrorNotAGitRepo = errors.New("path is not a git repository")
