package ui

import (
	"fmt"
	"strings"

	t "github.com/evertras/bubble-table/table"
)

func (m UIModel) buildTable() t.Model {
	const uiCheckboxCol = "__ui_selected__"

	// reserved keys we never want to show as DB columns
	reserved := map[string]struct{}{
		uiCheckboxCol: {},
		"__selected":  {},
		"[x]":         {},
		"_selected":   {},
		"rowid":       {},
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
	totalCols := len(filteredCols) + 1 // +1 for checkbox column
	minColWidth := 8
	usableWidth := m.termWidth - totalCols
	colWidth := usableWidth / totalCols
	if colWidth < minColWidth {
		colWidth = minColWidth
	}

	// column defs: UI checkbox first (label is a space)
	cols := []t.Column{t.NewColumn(uiCheckboxCol, " ", minColWidth)}
	for _, c := range filteredCols {
		cols = append(cols, t.NewColumn(c, c, colWidth))
	}

	// rows
	var tRows []t.Row
	for rowIdx, rowData := range m.rows {
		row := t.RowData{}

		// checkbox
		if m.selectedRows[rowIdx] {
			row[uiCheckboxCol] = "[x]"
		} else {
			row[uiCheckboxCol] = "[ ]"
		}

		for colIdx, colName := range filteredCols {
			val := ""
			if v, ok := rowData[colName]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
			}
			if (colIdx+1) == m.selectedCol && rowIdx == m.selectedIndex {
				val = "[" + val + "]"
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
