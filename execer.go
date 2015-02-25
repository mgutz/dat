package dat

// Result is Execer result.
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// Execer is an object that can be execute/query a database.
type Execer interface {
	Exec() (*Result, error)
	QueryScalar(destinations ...interface{}) error
	QuerySlice(dest interface{}) error
	QueryStruct(dest interface{}) error
	QueryStructs(dest interface{}) error
}
