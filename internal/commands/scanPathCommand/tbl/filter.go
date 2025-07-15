package tbl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/redjax/syst/internal/utils/convert"
)

type FilterExpr struct {
	Column   string
	Operator string
	Value    string
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
		valBytes := convert.ParseByteSize(val) // implement this to handle "10MB", "1GB", etc.
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
