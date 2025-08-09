package ui

type queryResultMsg struct {
	columns []string
	rows    []map[string]interface{}
}

type tablesLoadedMsg []string

type deleteDoneMsg struct {
	err error
}

type dropDoneMsg struct {
	err error
}
