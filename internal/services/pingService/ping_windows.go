//go:build windows
// +build windows

package pingService

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/redjax/syst/internal/utils/spinner"
)

func runICMPPing(opts *Options) error {
	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Printf("[INFO] Starting ICMP ping to %s", opts.Target)
	}

	i := 1

	stopSpinner := spinner.StartSpinner("")

	// Defer printing the ping summary and stopping spinner on any exit
	defer func() {
		stopSpinner()
		PrettyPrintPingSummaryTable(opts)
	}()

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

		cmd := exec.CommandContext(opts.Ctx, "ping", "-n", "1", "-w", "1000", opts.Target) // timeout 1000ms
		output, err := cmd.CombinedOutput()
		latency := time.Since(start)

		outputStr := string(output)
		opts.Stats.Total++

		if err != nil || !(strings.Contains(outputStr, "Reply from") || strings.Contains(outputStr, "TTL=")) {
			stopSpinner()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			msg := fmt.Sprintf("[%s] [FAIL] Ping to %s failed: %v (#%d)", timestamp, opts.Target, err, i)
			fmt.Println(msg)
			stopSpinner = spinner.StartSpinner("")
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}
			opts.Stats.Failures++
		} else {
			stopSpinner()
			timestamp := time.Now().Format("2006-01-02 15:04:05")
			msg := fmt.Sprintf("[%s] [OK] Ping to %s succeeded in %s (#%d)", timestamp, opts.Target, latency, i)
			fmt.Println(msg)
			stopSpinner = spinner.StartSpinner("")
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

		if !sleepOrCancel(opts.Ctx, opts.Sleep) {
			stopSpinner() // Stop the spinner
			return nil
		}
	}

	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Printf("[INFO] Finished ping to %s", opts.Target)
	}

	return nil
}
