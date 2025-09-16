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

func NewSqliteCmd() *cobra.Command {
	sqliteCmd := &cobra.Command{
		Use:   "sqlite",
		Short: "Explore and interact with SQLite databases",
		Long:  "Run queries, browse data with an interactive TUI, and manage SQLite databases.",
	}
	sqliteCmd.AddCommand(newOpenCmd())
	sqliteCmd.AddCommand(newImportCmd())
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
