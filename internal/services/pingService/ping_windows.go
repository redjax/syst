//go:build windows

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
		countArg = "999999" // ping will stop on network error or ctrl+c

		// On Windows, ping by default waits 1s between pings
	}

	i := 0
	for opts.Count == 0 || i < opts.Count {
		i++
		cmd := exec.Command("ping", "-n", countArg, opts.Target)
		output, err := cmd.CombinedOutput()

		// Windows ping output check
		if err != nil || !(strings.Contains(string(output), "Reply from") || strings.Contains(string(output), "TTL=")) {
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
	return defaultHTTPPing(opts)
}
