package pingService

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redjax/syst/internal/utils/spinner"
)

func defaultHTTPPing(opts *Options) error {
	url := opts.Target
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	i := 1

	// Prepare & defer spinner
	stopSpinner := spinner.StartSpinner("")
	defer stopSpinner()

	for opts.Count == 0 || i <= opts.Count {
		select {
		case <-opts.Ctx.Done():
			msg := "\n[!] Interrupt received, stopping HTTP ping"
			fmt.Println(msg)
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}
			return nil
		default:
		}

		opts.Stats.Total++

		req, err := http.NewRequestWithContext(opts.Ctx, http.MethodHead, url, nil)

		var msg string

		if err != nil {
			// STOP spinner temporarily before printing result
			stopSpinner() // stop & clear spinner

			msg = fmt.Sprintf("[FAILURE] Request to %s failed to build: %v (#%d)", url, err, i)

			fmt.Println(msg)

			// RESTART spinner
			stopSpinner = spinner.StartSpinner("")

			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}

			opts.Stats.Failures++
		} else {
			start := time.Now()
			resp, err := client.Do(req)

			if err != nil {
				// STOP spinner temporarily before printing result
				stopSpinner() // stop & clear spinner

				msg = fmt.Sprintf("[%s] HTTP HEAD request to %s failed: %v (#%d)", resp.Status, url, err, i)
				fmt.Println(msg)

				// RESTART spinner
				stopSpinner = spinner.StartSpinner("")

				if opts.LogToFile && opts.Logger != nil {
					opts.Logger.Println(msg)
				}

				opts.Stats.Failures++
			} else {
				latency := time.Since(start)

				// STOP spinner temporarily before printing result
				stopSpinner() // stop & clear spinner

				msg = fmt.Sprintf("[%s] HTTP HEAD to %s in %s (#%d)",
					resp.Status, url, latency, i)

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

				resp.Body.Close()
			}
		}

		i++
		if opts.Count != 0 && i > opts.Count {
			break
		}

		time.Sleep(opts.Sleep)
	}

	stopSpinner() // Stop the spinner

	// Print latency summary after finishing
	PrettyPrintPingSummaryTable(opts)

	return nil
}
