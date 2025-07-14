package utils

import (
	"fmt"
)

func BytesToHumanReadable(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	// Units: KB, MB, GB, TB, PB, EB, ZB, YB
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

	if exp >= len(units) {
		return fmt.Sprintf("%.1f B", float64(bytes))
	}

	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
