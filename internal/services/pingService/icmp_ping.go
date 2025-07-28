package pingService

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	probing "github.com/prometheus-community/pro-bing"
	"github.com/redjax/syst/internal/utils/spinner"
)

func runICMPPing(opts *Options) error {
	pinger, err := probing.NewPinger(opts.Target)
	if err != nil {
		return fmt.Errorf("failed to create pinger: %w", err)
	}

	pinger.SetPrivileged(false)
	if opts.Count > 0 {
		pinger.Count = opts.Count
	}

	pinger.Interval = opts.Sleep

	opts.Stats = &PingStats{}

	// Start spinner (like in your HTTP ping)
	stopSpinner := spinner.StartSpinner("")
	defer stopSpinner()

	// Setup Ctrl-C and context cancellation handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-opts.Ctx.Done():
			stopSpinner() // stop spinner on context cancel
			pinger.Stop()
		case <-sigCh:
			fmt.Println("\n[!] Canceled by user (Ctrl-C)")
			stopSpinner()
			pinger.Stop()
		}
	}()

	pinger.OnRecv = func(pkt *probing.Packet) {
		opts.Stats.Total++
		opts.Stats.Successes++
		opts.Stats.Latencies = append(opts.Stats.Latencies, pkt.Rtt)
		opts.Stats.TotalLatency += pkt.Rtt

		if opts.Stats.MinLatency == 0 || pkt.Rtt < opts.Stats.MinLatency {
			opts.Stats.MinLatency = pkt.Rtt
		}
		if pkt.Rtt > opts.Stats.MaxLatency {
			opts.Stats.MaxLatency = pkt.Rtt
		}

		// Temporarily stop spinner to print message, then restart
		stopSpinner()

		msg := fmt.Sprintf("[OK] %d bytes from %s: icmp_seq=%d time=%v",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
		fmt.Println(msg)
		if opts.LogToFile && opts.Logger != nil {
			opts.Logger.Println(msg)
		}

		stopSpinner = spinner.StartSpinner("")
	}

	// Throttle failure message to avoid flooding on repeated timeouts
	var lastFailMsgTime time.Time

	pinger.OnRecvError = func(err error) {
		now := time.Now()
		// Only print failure if 1 second has passed since last failure message (adjust as needed)
		if now.Sub(lastFailMsgTime) > time.Second {
			stopSpinner()
			msg := fmt.Sprintf("[FAIL] No reply from %s (error: %v)", opts.Target, err)
			fmt.Println(msg)
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}
			stopSpinner = spinner.StartSpinner("")
			lastFailMsgTime = now
		}
	}

	pinger.OnFinish = func(stats *probing.Statistics) {
		opts.Stats.Total = stats.PacketsSent
		opts.Stats.Successes = stats.PacketsRecv
		opts.Stats.Failures = stats.PacketsSent - stats.PacketsRecv

		stopSpinner() // done, stop spinner
		PrettyPrintPingSummaryTable(opts)
	}

	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())

	err = pinger.Run()
	if err != nil && opts.Ctx.Err() == nil {
		stopSpinner()
		return fmt.Errorf("pinger error: %w", err)
	}

	return nil
}
