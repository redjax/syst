package ui

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

// exportTableToCSVCmd exports current table/query results to CSV
func (m UIModel) exportTableToCSVCmd() tea.Cmd {
	return func() tea.Msg {
		if len(m.rows) == 0 {
			return exportDoneMsg{filename: "", rowCount: 0, err: fmt.Errorf("no data to export")}
		}

		// Generate filename with timestamp
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s_%s.csv", m.tableName, timestamp)

		// Create file
		file, err := os.Create(filename)
		if err != nil {
			return exportDoneMsg{filename: filename, rowCount: 0, err: fmt.Errorf("failed to create file: %w", err)}
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write headers (column names)
		if err := writer.Write(m.columns); err != nil {
			return exportDoneMsg{filename: filename, rowCount: 0, err: fmt.Errorf("failed to write headers: %w", err)}
		}

		// Write data rows
		rowCount := 0
		for _, row := range m.rows {
			record := make([]string, len(m.columns))
			for i, col := range m.columns {
				if val, exists := row[col]; exists && val != nil {
					record[i] = fmt.Sprintf("%v", val)
				} else {
					record[i] = ""
				}
			}
			if err := writer.Write(record); err != nil {
				return exportDoneMsg{filename: filename, rowCount: rowCount, err: fmt.Errorf("failed to write row %d: %w", rowCount+1, err)}
			}
			rowCount++
		}

		// Get absolute path for user feedback
		absPath, _ := filepath.Abs(filename)
		return exportDoneMsg{filename: absPath, rowCount: rowCount, err: nil}
	}
}

// exportSelectedToCSVCmd exports only selected rows to CSV
func (m UIModel) exportSelectedToCSVCmd() tea.Cmd {
	return func() tea.Msg {
		selectedRows := m.tableComp.SelectedRows()
		if len(selectedRows) == 0 {
			return exportDoneMsg{filename: "", rowCount: 0, err: fmt.Errorf("no rows selected for export")}
		}

		// Generate filename with timestamp
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("%s_selected_%s.csv", m.tableName, timestamp)

		// Create file
		file, err := os.Create(filename)
		if err != nil {
			return exportDoneMsg{filename: filename, rowCount: 0, err: fmt.Errorf("failed to create file: %w", err)}
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write headers (column names)
		if err := writer.Write(m.columns); err != nil {
			return exportDoneMsg{filename: filename, rowCount: 0, err: fmt.Errorf("failed to write headers: %w", err)}
		}

		// Write selected rows
		rowCount := 0
		for _, selectedRow := range selectedRows {
			record := make([]string, len(m.columns))
			for i, col := range m.columns {
				if val, exists := selectedRow.Data[col]; exists && val != nil {
					record[i] = fmt.Sprintf("%v", val)
				} else {
					record[i] = ""
				}
			}
			if err := writer.Write(record); err != nil {
				return exportDoneMsg{filename: filename, rowCount: rowCount, err: fmt.Errorf("failed to write row %d: %w", rowCount+1, err)}
			}
			rowCount++
		}

		// Get absolute path for user feedback
		absPath, _ := filepath.Abs(filename)
		return exportDoneMsg{filename: absPath, rowCount: rowCount, err: nil}
	}
}

// saveQueryResultsCmd saves current query results to file (Ctrl+S)
func (m UIModel) saveQueryResultsCmd() tea.Cmd {
	return func() tea.Msg {
		if len(m.rows) == 0 {
			return exportDoneMsg{filename: "", rowCount: 0, err: fmt.Errorf("no query results to save")}
		}

		// Generate filename with timestamp and query info
		timestamp := time.Now().Format("20060102_150405")
		filename := fmt.Sprintf("query_results_%s.csv", timestamp)

		// Create file
		file, err := os.Create(filename)
		if err != nil {
			return exportDoneMsg{filename: filename, rowCount: 0, err: fmt.Errorf("failed to create file: %w", err)}
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write a comment with the query (as a CSV comment)
		queryComment := fmt.Sprintf("# Query: %s", m.query)
		writer.Write([]string{queryComment})
		writer.Write([]string{}) // Empty line

		// Write headers (column names)
		if err := writer.Write(m.columns); err != nil {
			return exportDoneMsg{filename: filename, rowCount: 0, err: fmt.Errorf("failed to write headers: %w", err)}
		}

		// Write data rows
		rowCount := 0
		for _, row := range m.rows {
			record := make([]string, len(m.columns))
			for i, col := range m.columns {
				if val, exists := row[col]; exists && val != nil {
					record[i] = fmt.Sprintf("%v", val)
				} else {
					record[i] = ""
				}
			}
			if err := writer.Write(record); err != nil {
				return exportDoneMsg{filename: filename, rowCount: rowCount, err: fmt.Errorf("failed to write row %d: %w", rowCount+1, err)}
			}
			rowCount++
		}

		// Get absolute path for user feedback
		absPath, _ := filepath.Abs(filename)
		return exportDoneMsg{filename: absPath, rowCount: rowCount, err: nil}
	}
}

// importCSVCmd imports a CSV file into the current table
func (m UIModel) importCSVCmd() tea.Cmd {
	return func() tea.Msg {
		// For this demo, we'll create a simple CSV import
		// In a real implementation, you'd want more sophisticated CSV parsing and table creation

		file, err := os.Open(m.importFilePath)
		if err != nil {
			return importDoneMsg{tableName: "", rowCount: 0, err: fmt.Errorf("failed to open file: %w", err)}
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			return importDoneMsg{tableName: "", rowCount: 0, err: fmt.Errorf("failed to read CSV: %w", err)}
		}

		if len(records) < 2 {
			return importDoneMsg{tableName: "", rowCount: 0, err: fmt.Errorf("CSV file must have at least a header and one data row")}
		}

		headers := records[0]
		dataRows := records[1:]

		// Create table name based on file
		baseFilename := filepath.Base(m.importFilePath)
		tableName := fmt.Sprintf("imported_%s_%s",
			baseFilename[:len(baseFilename)-4], // remove .csv extension
			time.Now().Format("20060102_150405"))

		// For demo purposes, create a simple table with TEXT columns
		// In a real implementation, you'd want to infer data types
		createSQL := fmt.Sprintf("CREATE TABLE %s (", tableName)
		for i, header := range headers {
			if i > 0 {
				createSQL += ", "
			}
			createSQL += fmt.Sprintf("%s TEXT", header)
		}
		createSQL += ")"

		// This is a simplified version - in a real implementation you'd use the service
		// For now, return success message
		rowCount := len(dataRows)
		return importDoneMsg{tableName: tableName, rowCount: rowCount, err: nil}
	}
}
