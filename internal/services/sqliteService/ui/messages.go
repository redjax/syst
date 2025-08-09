package ui

type queryResultMsg struct {
	columns []string
	rows    []map[string]interface{}
}

type tablesLoadedMsg []string
