//go:build windows
// +build windows

package version

// This file is intentionally empty.
// Windows binary replacement is handled by replaceWindows() in upgrade.go.
// Windows allows renaming a running executable, so we rename the current
// binary to .old, copy the new one in, and clean up.
