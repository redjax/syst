package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	t "github.com/evertras/bubble-table/table"
	sqliteservice "github.com/redjax/syst/internal/services/sqliteService"
)

type viewMode int

const (
	modeLauncher viewMode = iota
	modeTable
	modeExpandCell
	modeSchema
	modeTableInfo
	modeIndexes
	modeViews
	modeImport
)

type UIModel struct {
	// service
	svc *sqliteservice.SQLiteService

	// mode / navigation
	mode       viewMode
	tableName  string
	tables     []string
	tableIndex int

	// query & pagination
	query         string
	limit, offset int

	// DB results (clean DB-only column names)
	columns []string
	rows    []map[string]interface{}

	// selection / cursor
	selectedIndex int
	selectedCol   int
	selectedRows  map[int]bool

	// quick state
	dCount  int
	errMsg  string
	loading bool

	// query input (focus on '/')
	queryInput textinput.Model

	// expand (scrollable)
	expandRow int
	expandCol string
	expandVal string
	vp        viewport.Model

	// schema and metadata
	schemaInfo    []map[string]interface{}
	tableInfoData []map[string]interface{}
	indexesData   []map[string]interface{}
	viewsData     []map[string]interface{}

	// import wizard
	importFilePath  string
	importStep      int // 0=file selection, 1=table selection, 2=column mapping, 3=import
	importFileInput textinput.Model

	// table component and terminal size
	tableComp  t.Model
	termWidth  int
	termHeight int
}

func NewUIModel(svc *sqliteservice.SQLiteService, startTable string) UIModel {
	ti := textinput.New()
	ti.Placeholder = "Enter SQL query"
	ti.CharLimit = 1024
	ti.Width = 50
	ti.Blur()

	// Import file input
	importInput := textinput.New()
	importInput.Placeholder = "Enter CSV file path (e.g., ./data.csv)"
	importInput.CharLimit = 256
	importInput.Width = 60
	importInput.Blur()

	// minimal viewport until we get window size
	vp := viewport.Model{}
	m := UIModel{
		svc:             svc,
		limit:           20,
		offset:          0,
		queryInput:      ti,
		importFileInput: importInput,
		selectedRows:    make(map[int]bool),
		selectedIndex:   0,
		selectedCol:     0,
		vp:              vp,
	}

	if startTable != "" {
		m.mode = modeTable
		m.tableName = startTable
		m.query = fmt.Sprintf("SELECT rowid, * FROM %s", startTable)
	} else {
		m.mode = modeLauncher
	}
	return m
}

func (m UIModel) Init() tea.Cmd {
	// start by loading tables or running query depending on mode
	if m.mode == modeLauncher {
		return m.loadTablesCmd()
	}
	return m.runQueryCmd()
}
