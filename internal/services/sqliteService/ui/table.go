package ui

import (
	"fmt"

	t "github.com/evertras/bubble-table/table"
)

func (m *UIModel) buildTable() t.Model {
	// 1. Filter m.columns so no “[x]” or “__selected”
	filteredCols := []string{}
	for _, c := range m.columns {
		if c != "__selected" && c != "[x]" && c != "_selected" {
			filteredCols = append(filteredCols, c)
		}
	}

	// 2. Clean each row so no “[x]” or “__selected” keys exist
	cleanedRows := make([]map[string]interface{}, len(m.rows))
	for i, row := range m.rows {
		clean := make(map[string]interface{}, len(row))
		for k, v := range row {
			if k != "__selected" && k != "[x]" && k != "_selected" {
				clean[k] = v
			}
		}
		cleanedRows[i] = clean
	}

	// 3. Calc width
	totalCols := len(filteredCols) + 1
	if totalCols == 0 {
		return t.New(nil)
	}
	minColWidth := 8
	usableWidth := m.termWidth - totalCols
	colWidth := usableWidth / totalCols
	if colWidth < minColWidth {
		colWidth = minColWidth
	}

	// 4. Define columns: one checkbox + the real DB columns
	cols := []t.Column{t.NewColumn("__selected", "[x]", minColWidth)}
	for _, c := range filteredCols {
		cols = append(cols, t.NewColumn(c, c, colWidth))
	}

	// 5. Build table rows
	var tRows []t.Row
	for rowIdx, rowData := range cleanedRows {
		row := t.RowData{}

		if m.selectedRows[rowIdx] {
			row["__selected"] = "[x]"
		} else {
			row["__selected"] = "[ ]"
		}

		for colIdx, colName := range filteredCols {
			val := ""
			if v, ok := rowData[colName]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
			}
			// highlight current cell
			if colIdx == m.selectedCol && rowIdx == m.selectedIndex {
				val = "[" + val + "]"
			}
			row[colName] = val
		}

		tRows = append(tRows, t.NewRow(row))
	}

	return t.New(cols).WithRows(tRows).SelectableRows(true).Focused(true)
}
