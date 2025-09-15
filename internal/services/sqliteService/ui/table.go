package ui

import (
	"fmt"
	"strings"

	t "github.com/evertras/bubble-table/table"
)

func (m UIModel) buildTable() t.Model {
	// reserved keys we never want to show as DB columns
	reserved := map[string]struct{}{
		"__ui_selected__": {},
		"__selected":      {},
		"[x]":             {},
		"_selected":       {},
		"rowid":           {},
	}

	// Filter DB columns (trim & remove reserved)
	filteredCols := []string{}
	seen := make(map[string]bool)
	for _, c := range m.columns {
		n := strings.TrimSpace(c)
		if n == "" {
			continue
		}
		if _, ok := reserved[n]; ok {
			continue
		}
		if seen[n] {
			continue
		}
		seen[n] = true
		filteredCols = append(filteredCols, n)
	}

	// If there's nothing to show, give an empty model
	if len(filteredCols) == 0 {
		return t.New(nil)
	}

	// widths
	totalCols := len(filteredCols)
	minColWidth := 8
	usableWidth := m.termWidth - totalCols
	colWidth := usableWidth / totalCols
	if colWidth < minColWidth {
		colWidth = minColWidth
	}

	// column defs: just the database columns
	cols := []t.Column{}
	for _, c := range filteredCols {
		cols = append(cols, t.NewColumn(c, c, colWidth))
	}

	// rows
	var tRows []t.Row
	for _, rowData := range m.rows {
		row := t.RowData{}

		for _, colName := range filteredCols {
			val := ""
			if v, ok := rowData[colName]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
			}
			row[colName] = val
		}

		tRows = append(tRows, t.NewRow(row))
	}

	return t.New(cols).
		WithRows(tRows).
		SelectableRows(true).
		Focused(true)
}
