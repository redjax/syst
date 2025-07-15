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
		i++
		start := time.Now()

		req, err := http.NewRequest("HEAD", url, nil)
		if err != nil {
			fmt.Printf("[FAIL] HEAD request failed to %s: %v\n", url, err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("[FAIL] HEAD request to %s failed: %v\n", url, err)
		} else {
			fmt.Printf("[OK] HTTP HEAD to %s [%d] in %s\n", url, resp.StatusCode, time.Since(start))
			resp.Body.Close()
		}

		if opts.Count != 0 && i >= opts.Count {
			break
		}
		time.Sleep(opts.Sleep)
	}

	return nil
}
