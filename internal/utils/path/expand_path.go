package path

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExpandPath(p string) (string, error) {
	if len(p) == 0 {
		return "", fmt.Errorf("empty path")
	}

	if p[:2] == "~/" || p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}

	return p, nil
}
