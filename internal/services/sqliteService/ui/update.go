package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// Quit from launcher
		if m.mode == modeLauncher && (msg.Type == tea.KeyCtrlC || msg.String() == "q") {
			return m, tea.Quit
		}

		// Query input mode
		if m.inQueryInput {
			var cmd tea.Cmd
			m.queryInput, cmd = m.queryInput.Update(msg)
			if msg.Type == tea.KeyEnter {
				m.query = m.queryInput.Value()
				m.offset = 0
				m.selectedIndex = 0
				m.inQueryInput = false
				m.loading = true
				return m, m.runQueryCmd()
			}
			if msg.Type == tea.KeyEsc {
				m.inQueryInput = false
			}
			return m, cmd
		}

		// Launcher mode navigation
		if m.mode == modeLauncher {
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
					m.mode = modeTable
					m.loading = true
					return m, m.runQueryCmd()
				}
			}
			return m, nil
		}

		// Table mode navigation
		if m.mode == modeTable {
			// Return to launcher with Esc
			if msg.Type == tea.KeyEsc {
				m.mode = modeLauncher
				m.loading = true
				return m, m.loadTablesCmd()
			}
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "j", "down":
				if m.selectedIndex < len(m.rows)-1 {
					m.selectedIndex++
				}
				m.dCount = 0
			case "k", "up":
				if m.selectedIndex > 0 {
					m.selectedIndex--
				}
				m.dCount = 0
			case "n":
				if len(m.rows) == m.limit {
					m.offset += m.limit
					m.selectedIndex = 0
					m.loading = true
					return m, m.runQueryCmd()
				}
			case "p":
				if m.offset >= m.limit {
					m.offset -= m.limit
					m.selectedIndex = 0
					m.loading = true
					return m, m.runQueryCmd()
				}
			case "enter":
				m.inQueryInput = true
				m.queryInput.SetValue(m.query)
				m.queryInput.Focus()
			case "d":
				m.dCount++
				if m.dCount == 2 {
					m.handleDelete()
					m.dCount = 0
					return m, m.runQueryCmd()
				}
			default:
				m.dCount = 0
			}
		}

	case queryResultMsg:
		m.columns = msg.columns
		m.rows = msg.rows
		m.loading = false
		if m.selectedIndex >= len(m.rows) {
			m.selectedIndex = max(0, len(m.rows)-1)
		}
		return m, nil

	case tablesLoadedMsg:
		m.tables = msg
		m.loading = false
		return m, nil

	case error:
		m.errMsg = msg.Error()
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m *UIModel) handleDelete() {
	if m.selectedIndex < len(m.rows) {
		row := m.rows[m.selectedIndex]
		if rid, ok := row["rowid"].(int64); ok {
			if err := m.svc.DeleteRow(m.tableName, rid); err != nil {
				m.errMsg = fmt.Sprintf("Delete error: %v", err)
			}
		} else {
			m.errMsg = "No rowid found"
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
