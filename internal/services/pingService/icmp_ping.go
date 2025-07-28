package pingService

import (
	"fmt"
	"os"
	"os/signal"
	"sync"

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

		msg := fmt.Sprintf("[OK] %d bytes from %s: icmp_seq=%d time=%v",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
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

	// Run ping in goroutine so we can monitor per-ping timeouts
	go func() {
		_ = pinger.Run()
		close(results)
	}()

	var lastSeq int
	for res := range results {
		stopSpinner()
		fmt.Println(res.msg)
		if opts.LogToFile && opts.Logger != nil {
			opts.Logger.Println(res.msg)
		}
		stopSpinner = spinner.StartSpinner("")
		lastSeq = res.seq
	}

	// After pinger exits, check for missing pings
	for i := 0; i <= lastSeq; i++ {
		mu.Lock()
		if !received[i] {
			opts.Stats.Total++
			opts.Stats.Failures++
			msg := fmt.Sprintf("[FAIL] No reply from %s (icmp_seq=%d)", opts.Target, i)
			fmt.Println(msg)
			if opts.LogToFile && opts.Logger != nil {
				opts.Logger.Println(msg)
			}
		}
		mu.Unlock()
	}

	return nil
}
