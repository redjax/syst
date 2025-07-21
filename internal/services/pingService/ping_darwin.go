//go:build darwin

package pingService

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/redjax/syst/internal/utils/spinner"
)

func runICMPPing(opts *Options) error {
	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Printf("[INFO] Starting ICMP ping to %s", opts.Target)
	}

	i := 1

	// Prepare & defer spinner
	stopSpinner := spinner.StartSpinner("")
	defer stopSpinner()

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

		cmd := exec.Command("ping", "-c", "1", "-W", "1", opts.Target)
		// Start ping in a new process group on UNIX-like systems
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		done := make(chan struct{})
		go func() {
			select {
			case <-opts.Ctx.Done():
				if cmd.Process != nil {
					_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				}
			case <-done:
			}
		}()

		output, err := cmd.CombinedOutput()
		latency := time.Since(start)

		opts.Stats.Total++

		if err != nil || !strings.Contains(string(output), "bytes from") {
			// STOP spinner temporarily before printing result
			stopSpinner() // stop & clear spinner

			msg := fmt.Sprintf("[FAIL] Ping to %s failed: %v (#%d)", opts.Target, err, i)

			fmt.Println(msg)

			// RESTART spinner
			stopSpinner = spinner.StartSpinner("")

			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}

			opts.Stats.Failures++
		} else {
			// STOP spinner temporarily before printing result
			stopSpinner() // stop & clear spinner

			msg := fmt.Sprintf("[OK] Ping to %s succeeded in %s (#%d)", opts.Target, latency, i)

			fmt.Println(msg)

			// RESTART spinner
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

	// Stop spinner
	stopSpinner()

	// Print stats summary at the end
	PrettyPrintPingSummaryTable(opts)

	return nil
}

func runHTTPPing(opts *Options) error {
	return defaultHTTPPing(opts)
}
