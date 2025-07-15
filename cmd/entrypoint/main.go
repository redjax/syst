package main

import (
	"github.com/redjax/syst/internal/version"

	// Import the cmd directory with root.go
	"github.com/redjax/syst/cmd"
)

func main() {
	// Check if an update is needed
	version.TrySelfUpgrade()

	// Call the root command
	cmd.Execute()
}
