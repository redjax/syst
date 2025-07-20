//go:build windows

package pingService

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func runICMPPing(opts *Options) error {
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

		cmd := exec.Command("ping", "-n", "1", opts.Target)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		opts.Stats.Total++

		if err != nil || !(strings.Contains(outputStr, "Reply from") || strings.Contains(outputStr, "TTL=")) {
			msg := fmt.Sprintf("[FAIL] Ping to %s failed: %v (#%d)", opts.Target, err, i)

			fmt.Println(msg)

			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}

			opts.Stats.Failures++
		} else {
			msg := fmt.Sprintf("[OK] Ping to %s succeeded (#%d)", opts.Target, i)

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

func runHTTPPing(opts *Options) error {
	return defaultHTTPPing(opts)
}
