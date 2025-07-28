package pingCommand

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/redjax/syst/internal/services/pingService"
	"github.com/spf13/cobra"
)

func NewPingCommand() *cobra.Command {
	var (
		count     int
		sleep     int
		useHTTP   bool
		logToFile bool
	)

	cmd := &cobra.Command{
		Use:   "ping [target]",
		Short: "Ping a host (IP, hostname, or FQDN) using ICMP or HTTP HEAD",
		Long: `Pings a target host using ICMP or an HTTP HEAD request. 
Can be used to detect host availability and latency.

Supports flags like count, delay between pings, and file-based logging.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			stats := &pingService.PingStats{}

			opts := pingService.Options{
				Target:    target,
				Count:     count,
				Sleep:     time.Duration(sleep) * time.Second,
				UseHTTP:   useHTTP,
				LogToFile: logToFile,
				Ctx:       ctx,
				Stats:     stats,
			}

			err := pingService.RunPing(&opts)
			if err != nil {
				// Detect DNS lookup failure or similar errors; print failure but suppress help.
				if isDNSError(err) {
					cmd.Printf("[FAIL] %s: %v\n", target, err)
					return nil
				}

				// For other errors, print failure and suppress help.
				cmd.Printf("[FAIL] %s: %v\n", target, err)
				return nil
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 3, "Number of pings to send (0 = infinite)")
	cmd.Flags().IntVarP(&sleep, "sleep", "s", 1, "Seconds to sleep between pings")
	cmd.Flags().BoolVar(&useHTTP, "http", false, "Use HTTP HEAD request instead of ICMP ping")
	cmd.Flags().BoolVar(&logToFile, "log-file", false, "Log output to a temp file with date and host")

	return cmd
}
