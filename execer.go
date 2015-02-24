package dat

// Result is Execer result.
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// Executable is an object that can be queried.
type Execer interface {
	Exec() (*Result, error)
	//Query() (*sql.Rows, error)
	QueryScalar(destinations ...interface{}) error
	QuerySlice(dest interface{}) error
	QueryStruct(dest interface{}) error
	QueryStructs(dest interface{}) error
}
