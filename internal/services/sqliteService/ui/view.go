package ui

import (
	"fmt"
	"strings"
)

func (m UIModel) View() string {
	if m.mode == modeLauncher {
		return m.viewLauncher()
	}
	return m.viewTable()
}

func (m UIModel) viewLauncher() string {
	var b strings.Builder
	b.WriteString("SQLite Launcher - Select a table and press Enter. Esc/Q to quit.\n\n")
	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}
	if m.loading {
		b.WriteString("Loading tables...\n")
		return b.String()
	}
	for i, tbl := range m.tables {
		cursor := "  "
		if i == m.tableIndex {
			cursor = "=>"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, tbl))
	}
	return b.String()
}

func (m UIModel) viewTable() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Table: %s  (Esc to go back)\n\n", m.tableName))
	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}
	if m.loading {
		b.WriteString("Loading data...\n")
		return b.String()
	}
	if m.inQueryInput {
		b.WriteString("Query mode: Enter to run, Esc to cancel\n")
		b.WriteString(m.queryInput.View())
		return b.String()
	}

	// Header
	for _, col := range m.columns {
		b.WriteString(fmt.Sprintf("%-15s", col))
	}
	b.WriteString("\n")

	// Rows
	for i, row := range m.rows {
		cursor := "  "
		if i == m.selectedIndex {
			cursor = "=>"
		}
		b.WriteString(cursor + " ")
		for _, col := range m.columns {
			val := ""
			if v, ok := row[col]; ok && v != nil {
				val = fmt.Sprintf("%v", v)
			}
			if len(val) > 14 {
				val = val[:14] + "…"
			}
			b.WriteString(fmt.Sprintf("%-15s", val))
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("\nOffset: %d  Rows: %d  Limit: %d\n", m.offset, len(m.rows), m.limit))
	return b.String()
}
