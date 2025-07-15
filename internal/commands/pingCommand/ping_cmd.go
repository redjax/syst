package pingCommand

import (
	"context"
	"fmt"
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

			err := pingService.RunPing(opts)
			if err != nil {
				return err
			}

			// Print ping summary
			fmt.Printf("\n-------------\n")
			fmt.Println("Ping Summary:")
			fmt.Printf("  Total:     %d\n", stats.Total)
			fmt.Printf("  Succeeded: %d\n", stats.Successes)
			fmt.Printf("  Failed:    %d\n", stats.Failures)
			fmt.Printf("-------------\n")

			// Print log file below summary
			if opts.LogToFile && opts.LogFilePath != "" {
				fmt.Printf("Log file saved to: %s\n", opts.LogFilePath)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 3, "Number of pings to send (0 = infinite)")
	cmd.Flags().IntVarP(&sleep, "sleep", "s", 1, "Seconds to sleep between pings")
	cmd.Flags().BoolVar(&useHTTP, "http", false, "Send HTTP HEAD request instead of ICMP ping")
	cmd.Flags().BoolVar(&logToFile, "log-file", false, "Log output to a temp file with date and host/FQDN")

	return cmd
}
