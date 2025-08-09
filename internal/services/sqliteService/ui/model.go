package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	sqliteservice "github.com/redjax/syst/internal/services/sqliteService"
)

type viewMode int

const (
	modeLauncher viewMode = iota
	modeTable
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
	selectedIndex int
	dCount        int
	inQueryInput  bool
	queryInput    textinput.Model
	errMsg        string
	loading       bool
}

// NewUIModel creates a new UI model
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
