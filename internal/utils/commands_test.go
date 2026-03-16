package utils

import (
	"runtime"
	"testing"
)

func TestIsCommandAvailable(t *testing.T) {
	// A command that exists on every platform
	var knownCmd string
	if runtime.GOOS == "windows" {
		knownCmd = "cmd"
	} else {
		knownCmd = "sh"
	}

	if !IsCommandAvailable(knownCmd) {
		t.Errorf("expected %q to be available", knownCmd)
	}

	if IsCommandAvailable("this-command-should-not-exist-xyz-42") {
		t.Error("expected nonexistent command to be unavailable")
	}
}
