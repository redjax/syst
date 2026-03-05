//go:build !windows
// +build !windows

package version

// This file is intentionally empty.
// Windows binary replacement is handled by replaceWindows() in upgrade.go,
// which is only called on runtime.GOOS == "windows".
