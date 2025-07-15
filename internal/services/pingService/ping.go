package pingService

import (
	"strings"
	"time"
)

type Options struct {
	Target    string
	Count     int           // 0 = infinite
	Sleep     time.Duration // e.g. 1s, 2s
	UseHTTP   bool
	LogToFile bool // TODO: to be implemented
}

func RunPing(opts Options) error {
	opts.Target = strings.TrimSpace(opts.Target)
	if opts.Target == "" {
		return ErrEmptyTarget
	}

	if opts.UseHTTP {
		return runHTTPPing(opts)
	}
	return runICMPPing(opts)
}
