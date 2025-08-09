package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// runQueryCmd runs the current SQL query (with limit/offset) in a goroutine.
func (m UIModel) runQueryCmd() tea.Cmd {
	return func() tea.Msg {
		cols, rows, err := m.svc.QueryWithPagination(m.query, nil, m.limit, m.offset)
		if err != nil {
			return fmt.Errorf("query error: %w", err)
		}
		return queryResultMsg{columns: cols, rows: rows}
	}
}

// loadTablesCmd loads table list
func (m UIModel) loadTablesCmd() tea.Cmd {
	return func() tea.Msg {
		tables, err := m.svc.GetTables()
		if err != nil {
			return fmt.Errorf("tables load error: %w", err)
		}
		return tablesLoadedMsg(tables)
	}
}

// deleteRowsCmd deletes the supplied rowids from the currently-open table
func (m UIModel) deleteRowsCmd(rowIDs []int64) tea.Cmd {
	return func() tea.Msg {
		var lastErr error
		for _, id := range rowIDs {
			if err := m.svc.DeleteRow(m.tableName, id); err != nil {
				lastErr = err
			}
		}
		return deleteDoneMsg{err: lastErr}
	}
}

// dropTableCmd drops the given table name
func (m UIModel) dropTableCmd(tableName string) tea.Cmd {
	return func() tea.Msg {
		err := m.svc.DropTable(tableName)
		return dropDoneMsg{err: err}
	}
}
