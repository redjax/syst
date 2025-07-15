package tbl

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Helper: parse time in your format
func parseTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

// SortResults sorts a scan results table by given properties (size, created, etc)
func SortResults(results [][]string, column, order string) {
	// If user asks for sizeparsed, sort by size instead
	if column == "sizeparsed" {
		column = "size"
	}

	idx, ok := columnIndices[strings.ToLower(column)]
	if !ok {
		idx = 0 // Default to "name"
	}

	desc := strings.ToLower(order) == "desc"

	sort.Slice(results, func(i, j int) bool {
		a, b := results[i][idx], results[j][idx]

		// Numeric sort for size
		if column == "size" {
			ai, _ := strconv.ParseInt(a, 10, 64)
			bi, _ := strconv.ParseInt(b, 10, 64)
			if desc {
				return ai > bi
			}

			return ai < bi
		}

		// Time sort for created/modified
		if column == "created" || column == "modified" {
			at, _ := parseTime(a)
			bt, _ := parseTime(b)
			if desc {
				return at.After(bt)
			}

			return at.Before(bt)
		}

		// Alphanumeric sort for others (case-insensitive)
		la, lb := strings.ToLower(a), strings.ToLower(b)
		if desc {
			return la > lb
		}

		return la < lb
	})
}

// ByteCount returns number of bytes in string form
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0

	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
