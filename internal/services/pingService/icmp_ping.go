//go:build !windows
// +build !windows

package pingService

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
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

	stopSpinner := spinner.StartSpinner("")
	defer stopSpinner()

	sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	go func() {
		select {
		case <-opts.Ctx.Done():
			stopSpinner()

			pinger.Stop()
		case <-sigCh:
			fmt.Println("\n[!] Canceled by user (Ctrl-C)")
			stopSpinner()

			pinger.Stop()
		}
	}()

	type pingResult struct {
		seq int
		ok  bool
		msg string
	}

	results := make(chan pingResult, 100)
	var mu sync.Mutex

	received := map[int]bool{}
	sentTime := map[int]time.Time{} // track when each seq was sent

	pinger.OnRecv = func(pkt *probing.Packet) {
		mu.Lock()
		defer mu.Unlock()
		received[pkt.Seq] = true

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

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		msg := fmt.Sprintf("[%s] [OK] %d bytes from %s: icmp_seq=%d time=%v",
			timestamp, pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)

		results <- pingResult{seq: pkt.Seq, ok: true, msg: msg}
	}

	pinger.OnFinish = func(stats *probing.Statistics) {
		stopSpinner()

		opts.Stats.Total = stats.PacketsSent
		opts.Stats.Successes = stats.PacketsRecv
		opts.Stats.Failures = stats.PacketsSent - stats.PacketsRecv

		PrettyPrintPingSummaryTable(opts)
	}

	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())

	// Run pinger in goroutine
	go func() {
		_ = pinger.Run()
		close(results)
	}()

	// Separate goroutine to track timeouts per ping sequence
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		timeout := pinger.Interval * 2 // adjust timeout window

		for {
			select {
			case <-opts.Ctx.Done():
				return
			case <-ticker.C:
				mu.Lock()
				for seq, sentAt := range sentTime {
					if received[seq] {
						// already received
						continue
					}

					if time.Since(sentAt) > timeout {
						stopSpinner()

						timestamp := time.Now().Format("2006-01-02 15:04:05")
						// timeout exceeded: print failure and mark it handled
						fmt.Printf("[%s] [FAIL] No reply from %s (icmp_seq=%d)\n", timestamp, opts.Target, seq)
						if opts.LogToFile && opts.Logger != nil {
							opts.Logger.Printf("[%s] [FAIL] No reply from %s (icmp_seq=%d)\n", timestamp, opts.Target, seq)
						}

						opts.Stats.Total++
						opts.Stats.Failures++

						delete(sentTime, seq) // remove so not printed again

						stopSpinner = spinner.StartSpinner("")
					}
				}
				mu.Unlock()
			}
		}
	}()

	// Capture sent pings by periodically reading stats
	go func() {
		ticker := time.NewTicker(pinger.Interval)
		defer ticker.Stop()

		var lastSent int = -1

		for {
			select {
			case <-opts.Ctx.Done():
				return
			case <-ticker.C:
				stats := pinger.Statistics()
				mu.Lock()

				for seq := lastSent + 1; seq < stats.PacketsSent; seq++ {
					sentTime[seq] = time.Now()
					opts.Stats.Total++
				}

				lastSent = stats.PacketsSent - 1

				mu.Unlock()
			}
		}
	}()

	// Main loop to print results from OnRecv
	for res := range results {
		stopSpinner()
		fmt.Println(res.msg)

		if opts.LogToFile && opts.Logger != nil {
			opts.Logger.Println(res.msg)
		}

		stopSpinner = spinner.StartSpinner("")
	}

	// After pinger exits, remaining missing failures will have been printed by timeout goroutine.

	return nil
}
