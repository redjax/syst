package pingService

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func defaultHTTPPing(opts *Options) error {
	url := opts.Target
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	i := 0
	for opts.Count == 0 || i < opts.Count {
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
			msg = fmt.Sprintf("[FAIL] Request to %s failed to build: %v (#%d)", url, err, i)

			fmt.Println(msg)

			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}

			opts.Stats.Failures++
		} else {
			start := time.Now()

			resp, err := client.Do(req)
			if err != nil {
				msg = fmt.Sprintf("[FAIL] HTTP HEAD request to %s failed: %v (#%d)", url, err, i)

				fmt.Println(msg)

				if opts.LogToFile && opts.Logger != nil {
					opts.Logger.Println(msg)
				}

				opts.Stats.Failures++
			} else {
				msg = fmt.Sprintf("[OK] HTTP HEAD to %s [%d] in %s (#%d)", url, resp.StatusCode, time.Since(start), i)

				fmt.Println(msg)

				if opts.LogToFile && opts.Logger != nil {
					opts.Logger.Println(msg)
				}

				opts.Stats.Successes++

				resp.Body.Close()
			}
		}

		i++
		if opts.Count != 0 && i >= opts.Count {
			break
		}

		time.Sleep(opts.Sleep)
	}

	return nil
}
