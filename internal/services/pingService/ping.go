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
}

func RunPing(opts *Options) error {
	opts.Target = strings.TrimSpace(opts.Target)
	if opts.Target == "" {
		return ErrEmptyTarget
	}

	// Setup logging if requested
	if opts.LogToFile {
		filePath := createLogFilePath(opts.Target)
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}

		// Only write to the file — not stdout
		logger := log.New(f, "", log.LstdFlags)
		opts.Logger = logger
		opts.LogFilePath = filePath

		// You can manually log as the first entry:
		logger.Printf("Logging started for ping to %s", opts.Target)
	}

	if opts.UseHTTP {
		return runHTTPPing(opts)
	}
	return runICMPPing(opts)
}

func createLogFilePath(target string) string {
	// Sanitize target for filename
	safeTarget := strings.ReplaceAll(target, ":", "-")
	safeTarget = strings.ReplaceAll(safeTarget, "/", "_")
	dateStr := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("%s_pingo_%s.log", dateStr, safeTarget)
	return filepath.Join(os.TempDir(), fileName)
}
