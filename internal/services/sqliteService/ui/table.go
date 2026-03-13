package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	// column defs: mark selected column header with indicator
	cols := []t.Column{}
	for i, c := range filteredCols {
		title := c
		if i == m.selectedCol {
			title = "► " + c
		}
		cols = append(cols, t.NewColumn(c, title, colWidth))
	}

	// Row highlight style for the focused row
	highlightStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("236")) // subtle dark gray background

	// rows
	var tRows []t.Row
	for _, rowData := range m.rows {
		row := t.RowData{}

		// Include ALL data from the original row (including rowid for deletion)
		for k, v := range rowData {
			if k == "rowid" {
				// Keep rowid as-is for deletion
				row[k] = v
			} else {
				// Format display columns as strings
				val := ""
				if v != nil {
					val = fmt.Sprintf("%v", v)
				}
				row[k] = val
			}
		}

		tRows = append(tRows, t.NewRow(row))
	}

	return t.New(cols).
		WithRows(tRows).
		SelectableRows(true).
		Focused(true).
		HighlightStyle(highlightStyle)
}
