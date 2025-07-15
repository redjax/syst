//go:build linux

package pingService

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func runICMPPing(opts Options) error {
	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Printf("[INFO] Starting ICMP ping to %s", opts.Target)
	}

	i := 0
	for opts.Count == 0 || i < opts.Count {
		select {
		case <-opts.Ctx.Done():
			msg := "\n[!] Interrupt received, stopping ICMP ping"
			fmt.Println(msg)
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}
			return nil
		default:
		}

		cmd := exec.Command("ping", "-c", "1", opts.Target)
		output, err := cmd.CombinedOutput()
		opts.Stats.Total++

		if err != nil || !strings.Contains(string(output), "bytes from") {
			msg := fmt.Sprintf("[FAIL] Ping to %s failed: %v", opts.Target, err)
			fmt.Println(msg)
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}
			opts.Stats.Failures++
		} else {
			msg := fmt.Sprintf("[OK] Ping to %s succeeded", opts.Target)
			fmt.Println(msg)
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}
			opts.Stats.Successes++
		}

		i++
		if opts.Count != 0 && i >= opts.Count {
			break
		}
		time.Sleep(opts.Sleep)
	}

	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Printf("[INFO] Finished ping to %s", opts.Target)
	}

	return nil
}

func runHTTPPing(opts Options) error {
	return defaultHTTPPing(opts)
}
