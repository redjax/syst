//go:build linux

package pingService

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func runICMPPing(opts Options) error {
	fmt.Printf("Starting ICMP ping to %s (Ctrl+C to stop)...\n", opts.Target)

	i := 0
	for opts.Count == 0 || i < opts.Count {
		select {
		case <-opts.Ctx.Done():
			fmt.Println("\n[!] Interrupt received, stopping.")
			return nil
		default:
			// continue
		}

		cmd := exec.Command("ping", "-c", "1", opts.Target)
		output, err := cmd.CombinedOutput()
		opts.Stats.Total++

		if err != nil || !strings.Contains(string(output), "bytes from") {
			fmt.Printf("[FAIL] Ping to %s failed\n", opts.Target)
			opts.Stats.Failures++
		} else {
			fmt.Printf("[OK] Ping to %s succeeded\n", opts.Target)
			opts.Stats.Successes++
		}

		i++
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
