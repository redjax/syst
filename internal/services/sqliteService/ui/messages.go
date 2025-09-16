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

type schemaLoadedMsg struct {
	tableName string
	schema    []map[string]interface{}
	err       error
}

type tableInfoLoadedMsg struct {
	tableName string
	info      []map[string]interface{}
	err       error
}

type indexesLoadedMsg struct {
	indexes []map[string]interface{}
	err     error
}

type viewsLoadedMsg struct {
	views []map[string]interface{}
	err   error
}
