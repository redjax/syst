package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// runQueryCmd loads current table/query results
func (m UIModel) runQueryCmd() tea.Cmd {
	return func() tea.Msg {
		cols, rows, err := m.svc.QueryWithPagination(m.query, nil, m.limit, m.offset)
		if err != nil {
			return fmt.Errorf("query error: %w", err)
		}
		return queryResultMsg{columns: cols, rows: rows}
	}
}

// loadTablesCmd fetches list of tables
func (m UIModel) loadTablesCmd() tea.Cmd {
	return func() tea.Msg {
		tables, err := m.svc.GetTables()
		if err != nil {
			return err
		}
		return tablesLoadedMsg(tables)
	}
}
