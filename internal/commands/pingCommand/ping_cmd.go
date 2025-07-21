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
		// Number of times to ping target
		count int
		// Number of second(s) to sleep between pings
		sleep int
		// Send HTTP HEAD request instead of ICMP ping
		useHTTP bool
		// When present, output ping logs to a file
		logToFile bool
	)

	cmd := &cobra.Command{
		Use:   "ping [target]",
		Short: "Ping a host (IP, hostname, or FQDN) using ICMP or HTTP HEAD",
		Long: `Pings a target host using ICMP or an HTTP HEAD request. 
Can be used to detect host availability and latency.

Supports flags like count, delay between pings, and file-based logging.`,
		Args: cobra.ExactArgs(1), // Requires exactly one argument (target)
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

			// Start ping
			err := pingService.RunPing(&opts)
			if err != nil {
				return err
			}

			return nil
		},
	}

	// Add ping command flags
	cmd.Flags().IntVarP(&count, "count", "c", 3, "Number of pings to send (0 = infinite)")
	// Main flag with primary shorthand -s
	cmd.Flags().IntVarP(&sleep, "sleep", "s", 1, "Seconds to sleep between pings (hidden alias: -t)")

	// Secondary alias flag -t (hidden, same variable)
	cmd.Flags().IntVarP(&sleep, "sleep-alias", "t", 1, "")
	_ = cmd.Flags().MarkHidden("sleep-alias")
	cmd.Flags().BoolVar(&useHTTP, "http", false, "Send HTTP HEAD request instead of ICMP ping")
	cmd.Flags().BoolVar(&logToFile, "log-file", false, "Log output to a temp file with date and host/FQDN")

	return cmd
}
