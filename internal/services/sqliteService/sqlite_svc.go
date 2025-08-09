package sqliteservice

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteService struct {
	db *sql.DB
}

// NewSQLiteService opens the database file
func NewSQLiteService(dbPath string) (*SQLiteService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &SQLiteService{db: db}, nil
}

// Close closes the DB
func (s *SQLiteService) Close() error {
	return s.db.Close()
}

// QueryWithPagination runs a query with LIMIT/OFFSET and returns columns, rows
func (s *SQLiteService) QueryWithPagination(query string, args []interface{}, limit, offset int) ([]string, []map[string]interface{}, error) {
	pagedQuery := fmt.Sprintf("%s LIMIT %d OFFSET %d", query, limit, offset)
	rows, err := s.db.Query(pagedQuery, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return nil, nil, err
		}
		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			rowMap[colName] = *val
		}
		results = append(results, rowMap)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return cols, results, nil
}

// DeleteRow deletes from a table by rowid
func (s *SQLiteService) DeleteRow(table string, rowid int64) error {
	if strings.TrimSpace(table) == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE rowid = ?", table)
	res, err := s.db.Exec(query, rowid)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no row deleted; rowid may not exist")
	}
	return nil
}

// GetTables returns all table names
func (s *SQLiteService) GetTables() ([]string, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

// DropTable drops a table by name
func (s *SQLiteService) DropTable(table string) error {
	if strings.TrimSpace(table) == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", table, err)
	}
	return nil
}
