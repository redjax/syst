package sqliteservice

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteService struct {
	db *sql.DB
}

// validTableName checks if a table name is safe to use in SQL queries
// It only allows alphanumeric characters and underscores
func validTableName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, name)
	return matched
}

// ValidTableName is the exported version of validTableName for use by other packages.
func ValidTableName(name string) bool {
	return validTableName(name)
}

// ValidColumnName checks if a column name is safe to use in SQL queries.
func ValidColumnName(name string) bool {
	return validTableName(name) // same rules as table names
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

// scanRowAsStrings scans a single row, reading all values as raw bytes.
// This bypasses the driver's automatic type conversion (e.g. timestamp parsing)
// and preserves the original stored representation.
func scanRowAsStrings(cols []string, rows *sql.Rows) (map[string]interface{}, error) {
	rawValues := make([]sql.RawBytes, len(cols))
	scanArgs := make([]interface{}, len(cols))
	for i := range rawValues {
		scanArgs[i] = &rawValues[i]
	}
	if err := rows.Scan(scanArgs...); err != nil {
		return nil, err
	}
	rowMap := make(map[string]interface{})
	for i, colName := range cols {
		if rawValues[i] == nil {
			rowMap[colName] = nil
		} else {
			s := string(rawValues[i])
			// Preserve numeric types for internal use (e.g. rowid deletion)
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				rowMap[colName] = n
			} else if f, err := strconv.ParseFloat(s, 64); err == nil {
				rowMap[colName] = f
			} else {
				rowMap[colName] = s
			}
		}
	}
	return rowMap, nil
}

// isDatetimeType returns true if the column type declaration is a datetime-like type
// that the go-sqlite3 driver would attempt to parse as time.Time.
func isDatetimeType(colType string) bool {
	t := strings.ToLower(strings.TrimSpace(colType))
	return t == "date" || t == "datetime" || t == "timestamp"
}

// BuildTableQuery constructs a SELECT query for the given table that wraps
// datetime-like columns in CAST(... AS TEXT) so the go-sqlite3 driver returns
// the raw stored value instead of attempting (and potentially failing) to parse
// it as a Go time.Time.
func (s *SQLiteService) BuildTableQuery(table string) (string, error) {
	if !validTableName(table) {
		return "", fmt.Errorf("invalid table name: %s", table)
	}

	// #nosec G201 - table name is validated above
	pragmaRows, err := s.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return "", fmt.Errorf("failed to get table info: %w", err)
	}
	defer pragmaRows.Close()

	var cols []string
	for pragmaRows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue interface{}
		if err := pragmaRows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return "", err
		}
		if isDatetimeType(colType) {
			cols = append(cols, fmt.Sprintf("CAST(\"%s\" AS TEXT) AS \"%s\"", name, name))
		} else {
			cols = append(cols, fmt.Sprintf("\"%s\"", name))
		}
	}
	if err := pragmaRows.Err(); err != nil {
		return "", err
	}

	if len(cols) == 0 {
		return fmt.Sprintf("SELECT rowid, * FROM %s", table), nil
	}

	return fmt.Sprintf("SELECT rowid, %s FROM %s", strings.Join(cols, ", "), table), nil
}

// QueryWithPagination runs a query with LIMIT/OFFSET and returns columns, rows
func (s *SQLiteService) QueryWithPagination(query string, args []interface{}, limit, offset int) ([]string, []map[string]interface{}, error) {
	// Strip any existing LIMIT/OFFSET to avoid double pagination
	cleaned := regexp.MustCompile(`(?i)\s+LIMIT\s+\d+(\s+OFFSET\s+\d+)?\s*$`).ReplaceAllString(query, "")
	pagedQuery := fmt.Sprintf("%s LIMIT %d OFFSET %d", cleaned, limit, offset)
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
		rowMap, err := scanRowAsStrings(cols, rows)
		if err != nil {
			return nil, nil, err
		}
		results = append(results, rowMap)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return cols, results, nil
}

// Query runs a query without pagination and returns columns, rows
func (s *SQLiteService) Query(query string, args []interface{}) ([]string, []map[string]interface{}, error) {
	rows, err := s.db.Query(query, args...)
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
		rowMap, err := scanRowAsStrings(cols, rows)
		if err != nil {
			return nil, nil, err
		}
		results = append(results, rowMap)
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return cols, results, nil
}

// DeleteRow deletes from a table by rowid (or id if rowid not available)
func (s *SQLiteService) DeleteRow(table string, rowid int64) error {
	if strings.TrimSpace(table) == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Validate table name to prevent SQL injection
	if !validTableName(table) {
		return fmt.Errorf("invalid table name: must contain only alphanumeric characters and underscores")
	}

	// Try rowid first
	// #nosec G201 - table name is validated above to contain only alphanumeric chars and underscores
	query := fmt.Sprintf("DELETE FROM %s WHERE rowid = ?", table)
	res, err := s.db.Exec(query, rowid)
	if err == nil {
		affected, err := res.RowsAffected()
		if err == nil && affected > 0 {
			return nil
		}
	}

	// If rowid failed, try id column
	// #nosec G201 - table name is validated above to contain only alphanumeric chars and underscores
	query = fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	res, err = s.db.Exec(query, rowid)
	if err != nil {
		return fmt.Errorf("delete failed with both rowid and id: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no row deleted; id %d may not exist", rowid)
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
	if !validTableName(table) {
		return fmt.Errorf("invalid table name: must contain only alphanumeric characters and underscores")
	}
	// #nosec G201 - table name is validated above
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", table)
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", table, err)
	}
	return nil
}

// GetTableRowCount returns the total number of rows in a table.
func (s *SQLiteService) GetTableRowCount(table string) (int, error) {
	if !validTableName(table) {
		return 0, fmt.Errorf("invalid table name: %s", table)
	}
	// #nosec G201 - table name is validated above
	var count int
	err := s.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Exec executes a statement without returning rows
func (s *SQLiteService) Exec(query string, args ...interface{}) error {
	_, err := s.db.Exec(query, args...)
	return err
}
