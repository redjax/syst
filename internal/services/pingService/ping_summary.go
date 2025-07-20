package pingService

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

// Prints log stats as a simple newline
func PrintPingSummary(opts *Options) {
	if opts.Stats == nil || opts.Stats.Successes == 0 {
		return
	}

	avg := opts.Stats.TotalLatency / time.Duration(opts.Stats.Successes)

	summary := fmt.Sprintf(
		"[STATS] Target: %s, Success: %d | Failure: %d | Latency Min: %s | Latency Max: %s | Avg. Latency: %s",
		opts.Target,
		opts.Stats.Successes,
		opts.Stats.Failures,
		opts.Stats.MinLatency,
		opts.Stats.MaxLatency,
		avg,
	)

	fmt.Printf("\n%s\n", summary)

	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Println(summary)
	}

	if opts.LogToFile && opts.LogFilePath != "" {
		fmt.Printf("Log file saved to:\n%s\n", opts.LogFilePath)
	}
}

// Prints ping stats summary as a table
func PrintPingSummaryTable(opts *Options) {
	if opts.Stats == nil {
		return
	}

	fmt.Println("\n--- Ping Summary ---")
	fmt.Printf("Target: %s\n", opts.Target)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "\nSTATISTIC\tVALUE")
	fmt.Fprintf(writer, "Total Pings\t%d\n", opts.Stats.Total)
	fmt.Fprintf(writer, "Successes\t%d\n", opts.Stats.Successes)
	fmt.Fprintf(writer, "Failures\t%d\n", opts.Stats.Failures)

	if opts.Stats.Successes > 0 {
		avg := opts.Stats.TotalLatency / time.Duration(opts.Stats.Successes)
		fmt.Fprintf(writer, "Min Latency\t%v\n", opts.Stats.MinLatency)
		fmt.Fprintf(writer, "Max Latency\t%v\n", opts.Stats.MaxLatency)
		fmt.Fprintf(writer, "Avg Latency\t%v\n", avg)
	} else {
		fmt.Fprintf(writer, "Min Latency\t%s\n", "n/a")
		fmt.Fprintf(writer, "Max Latency\t%s\n", "n/a")
		fmt.Fprintf(writer, "Avg Latency\t%s\n", "n/a")
	}

	fmt.Fprintf(writer, "Sleep Interval\t%v\n", opts.Sleep)
	fmt.Fprintf(writer, "Protocol\t%s\n", protocolLabel(opts))
	writer.Flush()

	if opts.LogToFile && opts.Logger != nil {
		opts.Logger.Println("[STATS]")
		opts.Logger.Printf("Successes:   %d\n", opts.Stats.Successes)
		opts.Logger.Printf("Failures:    %d\n", opts.Stats.Failures)
		if opts.Stats.Successes > 0 {
			avg := opts.Stats.TotalLatency / time.Duration(opts.Stats.Successes)
			opts.Logger.Printf("Min Latency: %v\n", opts.Stats.MinLatency)
			opts.Logger.Printf("Max Latency: %v\n", opts.Stats.MaxLatency)
			opts.Logger.Printf("Avg Latency: %v\n", avg)
		}
	}

	if opts.LogToFile && opts.LogFilePath != "" {
		fmt.Printf("\nLog file saved to:\n%s\n", opts.LogFilePath)
	}
}

func protocolLabel(opts *Options) string {
	if opts.UseHTTP {
		return "HTTP"
	}

	return "ICMP"
}
