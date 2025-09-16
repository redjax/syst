package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// helper: convert interface{} -> int64 if possible
func toInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case int32:
		return int64(t), true
	case float64:
		return int64(t), true
	case []byte:
		s := string(t)
		if i, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			return i, true
		}
	case string:
		if i, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
			return i, true
		}
	}
	return 0, false
}

// helper: get map keys for debugging
func getMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// If in expand mode forward the message to the viewport (allows scrolling)
		if m.mode == modeExpandCell {
			// ESC to close expand
			if msg.Type == tea.KeyEsc {
				m.mode = modeTable
				return m, nil
			}
			// let the viewport handle scroll keys
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}

		// If in schema/info/indexes/views mode, handle ESC to return to table
		if m.mode == modeSchema || m.mode == modeTableInfo || m.mode == modeIndexes || m.mode == modeViews {
			if msg.Type == tea.KeyEsc {
				m.mode = modeTable
				return m, nil
			}
			// Allow navigation between different info modes
			switch msg.String() {
			case "s":
				m.loading = true
				m.mode = modeSchema
				return m, m.loadSchemaCmd(m.tableName)
			case "i":
				m.loading = true
				m.mode = modeTableInfo
				return m, m.loadTableInfoCmd(m.tableName)
			case "I":
				m.loading = true
				m.mode = modeIndexes
				return m, m.loadIndexesCmd()
			case "v":
				m.loading = true
				m.mode = modeViews
				return m, m.loadViewsCmd()
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		// If in import mode, handle import wizard steps
		if m.mode == modeImport {
			if msg.Type == tea.KeyEsc {
				m.mode = modeTable
				m.importFileInput.Blur()
				return m, nil
			}

			switch m.importStep {
			case 0: // File path input
				if m.importFileInput.Focused() {
					var cmd tea.Cmd
					m.importFileInput, cmd = m.importFileInput.Update(msg)
					if msg.String() == "enter" {
						filePath := m.importFileInput.Value()
						if filePath == "" {
							filePath = "sample_import.csv" // fallback for demo
						}
						m.importFilePath = filePath
						m.errMsg = "Import wizard: Press 'y' to confirm import or 'n' to cancel"
						m.importStep = 1
						m.importFileInput.Blur()
					}
					return m, cmd
				} else {
					// Auto-focus the input when entering import mode
					m.importFileInput.Focus()
					return m, nil
				}
			case 1: // Confirmation
				if msg.String() == "y" {
					m.loading = true
					return m, m.importCSVCmd()
				} else if msg.String() == "n" {
					m.mode = modeTable
					m.errMsg = "Import cancelled"
					m.importFileInput.SetValue("")
				}
			}
			return m, nil
		} // If query input is focused, let it handle key updates
		if m.queryInput.Focused() {
			var cmd tea.Cmd
			m.queryInput, _ = m.queryInput.Update(msg)
			if msg.String() == "enter" {
				// run query
				m.query = m.queryInput.Value()
				m.offset = 0
				m.loading = true
				m.queryInput.Blur()
				return m, m.runQueryCmd()
			}
			if msg.String() == "esc" {
				m.queryInput.Blur()
			}
			// let the viewport handle scroll keys
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}

		// Normal key handling per mode
		// -------- Launcher Mode --------
		if m.mode == modeLauncher {
			if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
				return m, tea.Quit
			}
			switch msg.String() {
			case "up", "k":
				if m.tableIndex > 0 {
					m.tableIndex--
				}
			case "down", "j":
				if m.tableIndex < len(m.tables)-1 {
					m.tableIndex++
				}
			case "enter":
				if len(m.tables) > 0 {
					m.tableName = m.tables[m.tableIndex]
					m.query = fmt.Sprintf("SELECT rowid, * FROM %s", m.tableName)
					m.offset = 0
					m.loading = true
					m.mode = modeTable
					return m, m.runQueryCmd()
				}
			case "d":
				m.dCount++
				if m.dCount == 2 && len(m.tables) > 0 {
					tableToDrop := m.tables[m.tableIndex]
					m.dCount = 0
					m.loading = true
					return m, m.dropTableCmd(tableToDrop)
				}
			default:
				m.dCount = 0
			}
			return m, nil
		}

		// -------- Table Mode --------
		if m.mode == modeTable {
			// Escape back to launcher
			if msg.Type == tea.KeyEsc {
				m.mode = modeLauncher
				m.loading = true
				return m, m.loadTablesCmd()
			}

			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case "up", "k":
				// row navigation - move up
				var cmd tea.Cmd
				m.tableComp, cmd = m.tableComp.Update(msg)
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
				return m, cmd

			case "down", "j":
				// row navigation - move down
				var cmd tea.Cmd
				m.tableComp, cmd = m.tableComp.Update(msg)
				if m.selectedIndex < len(m.rows)-1 {
					m.selectedIndex++
				}
				return m, cmd

			case "left", "h":
				// column navigation - move left
				if m.selectedCol > 0 {
					m.selectedCol--
					m.errMsg = fmt.Sprintf("INFO: Column %d/%d (%s)", m.selectedCol+1, len(m.columns), m.columns[m.selectedCol])
				}
				return m, nil

			case "right", "l":
				// column navigation - move right
				if m.selectedCol < len(m.columns)-1 {
					m.selectedCol++
					m.errMsg = fmt.Sprintf("INFO: Column %d/%d (%s)", m.selectedCol+1, len(m.columns), m.columns[m.selectedCol])
				}
				return m, nil

			case "/":
				// focus the query input for typing a new SQL
				m.queryInput.Focus()
				return m, nil

			case " ":
				// let the table component handle row selection
				m.errMsg = "INFO: Space pressed, delegating to table"
				var cmd tea.Cmd
				m.tableComp, cmd = m.tableComp.Update(msg)
				// Check selection after update
				selectedRows := m.tableComp.SelectedRows()
				m.errMsg += fmt.Sprintf(", now %d rows selected", len(selectedRows))
				return m, cmd

			case "e":
				// Get the currently highlighted row and expand the current column
				highlightedRow := m.tableComp.HighlightedRow()
				if len(m.columns) > 0 && m.selectedCol < len(m.columns) && len(highlightedRow.Data) > 0 {
					cellKey := m.columns[m.selectedCol]
					var cellContent string

					if val, exists := highlightedRow.Data[cellKey]; exists {
						cellContent = fmt.Sprintf("%v", val)
					} else {
						cellContent = "(no data)"
					}

					m.expandCol = cellKey
					m.expandVal = cellContent
					expandContent := fmt.Sprintf("Column: %s\n\nContent:\n%s", cellKey, cellContent)
					m.vp.SetContent(expandContent)
					m.mode = modeExpandCell
					m.errMsg = fmt.Sprintf("INFO: Expanded cell %s", cellKey)
				} else {
					m.errMsg = "INFO: No cell content to expand"
				}
				return m, nil

			case "s":
				// Show table schema
				m.loading = true
				m.mode = modeSchema
				return m, m.loadSchemaCmd(m.tableName)

			case "i":
				// Show table info
				m.loading = true
				m.mode = modeTableInfo
				return m, m.loadTableInfoCmd(m.tableName)

			case "I":
				// Show database indexes
				m.loading = true
				m.mode = modeIndexes
				return m, m.loadIndexesCmd()

			case "v":
				// Show database views
				m.loading = true
				m.mode = modeViews
				return m, m.loadViewsCmd()

			case "x":
				// Export current table/query results to CSV
				m.loading = true
				return m, m.exportTableToCSVCmd()

			case "X":
				// Export selected rows to CSV
				m.loading = true
				return m, m.exportSelectedToCSVCmd()

			case "ctrl+s":
				// Save current query results to file
				m.loading = true
				return m, m.saveQueryResultsCmd()

			case "m":
				// Start import wizard
				m.mode = modeImport
				m.importStep = 0
				m.importFilePath = ""
				m.importFileInput.SetValue("")
				m.importFileInput.Focus()
				m.errMsg = "Import wizard: Enter CSV file path"
				return m, nil

			case "n":
				if len(m.rows) == m.limit {
					m.offset += m.limit
					m.loading = true
					return m, m.runQueryCmd()
				}
			case "p":
				if m.offset >= m.limit {
					m.offset -= m.limit
					m.loading = true
					return m, m.runQueryCmd()
				}
			case "d":
				m.dCount++
				m.errMsg = fmt.Sprintf("d pressed, dCount now: %d", m.dCount)
				if m.dCount == 1 {
					// Single d pressed - just show debug info
					selectedRows := m.tableComp.SelectedRows()
					m.errMsg += fmt.Sprintf(", single d: %d selected rows", len(selectedRows))
				}
				if m.dCount == 2 {
					var rowIDs []int64

					// Get selected rows from the bubble-table component
					selectedRows := m.tableComp.SelectedRows()

					if len(selectedRows) > 0 {
						m.errMsg = fmt.Sprintf("DD: %d selected rows.", len(selectedRows))

						// Process ALL selected rows to collect their IDs
						for i, row := range selectedRows {
							var idValue interface{}
							var idKey string
							if v, ok := row.Data["rowid"]; ok {
								idValue = v
								idKey = "rowid"
							} else if v, ok := row.Data["id"]; ok {
								idValue = v
								idKey = "id"
							}

							if idValue != nil {
								if id, ok := toInt64(idValue); ok {
									rowIDs = append(rowIDs, id)
									m.errMsg += fmt.Sprintf(" Row[%d]: %s=%d", i, idKey, id)
								} else {
									m.errMsg += fmt.Sprintf(" Row[%d]: conversion failed", i)
								}
							} else {
								m.errMsg += fmt.Sprintf(" Row[%d]: no ID", i)
							}
						}
					} else {
						m.errMsg = "DD: NO SELECTED ROWS"
						// Try highlighted row as fallback
						highlightedRow := m.tableComp.HighlightedRow()
						if v, ok := highlightedRow.Data["rowid"]; ok {
							if id, ok := toInt64(v); ok {
								rowIDs = append(rowIDs, id)
								m.errMsg += fmt.Sprintf(", using highlighted rowid %d", id)
							}
						} else if v, ok := highlightedRow.Data["id"]; ok {
							if id, ok := toInt64(v); ok {
								rowIDs = append(rowIDs, id)
								m.errMsg += fmt.Sprintf(", using highlighted id %d", id)
							}
						}
					}

					m.dCount = 0
					if len(rowIDs) > 0 {
						m.errMsg = fmt.Sprintf("Deleting %d rows with IDs: %v", len(rowIDs), rowIDs)
						m.loading = true
						return m, m.deleteRowsCmd(rowIDs)
					} else {
						m.errMsg += " - No valid rowIDs found"
					}
				}
			case "D": // Capital D for immediate delete (testing)
				var rowIDs []int64

				// Get selected rows from the bubble-table component
				selectedRows := m.tableComp.SelectedRows()
				m.errMsg = fmt.Sprintf("Capital D INFO: Found %d selected rows", len(selectedRows))

				if len(selectedRows) > 0 {
					// Process selected rows
					for i, row := range selectedRows {
						m.errMsg += fmt.Sprintf(", row[%d] has keys: %v", i, getMapKeys(row.Data))
						if v, ok := row.Data["rowid"]; ok {
							m.errMsg += fmt.Sprintf(", rowid=%v", v)
							if id, ok := toInt64(v); ok {
								rowIDs = append(rowIDs, id)
								m.errMsg += fmt.Sprintf(", converted to int64=%d", id)
							} else {
								m.errMsg += ", conversion to int64 failed"
							}
						} else if v, ok := row.Data["id"]; ok {
							m.errMsg += fmt.Sprintf(", id=%v", v)
							if id, ok := toInt64(v); ok {
								rowIDs = append(rowIDs, id)
								m.errMsg += fmt.Sprintf(", converted to int64=%d", id)
							} else {
								m.errMsg += ", conversion to int64 failed"
							}
						} else {
							m.errMsg += ", no rowid or id key found"
						}
					}
				} else {
					// If no rows selected, try to get the current highlighted row
					highlightedRow := m.tableComp.HighlightedRow()
					m.errMsg += fmt.Sprintf(", highlighted row keys: %v", getMapKeys(highlightedRow.Data))
					if v, ok := highlightedRow.Data["rowid"]; ok {
						if id, ok := toInt64(v); ok {
							rowIDs = append(rowIDs, id)
							m.errMsg += fmt.Sprintf(", using highlighted rowid %d", id)
						}
					} else if v, ok := highlightedRow.Data["id"]; ok {
						if id, ok := toInt64(v); ok {
							rowIDs = append(rowIDs, id)
							m.errMsg += fmt.Sprintf(", using highlighted id %d", id)
						}
					}
				}

				if len(rowIDs) > 0 {
					m.errMsg = fmt.Sprintf("Capital D INFO: Deleting %d rows with IDs: %v", len(rowIDs), rowIDs)
					m.loading = true
					return m, m.deleteRowsCmd(rowIDs)
				} else {
					m.errMsg = "Capital D INFO: No rows to delete - no valid rowIDs found"
				}
			case "S": // Capital S to show current selections (testing)
				selectedRows := m.tableComp.SelectedRows()
				m.errMsg = fmt.Sprintf("Selection check: %d rows selected", len(selectedRows))
				if len(selectedRows) > 0 {
					for i, row := range selectedRows {
						keys := getMapKeys(row.Data)
						m.errMsg += fmt.Sprintf(", row[%d] keys: %v", i, keys)
					}
				}
			}

			return m, nil
		}

	// -------- Query Results Loaded --------
	case queryResultMsg:
		// INFO: Show raw query results
		if len(msg.rows) > 0 {
			rawKeys := make([]string, 0, len(msg.rows[0]))
			for k := range msg.rows[0] {
				rawKeys = append(rawKeys, k)
			}
			m.errMsg = fmt.Sprintf("RAW QUERY KEYS: %v", rawKeys)
		}

		// sanitize and trim column keys and drop reserved names
		reserved := map[string]struct{}{
			"__ui_selected__": {},
			"__selected":      {},
			"[x]":             {},
			"_selected":       {},
			"rowid":           {}, // exclude from visible columns
		}
		filtered := []string{}
		for _, c := range msg.columns {
			n := strings.TrimSpace(c)
			if _, isReserved := reserved[n]; isReserved {
				continue
			}
			filtered = append(filtered, n)
		}
		m.columns = filtered

		// sanitize rows: trim keys and drop reserved keys (but keep rowid for deletion)
		cleanedRows := make([]map[string]interface{}, len(msg.rows))
		for i, row := range msg.rows {
			clean := make(map[string]interface{}, 0)
			for k, v := range row {
				n := strings.TrimSpace(k)
				// Keep rowid for deletion, but exclude other reserved keys
				if n == "rowid" {
					clean[n] = v
					continue
				}
				if _, isReserved := reserved[n]; isReserved {
					continue
				}
				clean[n] = v
			}
			cleanedRows[i] = clean
		}
		m.rows = cleanedRows

		m.loading = false
		m.tableComp = m.buildTable()

		// Clamp selected index and column within valid bounds
		if m.selectedIndex >= len(m.rows) {
			m.selectedIndex = 0
		}
		if m.selectedCol < 0 || m.selectedCol >= len(m.columns) {
			m.selectedCol = 0
		}

		return m, nil

	case deleteDoneMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
		} else {
			m.errMsg = ""
		}
		m.loading = false
		// reload current page
		return m, m.runQueryCmd()

	case dropDoneMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.loading = false
			return m, nil
		}
		// reload tables list
		m.loading = false
		return m, m.loadTablesCmd()

	case tablesLoadedMsg:
		m.tables = msg
		m.loading = false
		return m, nil

	case schemaLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.loading = false
			m.mode = modeTable
			return m, nil
		}
		m.schemaInfo = msg.schema
		m.loading = false
		return m, nil

	case tableInfoLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.loading = false
			m.mode = modeTable
			return m, nil
		}
		m.tableInfoData = msg.info
		m.loading = false
		return m, nil

	case indexesLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.loading = false
			m.mode = modeTable
			return m, nil
		}
		m.indexesData = msg.indexes
		m.loading = false
		return m, nil

	case viewsLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			m.loading = false
			m.mode = modeTable
			return m, nil
		}
		m.viewsData = msg.views
		m.loading = false
		return m, nil

	case exportDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Export failed: %s", msg.err.Error())
		} else {
			m.errMsg = fmt.Sprintf("✅ Exported %d rows to: %s", msg.rowCount, msg.filename)
		}
		return m, nil

	case importDoneMsg:
		m.loading = false
		m.mode = modeTable
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Import failed: %s", msg.err.Error())
		} else {
			m.errMsg = fmt.Sprintf("✅ Imported %d rows into table: %s", msg.rowCount, msg.tableName)
			// Refresh the tables list to show the new table
			return m, m.loadTablesCmd()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

		// resize viewport (expand view) if present
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height - 6

		m.tableComp = m.buildTable()
		return m, nil

	case error:
		// DB/command errors are sent back as errors
		m.errMsg = msg.Error()
		m.loading = false
		return m, nil
	}

	return m, nil
}
