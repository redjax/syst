//go:build windows

package pingService

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func runICMPPing(opts Options) error {
	fmt.Printf("Starting ICMP pings to %s...\n", opts.Target)

	i := 0
	for opts.Count == 0 || i < opts.Count {
		select {
		case <-opts.Ctx.Done():
			fmt.Println("\n[!] Interrupt received, stopping ICMP ping")
			return nil
		default:
		}

		cmd := exec.Command("ping", "-n", "1", opts.Target)
		output, err := cmd.CombinedOutput()
		opts.Stats.Total++

		if err != nil || !(strings.Contains(string(output), "Reply from") || strings.Contains(string(output), "TTL=")) {
			fmt.Printf("[FAIL] Ping to %s failed: %v\n", opts.Target, err)
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
	return defaultHTTPPing(opts)
}
