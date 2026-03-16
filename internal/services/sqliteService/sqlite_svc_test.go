package sqliteservice

import (
	"testing"
)

func TestValidTableName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple", "users", true},
		{"with underscore", "user_table", true},
		{"with numbers", "table123", true},
		{"mixed", "My_Table_99", true},
		{"with spaces", "my table", false},
		{"with dash", "my-table", false},
		{"with dot", "my.table", false},
		{"SQL injection attempt", "users; DROP TABLE users", false},
		{"empty", "", false},
		{"with quotes", `"users"`, false},
		{"with parens", "users()", false},
		{"with semicolon", "users;", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidTableName(tt.input)
			if got != tt.want {
				t.Errorf("ValidTableName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidColumnName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple", "name", true},
		{"with underscore", "first_name", true},
		{"with numbers", "col1", true},
		{"with space", "col name", false},
		{"SQL injection", "1; DROP TABLE users--", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidColumnName(tt.input)
			if got != tt.want {
				t.Errorf("ValidColumnName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsDatetimeType(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"date", true},
		{"datetime", true},
		{"timestamp", true},
		{"DATE", true},
		{"DATETIME", true},
		{"TIMESTAMP", true},
		{"  datetime  ", true},
		{"text", false},
		{"integer", false},
		{"real", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isDatetimeType(tt.input)
		if got != tt.want {
			t.Errorf("isDatetimeType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func newTestDB(t *testing.T) *SQLiteService {
	t.Helper()
	svc, err := NewSQLiteService(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteService(:memory:) error: %v", err)
	}
	t.Cleanup(func() { svc.Close() })
	return svc
}

func TestNewSQLiteService(t *testing.T) {
	svc := newTestDB(t)
	if svc == nil {
		t.Fatal("NewSQLiteService returned nil")
	}
}

func TestNewSQLiteService_InvalidPath(t *testing.T) {
	_, err := NewSQLiteService("/nonexistent/path/to/db.sqlite")
	// sql.Open doesn't fail on invalid path, only on first query
	// So this tests that initialization at least doesn't panic
	if err != nil {
		// Some systems may error here, that's also acceptable
		t.Logf("NewSQLiteService with invalid path returned error (acceptable): %v", err)
	}
}

func TestSQLiteService_GetTables_Empty(t *testing.T) {
	svc := newTestDB(t)
	tables, err := svc.GetTables()
	if err != nil {
		t.Fatalf("GetTables error: %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("GetTables on empty DB = %v, want empty", tables)
	}
}

func TestSQLiteService_CreateAndQuery(t *testing.T) {
	svc := newTestDB(t)

	// Create a table
	err := svc.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}

	// Insert data
	err = svc.Exec("INSERT INTO test (name, age) VALUES (?, ?)", "Alice", 30)
	if err != nil {
		t.Fatalf("INSERT error: %v", err)
	}
	err = svc.Exec("INSERT INTO test (name, age) VALUES (?, ?)", "Bob", 25)
	if err != nil {
		t.Fatalf("INSERT error: %v", err)
	}

	// Verify tables
	tables, err := svc.GetTables()
	if err != nil {
		t.Fatalf("GetTables error: %v", err)
	}
	if len(tables) != 1 || tables[0] != "test" {
		t.Errorf("GetTables = %v, want [test]", tables)
	}

	// Query
	cols, rows, err := svc.Query("SELECT * FROM test", nil)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if len(cols) != 3 {
		t.Errorf("Query cols = %d, want 3", len(cols))
	}
	if len(rows) != 2 {
		t.Errorf("Query rows = %d, want 2", len(rows))
	}
}

func TestSQLiteService_QueryWithPagination(t *testing.T) {
	svc := newTestDB(t)

	err := svc.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}
	for i := 0; i < 25; i++ {
		err = svc.Exec("INSERT INTO items (val) VALUES (?)", "item")
		if err != nil {
			t.Fatalf("INSERT error: %v", err)
		}
	}

	// First page
	_, rows, err := svc.QueryWithPagination("SELECT * FROM items", nil, 10, 0)
	if err != nil {
		t.Fatalf("QueryWithPagination error: %v", err)
	}
	if len(rows) != 10 {
		t.Errorf("page 1 rows = %d, want 10", len(rows))
	}

	// Second page
	_, rows, err = svc.QueryWithPagination("SELECT * FROM items", nil, 10, 10)
	if err != nil {
		t.Fatalf("QueryWithPagination error: %v", err)
	}
	if len(rows) != 10 {
		t.Errorf("page 2 rows = %d, want 10", len(rows))
	}

	// Third page (partial)
	_, rows, err = svc.QueryWithPagination("SELECT * FROM items", nil, 10, 20)
	if err != nil {
		t.Fatalf("QueryWithPagination error: %v", err)
	}
	if len(rows) != 5 {
		t.Errorf("page 3 rows = %d, want 5", len(rows))
	}
}

func TestSQLiteService_QueryWithPagination_StripsExistingLimit(t *testing.T) {
	svc := newTestDB(t)

	err := svc.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}
	for i := 0; i < 10; i++ {
		err = svc.Exec("INSERT INTO items (val) VALUES (?)", "item")
		if err != nil {
			t.Fatalf("INSERT error: %v", err)
		}
	}

	// Query with existing LIMIT should be stripped and replaced
	_, rows, err := svc.QueryWithPagination("SELECT * FROM items LIMIT 3", nil, 5, 0)
	if err != nil {
		t.Fatalf("QueryWithPagination error: %v", err)
	}
	if len(rows) != 5 {
		t.Errorf("rows = %d, want 5 (existing LIMIT should be stripped)", len(rows))
	}
}

func TestSQLiteService_BuildTableQuery(t *testing.T) {
	svc := newTestDB(t)

	err := svc.Exec("CREATE TABLE logs (id INTEGER PRIMARY KEY, msg TEXT, created_at DATETIME)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}

	query, err := svc.BuildTableQuery("logs")
	if err != nil {
		t.Fatalf("BuildTableQuery error: %v", err)
	}

	// Should contain CAST for datetime column
	if query == "" {
		t.Error("BuildTableQuery returned empty string")
	}
	// The query should include rowid
	if !contains(query, "rowid") {
		t.Errorf("BuildTableQuery missing rowid: %q", query)
	}
	// The query should cast datetime columns
	if !contains(query, "CAST") {
		t.Errorf("BuildTableQuery should CAST datetime columns: %q", query)
	}
}

func TestSQLiteService_BuildTableQuery_InvalidTable(t *testing.T) {
	svc := newTestDB(t)
	_, err := svc.BuildTableQuery("invalid; DROP TABLE users")
	if err == nil {
		t.Error("BuildTableQuery with invalid table name should return error")
	}
}

func TestSQLiteService_DeleteRow(t *testing.T) {
	svc := newTestDB(t)

	err := svc.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}
	err = svc.Exec("INSERT INTO items (val) VALUES (?)", "to_delete")
	if err != nil {
		t.Fatalf("INSERT error: %v", err)
	}

	err = svc.DeleteRow("items", 1)
	if err != nil {
		t.Fatalf("DeleteRow error: %v", err)
	}

	// Verify it was deleted
	_, rows, err := svc.Query("SELECT * FROM items", nil)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("after delete, rows = %d, want 0", len(rows))
	}
}

func TestSQLiteService_DeleteRow_InvalidTable(t *testing.T) {
	svc := newTestDB(t)
	err := svc.DeleteRow("invalid; DROP TABLE foo", 1)
	if err == nil {
		t.Error("DeleteRow with invalid table name should return error")
	}
}

func TestSQLiteService_DeleteRow_EmptyTable(t *testing.T) {
	svc := newTestDB(t)
	err := svc.DeleteRow("", 1)
	if err == nil {
		t.Error("DeleteRow with empty table name should return error")
	}
}

func TestSQLiteService_DeleteRow_NonexistentRow(t *testing.T) {
	svc := newTestDB(t)
	err := svc.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}

	err = svc.DeleteRow("items", 999)
	if err == nil {
		t.Error("DeleteRow on nonexistent row should return error")
	}
}

func TestSQLiteService_DropTable(t *testing.T) {
	svc := newTestDB(t)

	err := svc.Exec("CREATE TABLE dropme (id INTEGER PRIMARY KEY)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}

	err = svc.DropTable("dropme")
	if err != nil {
		t.Fatalf("DropTable error: %v", err)
	}

	tables, _ := svc.GetTables()
	for _, tbl := range tables {
		if tbl == "dropme" {
			t.Error("table 'dropme' still exists after DropTable")
		}
	}
}

func TestSQLiteService_DropTable_InvalidName(t *testing.T) {
	svc := newTestDB(t)
	err := svc.DropTable("bad; DROP TABLE other")
	if err == nil {
		t.Error("DropTable with invalid name should return error")
	}
}

func TestSQLiteService_DropTable_Empty(t *testing.T) {
	svc := newTestDB(t)
	err := svc.DropTable("")
	if err == nil {
		t.Error("DropTable with empty name should return error")
	}
}

func TestSQLiteService_GetTableRowCount(t *testing.T) {
	svc := newTestDB(t)

	err := svc.Exec("CREATE TABLE items (id INTEGER PRIMARY KEY, val TEXT)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}
	for i := 0; i < 15; i++ {
		svc.Exec("INSERT INTO items (val) VALUES (?)", "x")
	}

	count, err := svc.GetTableRowCount("items")
	if err != nil {
		t.Fatalf("GetTableRowCount error: %v", err)
	}
	if count != 15 {
		t.Errorf("GetTableRowCount = %d, want 15", count)
	}
}

func TestSQLiteService_GetTableRowCount_InvalidName(t *testing.T) {
	svc := newTestDB(t)
	_, err := svc.GetTableRowCount("bad; DROP TABLE x")
	if err == nil {
		t.Error("GetTableRowCount with invalid table should return error")
	}
}

func TestSQLiteService_DatetimeColumn(t *testing.T) {
	svc := newTestDB(t)

	err := svc.Exec("CREATE TABLE events (id INTEGER PRIMARY KEY, name TEXT, created_at DATETIME)")
	if err != nil {
		t.Fatalf("CREATE TABLE error: %v", err)
	}
	err = svc.Exec("INSERT INTO events (name, created_at) VALUES (?, ?)", "test", "2025-01-15 10:30:00")
	if err != nil {
		t.Fatalf("INSERT error: %v", err)
	}

	// Use BuildTableQuery which wraps datetime in CAST
	query, err := svc.BuildTableQuery("events")
	if err != nil {
		t.Fatalf("BuildTableQuery error: %v", err)
	}

	_, rows, err := svc.Query(query, nil)
	if err != nil {
		t.Fatalf("Query error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	// The datetime value should come back as a string, not zero time
	val := rows[0]["created_at"]
	if val == nil {
		t.Fatal("created_at is nil")
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("created_at type = %T, want string", val)
	}
	if str != "2025-01-15 10:30:00" {
		t.Errorf("created_at = %q, want %q", str, "2025-01-15 10:30:00")
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
