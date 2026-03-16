package tbl

import (
	"testing"
)

// ── ParseFilter ──────────────────────────────────────────────

func TestParseFilter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantCol string
		wantOp  string
		wantVal string
		wantErr bool
	}{
		{"size greater", "size >100MB", "size", ">", "100MB", false},
		{"size less", "size <1GB", "size", "<", "1GB", false},
		{"size equals", "size =500KB", "size", "=", "500KB", false},
		{"name tilde", "name ~foo", "name", "~", "foo", false},
		{"name glob", "name *bar", "name", "*", "bar", false},
		{"mixed case column", "Size >10MB", "size", ">", "10MB", false},
		{"extra spaces", "  size   >   100MB  ", "size", ">", "100MB  ", false},
		{"empty string", "", "", "", "", true},
		{"no operator", "size100MB", "", "", "", true},
		{"only operator", ">", "", "", "", true},
		{"missing value", "size >", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseFilter(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expr.Column != tt.wantCol {
				t.Errorf("Column = %q, want %q", expr.Column, tt.wantCol)
			}
			if expr.Operator != tt.wantOp {
				t.Errorf("Operator = %q, want %q", expr.Operator, tt.wantOp)
			}
			if expr.Value != tt.wantVal {
				t.Errorf("Value = %q, want %q", expr.Value, tt.wantVal)
			}
		})
	}
}

// ── MatchFilter ──────────────────────────────────────────────

func TestMatchFilter_Size(t *testing.T) {
	tests := []struct {
		name string
		cell string // raw bytes as string
		op   string
		val  string
		want bool
	}{
		{"greater true", "2000000", ">", "1MB", true},
		{"greater false", "500000", ">", "1MB", false},
		{"less true", "500000", "<", "1MB", true},
		{"less false", "2000000", "<", "1MB", false},
		{"equal", "1048576", "=", "1048576", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchFilter(tt.cell, tt.op, tt.val, "size")
			if got != tt.want {
				t.Errorf("MatchFilter(%q, %q, %q, size) = %v, want %v",
					tt.cell, tt.op, tt.val, got, tt.want)
			}
		})
	}
}

func TestMatchFilter_Name(t *testing.T) {
	tests := []struct {
		name string
		cell string
		op   string
		val  string
		want bool
	}{
		{"exact equal", "README.md", "=", "readme.md", true},
		{"exact not equal", "README.md", "=", "other.md", false},
		{"contains tilde", "myfile.txt", "~", "file", true},
		{"contains tilde no match", "myfile.txt", "~", "xyz", false},
		{"glob star", "myfile.txt", "~", "my*", true},
		{"glob star no match", "myfile.txt", "~", "other*", false},
		{"glob middle", "myfile.txt", "~", "*file*", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchFilter(tt.cell, tt.op, tt.val, "name")
			if got != tt.want {
				t.Errorf("MatchFilter(%q, %q, %q, name) = %v, want %v",
					tt.cell, tt.op, tt.val, got, tt.want)
			}
		})
	}
}

func TestMatchFilter_Date(t *testing.T) {
	tests := []struct {
		name   string
		cell   string
		op     string
		val    string
		column string
		want   bool
	}{
		{"created after", "2024-06-15 10:00:00", ">", "2024-01-01", "created", true},
		{"created before", "2024-06-15 10:00:00", "<", "2025-01-01", "created", true},
		{"modified after", "2025-03-01 12:00:00", ">", "2025-01-01", "modified", true},
		{"modified not after", "2024-06-15 10:00:00", ">", "2025-01-01", "modified", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchFilter(tt.cell, tt.op, tt.val, tt.column)
			if got != tt.want {
				t.Errorf("MatchFilter(%q, %q, %q, %q) = %v, want %v",
					tt.cell, tt.op, tt.val, tt.column, got, tt.want)
			}
		})
	}
}

func TestMatchFilter_UnknownOp(t *testing.T) {
	// Unknown operator returns false
	if MatchFilter("anything", "!", "val", "name") {
		t.Error("expected false for unknown operator")
	}
}

// ── ApplyFilter ──────────────────────────────────────────────

