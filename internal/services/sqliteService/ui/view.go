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
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}
	for i, tbl := range m.tables {
		cursor := "  "
		if i == m.tableIndex {
			cursor = "=>"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, tbl))
	}
	b.WriteString("\n↑/↓: move | Enter: open | q: quit\n")
	return b.String()
}

func (m UIModel) viewTable() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Table: %s (Esc back)\n", m.tableName))
	b.WriteString("SQL> " + m.queryInput.View() + "\n\n")
	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}
	if m.loading {
		b.WriteString("Loading data...\n")
		return b.String()
	}
	b.WriteString(m.tableComp.View())
	b.WriteString("\nSpace: select row | e: expand cell | n/p: page | d: delete | q: quit\n")
	return b.String()
}

func (m UIModel) viewExpandedCell() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Expanded value - row %d, column %q\n\n", m.expandRow+1, m.expandCol))
	b.WriteString(m.expandVal + "\n")
	b.WriteString("\n[esc] to return\n")
	return b.String()
}
