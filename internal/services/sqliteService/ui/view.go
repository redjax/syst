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
	case modeSchema:
		return m.viewSchema()
	case modeTableInfo:
		return m.viewTableInfo()
	case modeIndexes:
		return m.viewIndexes()
	case modeViews:
		return m.viewViews()
	case modeImport:
		return m.viewImport()
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

	// Pagination indicator — always show both hints so user knows the keys
	page := m.offset/m.limit + 1
	hasNext := len(m.rows) == m.limit
	hasPrev := m.offset > 0
	var pageHints []string
	if hasPrev {
		pageHints = append(pageHints, "p: prev page")
	} else {
		pageHints = append(pageHints, "p: prev page (-)")
	}
	if hasNext {
		pageHints = append(pageHints, "n: next page")
	} else {
		pageHints = append(pageHints, "n: next page (-)")
	}
	b.WriteString(fmt.Sprintf("Page %d (%d rows) | %s\n", page, len(m.rows), strings.Join(pageHints, " | ")))

	b.WriteString("↑/↓: row | ←/→: column | Space: select | e: expand | x: export table | X: export selected | Ctrl+S: save results\n")
	b.WriteString("s: schema | i: info | I: indexes | v: views | dd: delete | /: query | m: import CSV | q: quit\n")
	return b.String()
}

func (m UIModel) viewExpandedCell() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Expanded value - row %d, column %q\n\n", m.expandRow+1, m.expandCol))
	b.WriteString(m.vp.View())
	b.WriteString("\n[esc] to return\n")
	return b.String()
}

func (m UIModel) viewSchema() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("📋 Schema: %s (Esc back)\n\n", m.tableName))

	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}

	if m.loading {
		b.WriteString("Loading schema...\n")
		return b.String()
	}

	if len(m.schemaInfo) == 0 {
		b.WriteString("No schema information available.\n")
	} else {
		b.WriteString("┌─────┬──────────────────┬──────────────┬─────────┬─────────────┬────────┐\n")
		b.WriteString("│ Pos │ Column Name      │ Data Type    │ NotNull │ Default     │ PK     │\n")
		b.WriteString("├─────┼──────────────────┼──────────────┼─────────┼─────────────┼────────┤\n")

		for _, row := range m.schemaInfo {
			cid := fmt.Sprintf("%v", row["cid"])
			name := fmt.Sprintf("%v", row["name"])
			dataType := fmt.Sprintf("%v", row["type"])
			notNull := fmt.Sprintf("%v", row["notnull"])
			defaultVal := fmt.Sprintf("%v", row["dflt_value"])
			pk := fmt.Sprintf("%v", row["pk"])

			if defaultVal == "<nil>" {
				defaultVal = ""
			}

			// Truncate long values
			if len(name) > 16 {
				name = name[:13] + "..."
			}
			if len(dataType) > 12 {
				dataType = dataType[:9] + "..."
			}
			if len(defaultVal) > 11 {
				defaultVal = defaultVal[:8] + "..."
			}

			b.WriteString(fmt.Sprintf("│ %-3s │ %-16s │ %-12s │ %-7s │ %-11s │ %-6s │\n",
				cid, name, dataType, notNull, defaultVal, pk))
		}
		b.WriteString("└─────┴──────────────────┴──────────────┴─────────┴─────────────┴────────┘\n")
	}

	b.WriteString("\n[Esc] back to table | s: schema | i: table info | I: indexes | v: views\n")
	return b.String()
}

func (m UIModel) viewTableInfo() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("ℹ️  Table Info: %s (Esc back)\n\n", m.tableName))

	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}

	if m.loading {
		b.WriteString("Loading table info...\n")
		return b.String()
	}

	if len(m.tableInfoData) == 0 {
		b.WriteString("No table information available.\n")
	} else {
		for _, row := range m.tableInfoData {
			b.WriteString(fmt.Sprintf("Table Name: %v\n", row["table_name"]))
			b.WriteString(fmt.Sprintf("Row Count:  %v\n", row["row_count"]))
			b.WriteString(fmt.Sprintf("Exists:     %v\n", row["exists_check"]))
		}
	}

	b.WriteString("\n[Esc] back to table | s: schema | i: table info | I: indexes | v: views\n")
	return b.String()
}