func TestApplyFilter(t *testing.T) {
	rows := [][]string{
		{"small.txt", "100", "2024-01-01 00:00:00", "2024-01-01 00:00:00", "user", "644"},
		{"big.txt", "2000000", "2024-06-01 00:00:00", "2024-06-01 00:00:00", "user", "644"},
		{"huge.txt", "5000000000", "2025-01-01 00:00:00", "2025-01-01 00:00:00", "user", "644"},
	}

	t.Run("filter by size", func(t *testing.T) {
		expr := &FilterExpr{Column: "size", Operator: ">", Value: "1MB"}
		got := ApplyFilter(rows, expr)
		if len(got) != 2 {
			t.Fatalf("expected 2 results, got %d", len(got))
		}
		if got[0][0] != "big.txt" {
			t.Errorf("first result = %q, want big.txt", got[0][0])
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		expr := &FilterExpr{Column: "name", Operator: "~", Value: "big"}
		got := ApplyFilter(rows, expr)
		if len(got) != 1 {
			t.Fatalf("expected 1 result, got %d", len(got))
		}
	})

	t.Run("nil filter returns all", func(t *testing.T) {
		got := ApplyFilter(rows, nil)
		if len(got) != 3 {
			t.Fatalf("expected 3 results, got %d", len(got))
		}
	})

	t.Run("unknown column returns all", func(t *testing.T) {
		expr := &FilterExpr{Column: "bogus", Operator: "=", Value: "x"}
		got := ApplyFilter(rows, expr)
		if len(got) != 3 {
			t.Fatalf("expected 3 results, got %d", len(got))
		}
	})
}

// ── SortResults ──────────────────────────────────────────────

func TestSortResults(t *testing.T) {
	makeRows := func() [][]string {
		return [][]string{
			{"charlie.txt", "300", "2024-03-01 00:00:00", "2024-03-01 00:00:00", "u", "644"},
			{"alpha.txt", "100", "2024-01-01 00:00:00", "2024-01-01 00:00:00", "u", "644"},
			{"bravo.txt", "200", "2024-02-01 00:00:00", "2024-02-01 00:00:00", "u", "644"},
		}
	}

	t.Run("sort by name asc", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "name", "asc")
		if rows[0][0] != "alpha.txt" || rows[1][0] != "bravo.txt" || rows[2][0] != "charlie.txt" {
			t.Errorf("unexpected order: %v %v %v", rows[0][0], rows[1][0], rows[2][0])
		}
	})

	t.Run("sort by name desc", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "name", "desc")
		if rows[0][0] != "charlie.txt" {
			t.Errorf("expected charlie.txt first, got %q", rows[0][0])
		}
	})

	t.Run("sort by size asc", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "size", "asc")
		if rows[0][1] != "100" || rows[2][1] != "300" {
			t.Errorf("unexpected size order: %v %v %v", rows[0][1], rows[1][1], rows[2][1])
		}
	})

	t.Run("sort by size desc", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "size", "desc")
		if rows[0][1] != "300" {
			t.Errorf("expected 300 first, got %q", rows[0][1])
		}
	})

	t.Run("sort by created asc", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "created", "asc")
		if rows[0][0] != "alpha.txt" {
			t.Errorf("expected alpha.txt first (earliest), got %q", rows[0][0])
		}
	})

	t.Run("sort by modified desc", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "modified", "desc")
		if rows[0][0] != "charlie.txt" {
			t.Errorf("expected charlie.txt first (latest), got %q", rows[0][0])
		}
	})

	t.Run("sizeparsed maps to size", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "sizeparsed", "asc")
		if rows[0][1] != "100" {
			t.Errorf("sizeparsed should sort as size, got %q first", rows[0][1])
		}
	})

	t.Run("unknown column defaults to name index", func(t *testing.T) {
		rows := makeRows()
		SortResults(rows, "bogus", "asc")
		// Should default to index 0 (name), so alpha first
		if rows[0][0] != "alpha.txt" {
			t.Errorf("expected alpha.txt first for unknown column, got %q", rows[0][0])
		}
	})
}

// ── ByteCount helpers ────────────────────────────────────────

func TestByteCountSI(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{999, "999 B"},
		{1000, "1.0 kB"},
		{1500000, "1.5 MB"},
		{1000000000, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := ByteCountSI(tt.input)
			if got != tt.want {
				t.Errorf("ByteCountSI(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestByteCountIEC(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1048576, "1.0 MiB"},
		{1073741824, "1.0 GiB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := ByteCountIEC(tt.input)
			if got != tt.want {
				t.Errorf("ByteCountIEC(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── parseTime helper ─────────────────────────────────────────

func TestParseTime(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tm, err := parseTime("2024-06-15 10:30:45")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tm.Year() != 2024 || tm.Month() != 6 || tm.Day() != 15 {
			t.Errorf("parsed time mismatch: %v", tm)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseTime("not-a-date")
		if err == nil {
			t.Error("expected error for invalid date")
		}
	})
}
