package sqliteCommand

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	sqliteservice "github.com/redjax/syst/internal/services/sqliteService"
	sqliteui "github.com/redjax/syst/internal/services/sqliteService/ui"
	"github.com/spf13/cobra"
)

var dbPath string
var startTable string
var importDbPath string
var importCsvPath string
var importTableName string
var exportDbPath string
var exportTableName string
var exportOutputDir string

func NewSqliteCmd() *cobra.Command {
	sqliteCmd := &cobra.Command{
		Use:   "sqlite",
		Short: "Explore and interact with SQLite databases",
		Long:  "Run queries, browse data with an interactive TUI, and manage SQLite databases.",
	}
	sqliteCmd.AddCommand(newOpenCmd())
	sqliteCmd.AddCommand(newImportCmd())
	sqliteCmd.AddCommand(newExportCmd())
	return sqliteCmd
}

func newOpenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open",
		Short: "Open a SQLite database in the interactive TUI explorer",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				return fmt.Errorf("you must specify a database file with --db")
			}
			absPath, err := filepath.Abs(dbPath)
			if err != nil {
				return err
			}
			svc, err := sqliteservice.NewSQLiteService(absPath)
			if err != nil {
				return err
			}
			defer svc.Close()
			model := sqliteui.NewUIModel(svc, startTable)
			p := tea.NewProgram(model)
			_, err = p.Run()
			return err
		},
	}
	cmd.Flags().StringVarP(&dbPath, "db-file", "f", "", "Path to SQLite database file")
	cmd.Flags().StringVarP(&startTable, "table", "t", "", "Start in a specific table")
	return cmd
}

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import CSV data into a SQLite database",
		RunE: func(cmd *cobra.Command, args []string) error {
			if importDbPath == "" || importCsvPath == "" || importTableName == "" {
				return fmt.Errorf("you must specify --db, --csv, and --table")
			}
			absDbPath, _ := filepath.Abs(importDbPath)
			absCsvPath, _ := filepath.Abs(importCsvPath)
			return performCSVImport(absDbPath, absCsvPath, importTableName)
		},
	}
	cmd.Flags().StringVarP(&importDbPath, "db", "d", "", "Path to SQLite database file")
	cmd.Flags().StringVarP(&importCsvPath, "csv", "c", "", "Path to CSV file to import")
	cmd.Flags().StringVarP(&importTableName, "table", "t", "", "Name of table to import data into")
	cmd.MarkFlagRequired("db")
	cmd.MarkFlagRequired("csv")
	cmd.MarkFlagRequired("table")
	return cmd
}

func newExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export SQLite table data to CSV",
		Long:  "Export data from SQLite database tables to CSV files. Export a specific table or all tables.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if exportDbPath == "" {
				return fmt.Errorf("you must specify a database file with --db")
			}

			absDbPath, err := filepath.Abs(exportDbPath)
			if err != nil {
				return fmt.Errorf("failed to resolve database path: %w", err)
			}

			absOutputDir := ""
			if exportOutputDir != "" {
				absOutputDir, err = filepath.Abs(exportOutputDir)
				if err != nil {
					return fmt.Errorf("failed to resolve output directory: %w", err)
				}
			} else {
				absOutputDir = filepath.Dir(absDbPath) // Use DB directory if no output dir specified
			}

			return performCSVExport(absDbPath, exportTableName, absOutputDir)
		},
	}

	cmd.Flags().StringVarP(&exportDbPath, "db", "d", "", "Path to SQLite database file")
	cmd.Flags().StringVarP(&exportTableName, "table", "t", "", "Table name to export (if empty, exports all tables)")
	cmd.Flags().StringVarP(&exportOutputDir, "output", "o", "", "Output directory for CSV files (defaults to database directory)")
	cmd.MarkFlagRequired("db")

	return cmd
}

func performCSVExport(dbPath, tableName, outputDir string) error {
	svc, err := sqliteservice.NewSQLiteService(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer svc.Close()

	var tablesToExport []string

	if tableName != "" {
		// Export specific table
		tablesToExport = []string{tableName}
	} else {
		// Export all tables
		tables, err := svc.GetTables()
		if err != nil {
			return fmt.Errorf("failed to get table list: %w", err)
		}
		if len(tables) == 0 {
			return fmt.Errorf("no tables found in database")
		}
		tablesToExport = tables
		fmt.Printf("Found %d tables to export: %s\n", len(tables), strings.Join(tables, ", "))
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	totalRows := 0
	for _, table := range tablesToExport {
		rows, err := exportTable(svc, table, outputDir)
		if err != nil {
			return fmt.Errorf("failed to export table %s: %w", table, err)
		}
		totalRows += rows
	}

	fmt.Printf("✅ Successfully exported %d total rows from %d table(s) to %s\n",
		totalRows, len(tablesToExport), outputDir)

	return nil
}

func exportTable(svc *sqliteservice.SQLiteService, tableName, outputDir string) (int, error) {
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	columns, rows, err := svc.Query(query, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to query table %s: %w", tableName, err)
	}

	if len(rows) == 0 {
		fmt.Printf("⚠️  Table %s is empty, skipping\n", tableName)
		return 0, nil
	}

	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s.csv", tableName))
	file, err := os.Create(outputFile)
	if err != nil {
		return 0, fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(columns); err != nil {
		return 0, fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, row := range rows {
		record := make([]string, len(columns))
		for i, col := range columns {
			if val := row[col]; val != nil {
				record[i] = fmt.Sprintf("%v", val)
			} else {
				record[i] = ""
			}
		}
		if err := writer.Write(record); err != nil {
			return 0, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	fmt.Printf("📄 Exported %d rows from table %s to %s\n", len(rows), tableName, outputFile)
	return len(rows), nil
}

func performCSVImport(dbPath, csvPath, tableName string) error {
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV file: %w", err)
	}

	if len(records) < 1 {
		return fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	dataRows := records[1:]

	fmt.Printf("Importing %d rows into table %s...\n", len(dataRows), tableName)

	svc, err := sqliteservice.NewSQLiteService(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open/create database: %w", err)
	}
	defer svc.Close()

	var columnDefs []string
	for _, header := range headers {
		cleanHeader := strings.ReplaceAll(header, " ", "_")
		columnDefs = append(columnDefs, fmt.Sprintf("%s TEXT", cleanHeader))
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(columnDefs, ", "))
	if err := svc.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	placeholders := make([]string, len(headers))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s VALUES (%s)", tableName, strings.Join(placeholders, ", "))

	for i, row := range dataRows {
		values := make([]interface{}, len(row))
		for j, val := range row {
			values[j] = val
		}
		if err := svc.Exec(insertSQL, values...); err != nil {
			return fmt.Errorf("failed to insert row %d: %w", i+1, err)
		}
	}

	fmt.Printf("✅ Successfully imported %d rows into table %s\n", len(dataRows), tableName)
	fmt.Printf("You can explore with: syst sqlite open --db %s --table %s\n", dbPath, tableName)
	return nil
}
