package pingService

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Options struct {
	Target      string
	Count       int           // 0 = infinite
	Sleep       time.Duration // e.g. 1s, 2s
	UseHTTP     bool
	LogToFile   bool
	LogFilePath string
	Logger      *log.Logger

	Ctx   context.Context
	Stats *PingStats
}

type PingStats struct {
	Successes int
	Failures  int
	Total     int

	// Hold stats about each successful ping's latency
	Latencies []time.Duration
	// Sum of all latencies for calculating average
	TotalLatency time.Duration
	// Best/worst latency
	MinLatency time.Duration
	MaxLatency time.Duration
}

func RunPing(opts *Options) error {
	opts.Target = strings.TrimSpace(opts.Target)
	if opts.Target == "" {
		return fmt.Errorf("ping target cannot be empty")
	}

	// Setup logging if needed
	if opts.LogToFile {
		filePath := createLogFilePath(opts.Target)
		// Use 0600 permissions for log files (owner read/write only)
		// #nosec G304 - CLI tool creates log files at user-specified paths by design
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}

		logger := log.New(f, "", log.LstdFlags)
		opts.Logger = logger
		opts.LogFilePath = filePath
		logger.Printf("Logging started for ping to %s\n", opts.Target)
	}

	// Show info string
	numPings := "inf"
	if opts.Count != 0 {
		numPings = fmt.Sprintf("%d", opts.Count)
	}

	fmt.Printf("> Pinging %s [ # pings: %s | sleep: %v | protocol: %s ]\n",
		opts.Target, numPings, opts.Sleep, protocolLabel(opts))

	// Delegate ping logic
	if opts.UseHTTP {
		return defaultHTTPPing(opts) // keep your existing HTTP ping
	}

	return runICMPPing(opts) // this is now portable Go
}

func createLogFilePath(target string) string {
	// Sanitize target for filename
	safeTarget := strings.ReplaceAll(target, ":", "-")
	safeTarget = strings.ReplaceAll(safeTarget, "/", "_")
	dateStr := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("%s_pingo_%s.log", dateStr, safeTarget)

	return filepath.Join(os.TempDir(), fileName)
}

func sleepOrCancel(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
