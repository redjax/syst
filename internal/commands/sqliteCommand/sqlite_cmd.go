package sqliteCommand

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
	sqliteservice "github.com/redjax/syst/internal/services/sqliteService"
	sqliteui "github.com/redjax/syst/internal/services/sqliteService/ui"
)

var dbPath string
var startTable string

// NewSqliteCmd returns the parent `sqlite` command
func NewSqliteCmd() *cobra.Command {
	sqliteCmd := &cobra.Command{
		Use:   "sqlite",
		Short: "Explore and interact with SQLite databases",
		Long: `Run queries, browse data with an interactive TUI, and manage SQLite databases.

Examples:

  syst sqlite open --db-file mydata.sqlite
  syst sqlite open -f ./cache.db
        `,
	}

	// Attach subcommands
	sqliteCmd.AddCommand(newOpenCmd())

	return sqliteCmd
}

// newOpenCmd creates the `sqlite open` subcommand
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

			model := sqliteui.NewUIModel(svc, startTable) // pass startTable here
			p := tea.NewProgram(model)
			_, err = p.Run()
			return err
		},
	}

	cmd.Flags().StringVarP(&dbPath, "db-file", "f", "", "Path to SQLite database file (.db, .sqlite, .sqlite3)")
	cmd.Flags().StringVarP(&startTable, "table", "t", "", "Start in a specific table (skips launcher)")

	return cmd
}
