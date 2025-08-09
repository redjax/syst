package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all TUI messages and user input.
func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		// -------- Expand Mode --------
		if m.mode == modeExpandCell {
			if msg.Type == tea.KeyEsc {
				m.mode = modeTable
			}
			return m, nil
		}

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
					// Query does NOT include any __selected col
					m.query = fmt.Sprintf("SELECT rowid, * FROM %s", m.tableName)
					m.offset = 0
					m.loading = true
					m.mode = modeTable
					return m, m.runQueryCmd()
				}
			}
			return m, nil
		}

		// -------- Table Mode --------
		if m.mode == modeTable {
			// Esc goes back to launcher
			if msg.Type == tea.KeyEsc {
				m.mode = modeLauncher
				m.loading = true
				return m, m.loadTablesCmd()
			}

			// Forward to bubble-table for its own internal navigation
			var tblCmd tea.Cmd
			m.tableComp, tblCmd = m.tableComp.Update(msg)

			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit

			case "up", "k":
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
			case "down", "j":
				if m.selectedIndex < len(m.rows)-1 {
					m.selectedIndex++
				}
			case "left", "h":
				if m.selectedCol > 0 {
					m.selectedCol--
				}
			case "right", "l":
				if m.selectedCol < len(m.columns)-1 {
					m.selectedCol++
				}

			case " ":
				// Toggle selection checkbox for the row under cursor (space only)
				if m.selectedRows[m.selectedIndex] {
					delete(m.selectedRows, m.selectedIndex)
				} else {
					m.selectedRows[m.selectedIndex] = true
				}
				m.tableComp = m.buildTable()

			case "e":
				// expand current cell value
				if m.selectedIndex >= 0 && m.selectedIndex < len(m.rows) &&
					m.selectedCol >= 0 && m.selectedCol < len(m.columns) {
					colKey := m.columns[m.selectedCol]
					if val, ok := m.rows[m.selectedIndex][colKey]; ok && val != nil {
						m.expandRow = m.selectedIndex
						m.expandCol = colKey
						m.expandVal = fmt.Sprintf("%v", val)
						m.mode = modeExpandCell
					}
				}

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
				if m.dCount == 2 {
					m.handleDelete(m.selectedIndex)
					m.dCount = 0
					m.loading = true
					return m, m.runQueryCmd()
				}

			default:
				m.dCount = 0
			}

			return m, tblCmd
		}

	// -------- Query Results Loaded --------
	case queryResultMsg:
		// Filter columns on receiving new query result
		filtered := []string{}
		for _, c := range msg.columns {
			if c != "__selected" && c != "[x]" && c != "_selected" {
				filtered = append(filtered, c)
			}
		}
		m.columns = filtered
		m.rows = msg.rows
		m.loading = false
		m.tableComp = m.buildTable()
		if m.selectedIndex >= len(m.rows) {
			m.selectedIndex = 0
		}
		if m.selectedCol >= len(m.columns) {
			m.selectedCol = 0
		}
		return m, nil

	// -------- Tables List Loaded --------
	case tablesLoadedMsg:
		m.tables = msg
		m.loading = false
		return m, nil

	// -------- Terminal Resize --------
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.tableComp = m.buildTable()
		return m, nil

	// -------- Async Error --------
	case error:
		m.errMsg = msg.Error()
		m.loading = false
		return m, nil
	}

	return m, nil
}

// handleDelete deletes a row by index from m.rows
func (m *UIModel) handleDelete(rowIdx int) {
	if rowIdx < 0 || rowIdx >= len(m.rows) {
		m.errMsg = "No row selected"
		return
	}
	if rid, ok := m.rows[rowIdx]["rowid"].(int64); ok {
		if err := m.svc.DeleteRow(m.tableName, rid); err != nil {
			m.errMsg = fmt.Sprintf("Delete error: %v", err)
		}
	} else {
		m.errMsg = "No rowid found"
	}
}
