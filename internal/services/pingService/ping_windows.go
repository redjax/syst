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

	i := 1

	for opts.Count == 0 || i <= opts.Count {
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

		start := time.Now()

		cmd := exec.Command("ping", "-n", "1", opts.Target)
		output, err := cmd.CombinedOutput()
		latency := time.Since(start)

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
			msg := fmt.Sprintf("[OK] Ping to %s succeeded in %s (#%d)", opts.Target, latency, i)
			fmt.Println(msg)
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}

			opts.Stats.Successes++
			opts.Stats.Latencies = append(opts.Stats.Latencies, latency)
			opts.Stats.TotalLatency += latency

			if opts.Stats.MinLatency == 0 || latency < opts.Stats.MinLatency {
				opts.Stats.MinLatency = latency
			}

			if latency > opts.Stats.MaxLatency {
				opts.Stats.MaxLatency = latency
			}
		}

		i++

		if opts.Count != 0 && i > opts.Count {
			break
		}

		time.Sleep(opts.Sleep)
	}

	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Printf("[INFO] Finished ping to %s", opts.Target)
	}

	// Optional: display latency summary
	// if opts.Stats.Successes > 0 {
	// 	avg := opts.Stats.TotalLatency / time.Duration(opts.Stats.Successes)
	// 	summary := fmt.Sprintf("[STATS] Success: %d | Fail: %d | Min: %s | Max: %s | Avg: %s",
	// 		opts.Stats.Successes,
	// 		opts.Stats.Failures,
	// 		opts.Stats.MinLatency,
	// 		opts.Stats.MaxLatency,
	// 		avg,
	// 	)

	// 	fmt.Println(summary)

	// 	if opts.LogToFile && opts.Logger != nil {
	// 		opts.Logger.Println(summary)
	// 	}
	// }
	// PrintPingSummary(opts)
	PrintPingSummaryTable(opts)

	return nil
}

func runHTTPPing(opts *Options) error {
	return defaultHTTPPing(opts)
}
