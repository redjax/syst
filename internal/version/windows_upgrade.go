//go:build windows
// +build windows

package version

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// RunWindowsSelfUpgrade waits for oldExe to be unlocked, then renames newExe over oldExe and restarts it
func RunWindowsSelfUpgrade(oldExe, newExe string) error {
	const maxRetries = 10
	const retryDelay = 500 * time.Millisecond

	var err error
	for i := 0; i < maxRetries; i++ {
		err = os.Rename(newExe, oldExe)
		if err == nil {
			break
		}
		time.Sleep(retryDelay)
	}
	if err != nil {
		return fmt.Errorf("failed to replace executable after %d retries: %w", maxRetries, err)
	}

	// Restart new executable
	cmd := exec.Command(oldExe)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to restart executable: %w", err)
	}
	return nil
}
