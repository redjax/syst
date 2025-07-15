package pingService

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func defaultHTTPPing(opts Options) error {
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
			fmt.Println("\n[!] Interrupt received, stopping HTTP ping")
			return nil
		default:
			// Continue with ping
		}

		opts.Stats.Total++
		req, err := http.NewRequestWithContext(opts.Ctx, http.MethodHead, url, nil)
		if err != nil {
			fmt.Printf("[FAIL] Request to %s failed to build: %v\n", url, err)
			opts.Stats.Failures++
			continue
		}

		start := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[FAIL] HTTP HEAD request to %s failed: %v\n", url, err)
			opts.Stats.Failures++
		} else {
			fmt.Printf("[OK] HTTP HEAD to %s [%d] in %s\n", url, resp.StatusCode, time.Since(start))
			opts.Stats.Successes++
			resp.Body.Close()
		}

		i++
		if opts.Count != 0 && i >= opts.Count {
			break
		}
		time.Sleep(opts.Sleep)
	}

	return nil
}
