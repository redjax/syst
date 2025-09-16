package ui

import (
	"fmt"
	"strings"
)

func (m UIModel) View() string {
	switch m.mode {
	case modeLauncher:
		return m.viewLauncher()
	case modeTable:
		return m.viewTable()
	case modeExpandCell:
		return m.viewExpandedCell()
	default:
		return ""
	}
}

func (m UIModel) viewLauncher() string {
	var b strings.Builder
	b.WriteString("SQLite Table Launcher\n\n")
	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("%s\n\n", m.errMsg))
	}
	if len(m.tables) == 0 {
		b.WriteString("(no tables)\n\n")
	}
	for i, tbl := range m.tables {
		cursor := "  "
		if i == m.tableIndex {
			cursor = "=>"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, tbl))
	}
	b.WriteString("\n↑/↓ (k/j): move | Enter: open | dd: delete table | q: quit\n")
	return b.String()
}

func (m UIModel) viewTable() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Table: %s (Esc back)\n", m.tableName))
	b.WriteString("SQL> " + m.queryInput.View() + "\n\n")

	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("%s\n\n", m.errMsg))
	}

	if m.loading {
		b.WriteString("Loading data...\n")
		return b.String()
	}

	// table component view
	b.WriteString(m.tableComp.View())

	// Show current cell information with prominent visual indicator
	if len(m.columns) > 0 && m.selectedCol < len(m.columns) {
		currentCol := m.columns[m.selectedCol]
		highlightedRow := m.tableComp.HighlightedRow()

		var cellValue string
		if val, exists := highlightedRow.Data[currentCol]; exists && val != nil {
			cellValue = fmt.Sprintf("%v", val)
			// Truncate long values for display
			if len(cellValue) > 40 {
				cellValue = cellValue[:37] + "..."
			}
		} else {
			cellValue = "(empty)"
		}

		// Get row ID for reference (try rowid first, then id)
		var rowRef string
		if val, exists := highlightedRow.Data["rowid"]; exists && val != nil {
			rowRef = fmt.Sprintf("rowid:%v", val)
		} else if val, exists := highlightedRow.Data["id"]; exists && val != nil {
			rowRef = fmt.Sprintf("id:%v", val)
		} else {
			rowRef = "row:?"
		}

		b.WriteString("\n" + strings.Repeat("═", 80) + "\n")
		b.WriteString(fmt.Sprintf("📍 CURSOR POSITION: Row %d/%d (%s) | Column: %s (%d/%d) | Value: %q\n",
			m.selectedIndex+1, len(m.rows), rowRef, currentCol, m.selectedCol+1, len(m.columns), cellValue))
		b.WriteString(strings.Repeat("═", 80) + "\n")
	}

	b.WriteString("↑/↓: row | ←/→: column | Space: select row | e: expand cell | n/p: page | dd: delete | /: query | q: quit\n")
	return b.String()
}

func (m UIModel) viewExpandedCell() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Expanded value - row %d, column %q\n\n", m.expandRow+1, m.expandCol))
	b.WriteString(m.vp.View())
	b.WriteString("\n[esc] to return\n")
	return b.String()
}