func (m UIModel) viewIndexes() string {
	var b strings.Builder
	b.WriteString("🗂️  Database Indexes (Esc back)\n\n")

	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}

	if m.loading {
		b.WriteString("Loading indexes...\n")
		return b.String()
	}

	if len(m.indexesData) == 0 {
		b.WriteString("No indexes found.\n")
	} else {
		b.WriteString("┌────────────────────────┬────────────────────────┬──────────────────────────────┐\n")
		b.WriteString("│ Index Name             │ Table Name             │ SQL Definition               │\n")
		b.WriteString("├────────────────────────┼────────────────────────┼──────────────────────────────┤\n")

		for _, row := range m.indexesData {
			name := fmt.Sprintf("%v", row["name"])
			tableName := fmt.Sprintf("%v", row["table_name"])
			sql := fmt.Sprintf("%v", row["sql"])

			// Truncate long values
			if len(name) > 22 {
				name = name[:19] + "..."
			}
			if len(tableName) > 22 {
				tableName = tableName[:19] + "..."
			}
			if len(sql) > 28 {
				sql = sql[:25] + "..."
			}

			b.WriteString(fmt.Sprintf("│ %-22s │ %-22s │ %-28s │\n", name, tableName, sql))
		}
		b.WriteString("└────────────────────────┴────────────────────────┴──────────────────────────────┘\n")
	}

	b.WriteString("\n[Esc] back to table | s: schema | i: table info | I: indexes | v: views\n")
	return b.String()
}

func (m UIModel) viewViews() string {
	var b strings.Builder
	b.WriteString("👁️  Database Views (Esc back)\n\n")

	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("Error: %s\n\n", m.errMsg))
	}

	if m.loading {
		b.WriteString("Loading views...\n")
		return b.String()
	}

	if len(m.viewsData) == 0 {
		b.WriteString("No views found.\n")
	} else {
		b.WriteString("┌────────────────────────┬─────────────────────────────────────────────────────────┐\n")
		b.WriteString("│ View Name              │ SQL Definition                                              │\n")
		b.WriteString("├────────────────────────┼─────────────────────────────────────────────────────────┤\n")

		for _, row := range m.viewsData {
			name := fmt.Sprintf("%v", row["name"])
			sql := fmt.Sprintf("%v", row["sql"])

			// Truncate long values
			if len(name) > 22 {
				name = name[:19] + "..."
			}
			if len(sql) > 55 {
				sql = sql[:52] + "..."
			}

			b.WriteString(fmt.Sprintf("│ %-22s │ %-55s │\n", name, sql))
		}
		b.WriteString("└────────────────────────┴─────────────────────────────────────────────────────────┘\n")
	}

	b.WriteString("\n[Esc] back to table | s: schema | i: table info | I: indexes | v: views\n")
	return b.String()
}

func (m UIModel) viewImport() string {
	var b strings.Builder
	b.WriteString("📁 CSV Import Wizard\n\n")

	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("%s\n\n", m.errMsg))
	}

	switch m.importStep {
	case 0:
		b.WriteString("Step 1: File Selection\n")
		b.WriteString("Enter the path to your CSV file:\n\n")
		b.WriteString(m.importFileInput.View())
		b.WriteString("\n\nPress [Tab] for path completion | [Enter] to proceed | [Esc] to cancel\n")
	case 1:
		b.WriteString("Step 2: Confirmation\n")
		b.WriteString(fmt.Sprintf("File: %s\n\n", m.importFilePath))
		b.WriteString("This will create a new table with the CSV data.\n")
		b.WriteString("Press [y] to confirm import or [n] to cancel\n")
	}

	b.WriteString("\n[Esc] Cancel and return to table view\n")
	return b.String()
}
