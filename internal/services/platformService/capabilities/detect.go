package capabilities

import "os/exec"

// Returns path to a binary, if found (i.e. curl -> /usr/bin/curl)
func Which(binary string) (string, error) {
	return exec.LookPath(binary)
}

// Test if a command is available, i.e. 'curl'
func IsCommandAvailable(binary string) bool {
	_, err := exec.LookPath(binary)

	return err == nil
}
