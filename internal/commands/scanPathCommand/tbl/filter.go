package tbl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type FilterExpr struct {
	Column   string
	Operator string
	Value    string
}

// ParseByteSize parses strings like "10MB", "1GB", "512K", or "15615" (bytes) into an int64 byte count.
// Supports both SI (kB, MB, GB) and IEC (KiB, MiB, GiB) units, case-insensitive.
func ParseByteSize(s string) int64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	multipliers := map[string]int64{
		"B":   1,
		"K":   1 << 10, // 1024
		"KB":  1 << 10,
		"KIB": 1 << 10,
		"M":   1 << 20, // 1024*1024
		"MB":  1 << 20,
		"MIB": 1 << 20,
		"G":   1 << 30,
		"GB":  1 << 30,
		"GIB": 1 << 30,
		"T":   1 << 40,
		"TB":  1 << 40,
		"TIB": 1 << 40,
	}

	// Find where the number ends and the unit begins
	i := 0
	for ; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			if s[i] != '.' { // allow decimal point
				break
			}
		}
	}
	numPart := s[:i]
	unitPart := strings.TrimSpace(s[i:])

	// Default to bytes if no unit
	if unitPart == "" {
		unitPart = "B"
	}

	val, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0 // or handle error as needed
	}
	mult, ok := multipliers[unitPart]
	if !ok {
		return 0 // or handle error as needed
	}
	return int64(val * float64(mult))
}

func ParseFilter(filter string) (*FilterExpr, error) {
	// Regex matches: column <operator> value, e.g. "size <10MB"
	re := regexp.MustCompile(`^\s*([a-zA-Z]+)\s*([<>=~*])\s*(.+)$`)

	matches := re.FindStringSubmatch(filter)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid filter format")
	}
	return &FilterExpr{
		Column:   strings.ToLower(matches[1]),
		Operator: matches[2],
		Value:    matches[3],
	}, nil
}

func MatchFilter(cell, op, val, column string) bool {
	switch column {
	case "size":
		cellBytes, _ := strconv.ParseInt(cell, 10, 64)
		valBytes := ParseByteSize(val) // implement this to handle "10MB", "1GB", etc.
		switch op {
		case "<":
			return cellBytes < valBytes
		case ">":
			return cellBytes > valBytes
		case "=":
			return cellBytes == valBytes
		}
	case "created", "modified":
		cellTime, _ := time.Parse("2006-01-02 15:04:05", cell)
		valTime, _ := time.Parse("2006-01-02", val)
		switch op {
		case "<":
			return cellTime.Before(valTime)
		case ">":
			return cellTime.After(valTime)
		case "=":
			return cellTime.Equal(valTime)
		}
	default:
		switch op {
		case "=":
			return strings.EqualFold(cell, val)
		case "~", "*":
			pattern := val
			if strings.Contains(pattern, "*") {
				pattern = regexp.QuoteMeta(pattern)
				pattern = strings.ReplaceAll(pattern, "\\*", ".*")
				pattern = "^" + pattern + "$"
				re := regexp.MustCompile("(?i)" + pattern)
				return re.MatchString(cell)
			}
			return strings.Contains(strings.ToLower(cell), strings.ToLower(val))
		}
	}
	return false
}

func ApplyFilter(results [][]string, filter *FilterExpr) [][]string {
	if filter == nil {
		return results
	}
	// map[string]int for your columns
	idx, ok := columnIndices[filter.Column]
	if !ok {
		return results // Unknown column, skip filter
	}
	filtered := make([][]string, 0, len(results))
	for _, row := range results {
		if MatchFilter(row[idx], filter.Operator, filter.Value, filter.Column) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}
