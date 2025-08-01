//go:build !windows
// +build !windows

package version

func RunWindowsSelfUpgrade(oldExe, newExe string) error {
	// no-op on non-Windows platform
	return nil
}
