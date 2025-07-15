//go:build linux

package pingService

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func runICMPPing(opts Options) error {
	countArg := fmt.Sprintf("%d", opts.Count)
	if opts.Count == 0 {
		countArg = "999999" // no infinite in ping; pick a large arbitrary number
	}

	i := 0
	for opts.Count == 0 || i < opts.Count {
		i++
		cmd := exec.Command("ping", "-c", countArg, opts.Target)
		output, err := cmd.CombinedOutput()

		if err != nil || !strings.Contains(string(output), "bytes from") {
			fmt.Printf("[FAIL] Ping to %s failed: %v\n", opts.Target, err)
		} else {
			fmt.Printf("[OK] ICMP ping to %s successful\n", opts.Target)
		}

		if opts.Count != 0 && i >= opts.Count {
			break
		}
		time.Sleep(opts.Sleep)
	}

	return nil
}

func runHTTPPing(opts Options) error {
	// Shared code can go here, or you can delegate to a non-OS-specific file
	return defaultHTTPPing(opts)
}
