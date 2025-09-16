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

// loadSchemaCmd loads schema information for a table
func (m UIModel) loadSchemaCmd(tableName string) tea.Cmd {
	return func() tea.Msg {
		// Query PRAGMA table_info to get schema information
		query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		_, rows, err := m.svc.Query(query, nil)
		if err != nil {
			return schemaLoadedMsg{tableName: tableName, schema: nil, err: err}
		}
		return schemaLoadedMsg{tableName: tableName, schema: rows, err: nil}
	}
}

// loadTableInfoCmd loads detailed table information
func (m UIModel) loadTableInfoCmd(tableName string) tea.Cmd {
	return func() tea.Msg {
		// Get row count and other table statistics
		query := fmt.Sprintf(`
			SELECT 
				'%s' as table_name,
				COUNT(*) as row_count,
				(SELECT name FROM sqlite_master WHERE type='table' AND name='%s') as exists_check
			FROM %s`, tableName, tableName, tableName)

		_, rows, err := m.svc.QueryWithPagination(query, nil, 1000, 0)
		if err != nil {
			return tableInfoLoadedMsg{tableName: tableName, info: nil, err: err}
		}
		return tableInfoLoadedMsg{tableName: tableName, info: rows, err: nil}
	}
}

// loadIndexesCmd loads database indexes
func (m UIModel) loadIndexesCmd() tea.Cmd {
	return func() tea.Msg {
		query := `
			SELECT 
				name,
				tbl_name as table_name,
				sql,
				type
			FROM sqlite_master 
			WHERE type='index' 
			ORDER BY tbl_name, name`

		_, rows, err := m.svc.QueryWithPagination(query, nil, 1000, 0)
		if err != nil {
			return indexesLoadedMsg{indexes: nil, err: err}
		}
		return indexesLoadedMsg{indexes: rows, err: nil}
	}
}

// loadViewsCmd loads database views
func (m UIModel) loadViewsCmd() tea.Cmd {
	return func() tea.Msg {
		query := `
			SELECT 
				name,
				sql,
				type
			FROM sqlite_master 
			WHERE type='view' 
			ORDER BY name`

		_, rows, err := m.svc.QueryWithPagination(query, nil, 1000, 0)
		if err != nil {
			return viewsLoadedMsg{views: nil, err: err}
		}
		return viewsLoadedMsg{views: rows, err: nil}
	}
}
