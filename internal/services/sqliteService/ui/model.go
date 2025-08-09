package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	t "github.com/evertras/bubble-table/table"
	sqliteservice "github.com/redjax/syst/internal/services/sqliteService"
)

type viewMode int

const (
	modeLauncher viewMode = iota
	modeTable
	modeExpandCell
)

type UIModel struct {
	svc           *sqliteservice.SQLiteService
	mode          viewMode
	tableName     string
	tables        []string
	tableIndex    int
	query         string
	limit, offset int
	columns       []string
	rows          []map[string]interface{}

	// Selection tracking
	selectedIndex int          // current selected row index
	selectedCol   int          // current selected column index
	selectedRows  map[int]bool // multi-select checkboxes

	dCount     int
	queryInput textinput.Model
	errMsg     string
	loading    bool

	// Expand cell view
	expandRow int
	expandCol string
	expandVal string

	tableComp  t.Model
	termWidth  int
	termHeight int
}

func NewUIModel(svc *sqliteservice.SQLiteService, startTable string) UIModel {
	ti := textinput.New()
	ti.Placeholder = "Enter SQL query"
	ti.CharLimit = 256
	ti.Width = 50

	m := UIModel{
		svc:           svc,
		limit:         20,
		offset:        0,
		queryInput:    ti,
		selectedIndex: 0,
		selectedCol:   0,
		selectedRows:  make(map[int]bool),
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
	if m.mode == modeLauncher {
		return m.loadTablesCmd()
	}
	return m.runQueryCmd()
}
