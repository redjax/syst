//go:build darwin
// +build darwin

package platformservice

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func detectOSRelease() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err == nil {
		return "macOS " + strings.TrimSpace(string(out))
	}
	return ""
}

func detectUptime() time.Duration {
	out, err := exec.Command("sysctl", "-n", "kern.boottime").Output()
	if err == nil {
		// Output: { sec = 1710000000, usec = 0 } ...
		s := string(out)
		secIdx := strings.Index(s, "sec =")
		if secIdx >= 0 {
			rest := s[secIdx+5:]
			rest = strings.TrimSpace(rest)
			parts := strings.Split(rest, ",")
			if len(parts) > 0 {
				secStr := strings.TrimSpace(parts[0])
				sec, err := strconv.ParseInt(secStr, 10, 64)
				if err == nil {
					boot := time.Unix(sec, 0)
					return time.Since(boot)
				}
			}
		}
	}
	return 0
}

func detectTotalRAM() uint64 {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err == nil {
		memStr := strings.TrimSpace(string(out))
		if bytes, err := strconv.ParseUint(memStr, 10, 64); err == nil {
			return bytes
		}
	}
	return 0
}
