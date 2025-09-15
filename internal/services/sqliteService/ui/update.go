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

		// If query input is focused, let it handle key updates
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

			case "up", "k", "down", "j", "left", "h", "right", "l":
				// let the table component handle navigation
				var cmd tea.Cmd
				m.tableComp, cmd = m.tableComp.Update(msg)
				return m, cmd

			case "/":
				// focus the query input for typing a new SQL
				m.queryInput.Focus()
				return m, nil

			case " ":
				// let the table component handle row selection
				var cmd tea.Cmd
				m.tableComp, cmd = m.tableComp.Update(msg)
				return m, cmd

			case "e":
				// TODO: implement cell expansion using bubble-table's cursor position
				// For now, disabled until we can get the current selection from the table component
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
				if m.dCount == 2 {
					var rowIDs []int64
					if len(m.selectedRows) > 0 {
						for idx := range m.selectedRows {
							if idx >= 0 && idx < len(m.rows) {
								if v, ok := m.rows[idx]["rowid"]; ok {
									if id, ok := toInt64(v); ok {
										rowIDs = append(rowIDs, id)
									}
								}
							}
						}
						m.selectedRows = make(map[int]bool)
					} else {
						if m.selectedIndex >= 0 && m.selectedIndex < len(m.rows) {
							if v, ok := m.rows[m.selectedIndex]["rowid"]; ok {
								if id, ok := toInt64(v); ok {
									rowIDs = append(rowIDs, id)
								}
							}
						}
					}
					m.dCount = 0
					if len(rowIDs) > 0 {
						m.loading = true
						return m, m.deleteRowsCmd(rowIDs)
					}
				}
			default:
				m.dCount = 0
			}

			return m, nil
		}

	// -------- Query Results Loaded --------
	case queryResultMsg:
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

		// sanitize rows: trim keys and drop reserved keys
		cleanedRows := make([]map[string]interface{}, len(msg.rows))
		for i, row := range msg.rows {
			clean := make(map[string]interface{}, 0)
			for k, v := range row {
				n := strings.TrimSpace(k)
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
		if m.selectedCol < 1 || m.selectedCol >= len(m.columns) {
			m.selectedCol = 1
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
